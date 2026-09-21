package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"

	"manage-images/internal/exif"
	"manage-images/internal/prompt"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [file...]",
		Aliases: []string{"upload"},
		Short:   "Upload local images to the bucket with metadata",
		Run:     runCreate,
	}

	addTargetFlags(cmd)

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) {
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
			log.Fatal("Usage: manage-images create -dir <directory> -pattern <pattern> OR manage-images create -dir <directory> <file1> [file2...]")
		}
		for _, fileName := range args {
			filePaths = append(filePaths, filepath.Join(workingDir, fileName))
		}
	}

	if len(filePaths) == 0 {
		log.Fatal("No files to upload")
	}

	client := newClient()

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
