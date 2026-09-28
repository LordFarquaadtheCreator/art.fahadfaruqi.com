package r2

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	appconfig "manage-images/internal/config"
	"manage-images/internal/exif"
)

type Client struct {
	s3      *s3.Client
	bucket  string
	cdnBase string
}

// Object is one entry from a bucket listing.
type Object struct {
	Key          string
	Size         int64
	ETag         string
	LastModified time.Time
}

// ObjectInfo is the state of a single object: its generated fields plus the
// custom metadata it currently carries.
type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
	Metadata     map[string]string
}

func New(cfg *appconfig.Config) (*Client, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2.AccessKeyID,
			cfg.R2.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.R2.S3APIEndpoint)
	})

	return &Client{
		s3:      client,
		bucket:  cfg.R2.Bucket,
		cdnBase: strings.TrimRight(cfg.R2.CDNBase, "/"),
	}, nil
}

// CDNBase returns the public base URL the bucket is served from, or "" when
// none is configured.
func (c *Client) CDNBase() string {
	return c.cdnBase
}

// URL returns the public URL for a key, or "" when no cdn_base is configured.
func (c *Client) URL(key string) string {
	if c.cdnBase == "" {
		return ""
	}
	return c.cdnBase + "/" + key
}

// ContentTypeFor maps a key's extension to the content type R2 should serve it with.
func ContentTypeFor(key string) (string, error) {
	switch strings.ToLower(filepath.Ext(key)) {
	case ".jpg", ".jpeg":
		return "image/jpeg", nil
	case ".png":
		return "image/png", nil
	case ".webp":
		return "image/webp", nil
	default:
		return "", fmt.Errorf("unsupported file type: %s", filepath.Ext(key))
	}
}

// List returns every object whose key starts with prefix.
func (c *Client) List(prefix string) ([]Object, error) {
	input := &s3.ListObjectsV2Input{Bucket: &c.bucket}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	var objects []Object
	paginator := s3.NewListObjectsV2Paginator(c.s3, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}
		for _, object := range page.Contents {
			objects = append(objects, Object{
				Key:          aws.ToString(object.Key),
				Size:         aws.ToInt64(object.Size),
				ETag:         aws.ToString(object.ETag),
				LastModified: aws.ToTime(object.LastModified),
			})
		}
	}

	return objects, nil
}

// Head returns the current state of a single object.
func (c *Client) Head(key string) (*ObjectInfo, error) {
	out, err := c.s3.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: &c.bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to head %s: %w", key, err)
	}

	return &ObjectInfo{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		ETag:         aws.ToString(out.ETag),
		ContentType:  aws.ToString(out.ContentType),
		LastModified: aws.ToTime(out.LastModified),
		Metadata:     out.Metadata,
	}, nil
}

// Update rewrites an object's custom metadata, optionally under a new key. The
// copy replaces the metadata wholesale, so callers pass the metadata the object
// should end up with. When newKey differs from key, the old key is removed.
func (c *Client) Update(key, newKey string, metadata map[string]string, contentType string) error {
	if newKey == "" {
		newKey = key
	}

	_, err := c.s3.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:            &c.bucket,
		Key:               aws.String(newKey),
		CopySource:        aws.String(copySource(c.bucket, key)),
		ContentType:       aws.String(contentType),
		Metadata:          metadata,
		MetadataDirective: types.MetadataDirectiveReplace,
	})
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", key, err)
	}

	if newKey == key {
		return nil
	}

	return c.Delete(key)
}

// Delete removes a single object.
func (c *Client) Delete(key string) error {
	_, err := c.s3.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &c.bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete %s: %w", key, err)
	}

	return nil
}

func (c *Client) Upload(filePath string, metadata *exif.Metadata) error {
	if err := metadata.Validate(); err != nil {
		return err
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	key := filepath.Base(filePath)

	contentType, err := ContentTypeFor(filePath)
	if err != nil {
		return err
	}

	_, err = c.s3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &c.bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata:    metadata.Map(),
	})

	return err
}

// copySource builds the x-amz-copy-source value, escaping each path segment so
// keys holding spaces or other reserved characters survive the round trip.
func copySource(bucket, key string) string {
	segments := strings.Split(key, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}

	return bucket + "/" + strings.Join(segments, "/")
}
