package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"manage-images/internal/exif"
	"manage-images/internal/ffmpeg"
	"manage-images/internal/prompt"
	"manage-images/internal/r2"
)

// Cache policy by tier. Neither is immutable, so a re-upload at the same key
// heals within the header's lifetime instead of sticking for a year.
const (
	masterCacheControl     = "public, max-age=86400"
	compressedCacheControl = "public, max-age=604800"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [file...]",
		Aliases: []string{"upload"},
		Short:   "Upload local images to the bucket with metadata",
		Run:     runCreate,
	}

	addTargetFlags(cmd)

	cmd.Flags().String("master-prefix", "art/master", "Key prefix the master is written under")
	cmd.Flags().String("compressed-prefix", "art/compressed", "Key prefix the compressed sibling is written under")
	cmd.Flags().Int("compressed-width", 1600, "Width of the compressed sibling, in pixels; the height follows the master's ratio")
	cmd.Flags().Int("crf", 22, "AVIF crf for the compressed sibling — lower is better and larger")
	cmd.Flags().Int("preset", 6, "SVT-AV1 preset for the compressed sibling — slower presets spend quality, not bytes")

	cmd.Flags().String(
		"inherit",
		"",
		"Upload with the metadata of the existing object of the same name and this extension, without prompting, e.g. --inherit .png",
	)

	return cmd
}

// masterKeyFor is the key a local file is written under. The bucket keeps the
// local file name.
func masterKeyFor(prefix, filePath string) string {
	return prefix + "/" + filepath.Base(filePath)
}

// compressedKeyFor is the key of a master's compressed sibling: same stem, AVIF.
func compressedKeyFor(prefix, masterKey string) string {
	stem := strings.TrimSuffix(filepath.Base(masterKey), filepath.Ext(masterKey))
	return prefix + "/" + stem + ".avif"
}

// localFiles resolves -dir/-pattern or the positional arguments to local paths.
func localFiles(workingDir, pattern string, args []string) ([]string, error) {
	if pattern != "" {
		matches, err := filepath.Glob(filepath.Join(workingDir, pattern))
		if err != nil {
			return nil, fmt.Errorf("failed to match pattern: %w", err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("no files found matching pattern: %s", pattern)
		}

		return matches, nil
	}

	if len(args) < 1 {
		return nil, fmt.Errorf("usage: manage-images create -dir <directory> -pattern <pattern> OR manage-images create -dir <directory> <file1> [file2...]")
	}

	filePaths := make([]string, 0, len(args))
	for _, fileName := range args {
		filePaths = append(filePaths, filepath.Join(workingDir, fileName))
	}

	return filePaths, nil
}

func runCreate(cmd *cobra.Command, args []string) {
	workingDir, _ := cmd.Flags().GetString("dir")
	pattern, _ := cmd.Flags().GetString("pattern")
	inherit, _ := cmd.Flags().GetString("inherit")
	masterPrefix, _ := cmd.Flags().GetString("master-prefix")
	compressedPrefix, _ := cmd.Flags().GetString("compressed-prefix")
	width, _ := cmd.Flags().GetInt("compressed-width")
	crf, _ := cmd.Flags().GetInt("crf")
	preset, _ := cmd.Flags().GetInt("preset")

	filePaths, err := localFiles(workingDir, pattern, args)
	if err != nil {
		log.Fatal(err)
	}

	// The compressed sibling is generated here, so a machine without the encoder
	// fails before anything reaches the bucket: a master with no sibling is a
	// photograph the gallery would have to render at full size.
	if err := ffmpeg.Available(); err != nil {
		log.Fatalf("Cannot generate the compressed sibling: %v", err)
	}

	client := newClient()

	scratch, err := os.MkdirTemp("", "manage-images")
	if err != nil {
		log.Fatalf("Failed to create a working directory: %v", err)
	}
	defer os.RemoveAll(scratch)

	type pending struct {
		path     string
		metadata *exif.Metadata
	}

	defaultSet := ""
	var uploads []pending

	for i, filePath := range filePaths {
		if inherit != "" {
			// A re-encoded file replaces the item it was made from, so it takes that
			// item's metadata instead of being asked for the same answers again. The
			// upload time travels with it, because R2's own timestamp resets to now on
			// every upload and the gallery reads that value as the set's date.
			stem := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
			sibling := masterKeyFor(masterPrefix, stem+inherit)

			info, err := client.Head(sibling)
			if err != nil {
				log.Fatalf("No object to inherit metadata from for %s: %v", filepath.Base(filePath), err)
			}

			metadata, err := exif.FromMap(info.Metadata)
			if err != nil {
				log.Fatalf("Metadata on %s cannot be inherited: %v", sibling, err)
			}

			metadata.Exif["uploaded"] = info.LastModified.UTC().Format("2006-01-02T15:04:05.000Z")

			uploads = append(uploads, pending{filePath, metadata})

			continue
		}

		metadata, err := exif.Extract(filePath)
		if err != nil {
			log.Printf("Warning: failed to extract EXIF from %s: %v", filePath, err)
			metadata = &exif.Metadata{}
		}

		var setDefault string
		if i > 0 && defaultSet != "" {
			setDefault = defaultSet
		}

		metadata, err = prompt.ForMetadata(filePath, metadata, setDefault)
		if err != nil {
			log.Fatalf("Failed to prompt for metadata: %v", err)
		}

		if i == 0 {
			defaultSet = metadata.Set
		}

		uploads = append(uploads, pending{filePath, metadata})
	}

	for _, upload := range uploads {
		if err := publish(client, scratch, upload.path, upload.metadata, masterPrefix, compressedPrefix, width, crf, preset); err != nil {
			log.Printf("Upload failed: %v", err)
			continue
		}

		fmt.Printf("Uploaded %s successfully\n", filepath.Base(upload.path))
	}
}

// publish writes a master and its compressed sibling. The sibling is encoded
// first: a failure there leaves the bucket untouched rather than half-published.
func publish(client *r2.Client, scratch, filePath string, metadata *exif.Metadata, masterPrefix, compressedPrefix string, width, crf, preset int) error {
	if err := metadata.Validate(); err != nil {
		return err
	}

	masterKey := masterKeyFor(masterPrefix, filePath)
	compressedKey := compressedKeyFor(compressedPrefix, masterKey)
	compressedPath := filepath.Join(scratch, filepath.Base(compressedKey))

	if err := ffmpeg.Compress(filePath, compressedPath, width, crf, preset); err != nil {
		return err
	}

	if err := uploadWithDimensions(client, filePath, masterKey, metadata, masterCacheControl); err != nil {
		return err
	}

	return uploadWithDimensions(client, compressedPath, compressedKey, metadata, compressedCacheControl)
}

// uploadWithDimensions writes the file's own width and height into its metadata,
// so the gallery can reserve a cell's shape before the image arrives.
func uploadWithDimensions(client *r2.Client, filePath, key string, metadata *exif.Metadata, cacheControl string) error {
	width, height, err := ffmpeg.Dimensions(filePath)
	if err != nil {
		return err
	}

	custom := metadata.Map()
	custom["width"] = strconv.Itoa(width)
	custom["height"] = strconv.Itoa(height)

	return client.Upload(key, filePath, custom, cacheControl)
}
