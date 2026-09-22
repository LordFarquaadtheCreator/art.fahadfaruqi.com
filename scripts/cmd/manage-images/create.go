package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
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

	cmd.Flags().String(
		"inherit",
		"",
		"Upload with the metadata of the existing object of the same name and this extension, without prompting, e.g. --inherit .png",
	)

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

	inherit, _ := cmd.Flags().GetString("inherit")

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
			sibling := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath)) + inherit

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
