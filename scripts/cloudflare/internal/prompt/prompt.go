package prompt

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"manage-images/internal/exif"
)

func ForMetadata(filePath string, metadata *exif.Metadata, defaultSet string) (*exif.Metadata, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nExtracted EXIF data for %s:\n", filepath.Base(filePath))
	for k, v := range metadata.Exif {
		fmt.Printf("  %s: %s\n", k, v)
	}

	for metadata.Title == "" {
		fmt.Printf("Title for %s: ", filepath.Base(filePath))
		title, _ := reader.ReadString('\n')
		metadata.Title = strings.TrimSpace(title)
		if metadata.Title == "" {
			fmt.Println("Title is required. Please try again.")
		}
	}

	for metadata.AltText == "" {
		fmt.Print("Alt text: ")
		altText, _ := reader.ReadString('\n')
		altText = strings.TrimSpace(altText)
		if altText == "" {
			metadata.AltText = metadata.Title
		} else {
			metadata.AltText = altText
		}
	}

	for metadata.Description == "" {
		fmt.Print("Description: ")
		description, _ := reader.ReadString('\n')
		metadata.Description = strings.TrimSpace(description)
		if metadata.Description == "" {
			fmt.Println("Description is required. Please try again.")
		}
	}

	for metadata.Set == "" {
		if defaultSet != "" {
			fmt.Printf("Set [%s]: ", defaultSet)
			set, _ := reader.ReadString('\n')
			set = strings.TrimSpace(set)
			if set == "" {
				metadata.Set = defaultSet
			} else {
				metadata.Set = set
			}
		} else {
			fmt.Print("Set: ")
			set, _ := reader.ReadString('\n')
			metadata.Set = strings.TrimSpace(set)
			if metadata.Set == "" {
				fmt.Println("Set is required. Please try again.")
			}
		}
	}

	for metadata.Number == 0 {
		fmt.Print("Number (required for ordering): ")
		numberStr, _ := reader.ReadString('\n')
		numberStr = strings.TrimSpace(numberStr)
		if numberStr == "" {
			fmt.Println("Number is required. Please try again.")
			continue
		}
		if number, err := strconv.Atoi(numberStr); err == nil {
			metadata.Number = number
		} else {
			fmt.Printf("Invalid number: %v. Please try again.\n", err)
		}
	}

	return metadata, nil
}
