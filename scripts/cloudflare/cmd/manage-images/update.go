package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"manage-images/internal/r2"
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [key...]",
		Short: "Edit the metadata of matching images, renaming them when -name is given",
		Run:   runUpdate,
	}

	addTargetFlags(cmd)
	cmd.Flags().String("name", "", "New file name; the original extension is kept when the new name has none")
	cmd.Flags().String("title", "", "Title")
	cmd.Flags().String("alt-text", "", "Alt text")
	cmd.Flags().String("description", "", "Description")
	cmd.Flags().String("set", "", "Set the image belongs to")
	cmd.Flags().String("set-description", "", "Description shown under the set's heading")
	cmd.Flags().Int("number", 0, "Order of the image within its set")
	cmd.Flags().String("uploaded", "", "Upload date the gallery shows, RFC 3339; defaults to the object's current one")

	return cmd
}

func runUpdate(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("dir")
	pattern, _ := cmd.Flags().GetString("pattern")
	name, _ := cmd.Flags().GetString("name")

	changes, err := metadataChanges(cmd)
	if err != nil {
		log.Fatal(err)
	}
	if len(changes) == 0 && !cmd.Flags().Changed("name") {
		log.Fatal("Nothing to update: pass -name or at least one metadata flag")
	}

	client := newClient()

	keys, err := selectKeys(client, dir, pattern, args, false)
	if err != nil {
		log.Fatal(err)
	}

	failed := false

	for _, key := range keys {
		info, err := client.Head(key)
		if err != nil {
			log.Printf("Update failed: %v", err)
			failed = true
			continue
		}

		newKey := renamedKey(key, name)
		// The copy replaces metadata wholesale, so layer the changes over what
		// the object already carries rather than overwriting it.
		metadata := mergeMetadata(info.Metadata, changes)
		metadata = pinUploaded(metadata, info.LastModified)
		contentType := contentTypeFor(newKey, info.ContentType)

		if err := client.Update(key, newKey, metadata, contentType, cacheControlFor(key, info.CacheControl)); err != nil {
			log.Printf("Update failed: %v", err)
			failed = true
			continue
		}

		if newKey == key {
			fmt.Printf("Updated %s\n", key)
		} else {
			fmt.Printf("Updated %s -> %s\n", key, newKey)
		}
	}

	if failed {
		log.Fatal("Some updates failed")
	}
}

// metadataChanges collects the metadata fields the caller passed as flags. A flag
// that was not passed leaves that field alone.
func metadataChanges(cmd *cobra.Command) (map[string]string, error) {
	fields := map[string]string{
		"title":           "title",
		"alt-text":        "altText",
		"description":     "description",
		"set":             "set",
		"set-description": "set_description",
	}

	changes := map[string]string{}

	for flag, key := range fields {
		if !cmd.Flags().Changed(flag) {
			continue
		}

		value, _ := cmd.Flags().GetString(flag)
		if value == "" {
			return nil, fmt.Errorf("-%s cannot be empty", flag)
		}

		changes[key] = value
	}

	if cmd.Flags().Changed("number") {
		number, _ := cmd.Flags().GetInt("number")
		if number <= 0 {
			return nil, fmt.Errorf("-number must be greater than 0")
		}
		changes["number"] = fmt.Sprintf("%d", number)
	}

	if cmd.Flags().Changed("uploaded") {
		value, _ := cmd.Flags().GetString("uploaded")
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return nil, fmt.Errorf("-uploaded must be RFC 3339: %w", err)
		}
		changes["uploaded"] = parsed.UTC().Format(uploadedLayout)
	}

	return changes, nil
}

// uploadedLayout is the form the bucket's curated `uploaded` field takes.
const uploadedLayout = "2006-01-02T15:04:05.000Z"

// pinUploaded keeps a metadata edit from re-dating the photograph: the copy
// resets LastModified, and the Worker reads that when the object carries no
// curated `uploaded` field. An object that already has one keeps it.
func pinUploaded(metadata map[string]string, lastModified time.Time) map[string]string {
	for held := range metadata {
		if strings.EqualFold(held, "uploaded") {
			return metadata
		}
	}

	metadata["uploaded"] = lastModified.UTC().Format(uploadedLayout)

	return metadata
}

// cacheControlFor keeps the header an object already carries, falling back to
// the one its tier implies: the copy that replaces its metadata drops it
// otherwise, leaving the object on the zone's 4h floor.
func cacheControlFor(key, current string) string {
	if current != "" {
		return current
	}

	switch {
	case strings.HasPrefix(key, compressedKeyPrefix):
		return compressedCacheControl
	case strings.HasPrefix(key, masterKeyPrefix):
		return masterCacheControl
	}

	return ""
}

// contentTypeFor prefers the content type the new name implies and falls back to
// the one the object already carried.
func contentTypeFor(key, current string) string {
	contentType, err := r2.ContentTypeFor(key)
	if err != nil || contentType == "" {
		contentType = current
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return contentType
}
