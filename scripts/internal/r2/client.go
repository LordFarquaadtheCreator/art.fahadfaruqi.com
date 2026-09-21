package r2

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "upload-images/internal/config"
	"upload-images/internal/exif"
)

type Client struct {
	s3     *s3.Client
	bucket string
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
		s3:     client,
		bucket: cfg.R2.Bucket,
	}, nil
}

func (c *Client) Upload(filePath string, metadata *exif.Metadata) error {
	missingFields := []string{}
	if metadata.Title == "" {
		missingFields = append(missingFields, "title")
	}
	if metadata.AltText == "" {
		missingFields = append(missingFields, "altText")
	}
	if metadata.Description == "" {
		missingFields = append(missingFields, "description")
	}
	if metadata.Set == "" {
		missingFields = append(missingFields, "set")
	}
	if metadata.Number == 0 {
		missingFields = append(missingFields, "number")
	}
	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingFields, ", "))
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	key := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))

	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".webp":
		contentType = "image/webp"
	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}

	metaMap := map[string]string{
		"title":       metadata.Title,
		"altText":     metadata.AltText,
		"description": metadata.Description,
		"set":         metadata.Set,
		"number":      strconv.Itoa(metadata.Number),
	}

	for k, v := range metadata.Exif {
		metaMap[k] = v
	}

	_, err = c.s3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &c.bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		Metadata:    metaMap,
	})

	return err
}
