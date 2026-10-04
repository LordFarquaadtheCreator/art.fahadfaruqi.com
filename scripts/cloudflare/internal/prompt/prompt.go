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

	if metadata.ZoomLevel > 0 {
		return metadata, nil
	}

	// The zoom is optional, so it is asked for once and collected whole. The guards
	// below test "not yet set" rather than "not zero": 0 is a legal coordinate, and a
	// loop waiting for a non-zero value never leaves.
	fmt.Print("Add zoom animation? (y/N): ")
	zoomAnswer, _ := reader.ReadString('\n')
	zoomAnswer = strings.TrimSpace(strings.ToLower(zoomAnswer))
	if zoomAnswer != "y" && zoomAnswer != "yes" {
		return metadata, nil
	}

	var (
		zoomX        float64
		zoomY        float64
		zoomLevel    float64
		zoomDuration int
		haveX        bool
		haveY        bool
		haveLevel    bool
		haveDuration bool
	)

	for !haveX {
		fmt.Print("Zoom X (0-1, the point to settle on, 0.5 = centre): ")
		xStr, _ := reader.ReadString('\n')
		xStr = strings.TrimSpace(xStr)
		if xStr == "" {
			continue
		}
		if x, err := strconv.ParseFloat(xStr, 64); err == nil && x >= 0 && x <= 1 {
			zoomX, haveX = x, true
		} else {
			fmt.Println("Invalid X. Must be between 0 and 1.")
		}
	}

	for !haveY {
		fmt.Print("Zoom Y (0-1, the point to settle on, 0.5 = centre): ")
		yStr, _ := reader.ReadString('\n')
		yStr = strings.TrimSpace(yStr)
		if yStr == "" {
			continue
		}
		if y, err := strconv.ParseFloat(yStr, 64); err == nil && y >= 0 && y <= 1 {
			zoomY, haveY = y, true
		} else {
			fmt.Println("Invalid Y. Must be between 0 and 1.")
		}
	}

	for !haveLevel {
		fmt.Print("Zoom level (positive; 1 = no move, 2 = 2x in, 0.5 = 2x out): ")
		levelStr, _ := reader.ReadString('\n')
		levelStr = strings.TrimSpace(levelStr)
		if levelStr == "" {
			continue
		}
		if level, err := strconv.ParseFloat(levelStr, 64); err == nil && level > 0 {
			zoomLevel, haveLevel = level, true
		} else {
			fmt.Println("Invalid level. Must be a positive number.")
		}
	}

	for !haveDuration {
		fmt.Print("Zoom duration (milliseconds, e.g. 1500): ")
		durationStr, _ := reader.ReadString('\n')
		durationStr = strings.TrimSpace(durationStr)
		if durationStr == "" {
			continue
		}
		if duration, err := strconv.Atoi(durationStr); err == nil && duration > 0 {
			zoomDuration, haveDuration = duration, true
		} else {
			fmt.Println("Invalid duration. Must be a positive integer.")
		}
	}

	metadata.ZoomX = zoomX
	metadata.ZoomY = zoomY
	metadata.ZoomLevel = zoomLevel
	metadata.ZoomDuration = zoomDuration

	return metadata, nil
}
