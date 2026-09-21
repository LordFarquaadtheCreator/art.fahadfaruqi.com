package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"upload-images/internal/config"
	"upload-images/internal/exif"
	"upload-images/internal/prompt"
	"upload-images/internal/r2"
)

func runUpload(cmd *cobra.Command, args []string) {
	workingDir, _ := cmd.Flags().GetString("dir")
	pattern, _ := cmd.Flags().GetString("pattern")

	var filePaths []string

	if pattern != "" {
		fullPattern := filepath.Join(workingDir, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			log.Fatalf("Failed to match pattern: %v", err)
		}
		if len(matches) == 0 {
			log.Fatalf("No files found matching pattern: %s", pattern)
		}
		filePaths = matches
	} else {
		if len(args) < 1 {
			log.Fatal("Usage: upload-images -dir <directory> -pattern <pattern> OR upload-images -dir <directory> <file1> [file2...]")
		}
		for _, fileName := range args {
			filePaths = append(filePaths, filepath.Join(workingDir, fileName))
		}
	}

	if len(filePaths) == 0 {
		log.Fatal("No files to upload")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := r2.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create R2 client: %v", err)
	}

	defaultSet := ""
	var uploads []struct {
		path     string
		metadata *exif.Metadata
	}

	for i, filePath := range filePaths {
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

		uploads = append(uploads, struct {
			path     string
			metadata *exif.Metadata
		}{filePath, metadata})
	}

	var wg sync.WaitGroup
	results := make(chan string, len(uploads))
	errors := make(chan error, len(uploads))

	for _, upload := range uploads {
		wg.Add(1)
		go func(path string, metadata *exif.Metadata) {
			defer wg.Done()
			if err := client.Upload(path, metadata); err != nil {
				errors <- err
			} else {
				results <- filepath.Base(path)
			}
		}(upload.path, upload.metadata)
	}

	wg.Wait()
	close(results)
	close(errors)

	for result := range results {
		fmt.Printf("Uploaded %s successfully\n", result)
	}

	for err := range errors {
		log.Printf("Upload failed: %v", err)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "upload-images",
		Short: "Upload images to Cloudflare R2 with metadata",
		Run:   runUpload,
	}

	rootCmd.Flags().StringP("dir", "d", ".", "Working directory for image files")
	rootCmd.Flags().StringP("pattern", "p", "", "File pattern to match (e.g., '*.jpg', 'IMG_*.png')")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
