package prompt

import (
	"os"
	"strings"
	"testing"

	"manage-images/internal/exif"
)

// withStdin swaps the process's stdin for the length of a test: ForMetadata reads it
// directly rather than taking a reader.
func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	original := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = original })

	// Write in the background: a full pipe would block the test before the reader runs.
	go func() {
		defer w.Close()
		_, _ = w.WriteString(input)
	}()
	devNull, _ := os.Open(os.DevNull)
	stdout := os.Stdout
	os.Stdout = devNull
	t.Cleanup(func() { os.Stdout = stdout })
}

const answers = "A title\n\nA description\nA set\n7\n"

func TestPromptDeclinesZoom(t *testing.T) {
	withStdin(t, answers+"n\n")

	metadata, err := ForMetadata("x.jpg", &exif.Metadata{Exif: map[string]string{}}, "")
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.ZoomLevel != 0 || metadata.ZoomDuration != 0 {
		t.Fatalf("declined zoom still wrote %v/%v", metadata.ZoomLevel, metadata.ZoomDuration)
	}
}

// The defect this replaced: a guard of "not yet zero" spun forever on a legal X of 0.
func TestPromptAcceptsZeroCoordinate(t *testing.T) {
	withStdin(t, answers+"y\n0\n0\n0.5\n1200\n")

	metadata, err := ForMetadata("x.jpg", &exif.Metadata{Exif: map[string]string{}}, "")
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.ZoomX != 0 || metadata.ZoomY != 0 {
		t.Fatalf("parsed x/y as %v/%v, want 0/0", metadata.ZoomX, metadata.ZoomY)
	}
	if metadata.ZoomLevel != 0.5 || metadata.ZoomDuration != 1200 {
		t.Fatalf("parsed level/duration as %v/%v, want 0.5/1200", metadata.ZoomLevel, metadata.ZoomDuration)
	}
}

func TestPromptRetriesBadValues(t *testing.T) {
	in := answers + "y\n" + strings.Join([]string{
		"2",    // X out of range, rejected
		"0.2",  // X accepted
		"-1",   // Y out of range, rejected
		"0.8",  // Y accepted
		"0",    // level not positive, rejected
		"2",    // level accepted
		"soon", // duration not an integer, rejected
		"1500", // duration accepted
	}, "\n") + "\n"
	withStdin(t, in)

	metadata, err := ForMetadata("x.jpg", &exif.Metadata{Exif: map[string]string{}}, "")
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.ZoomX != 0.2 || metadata.ZoomY != 0.8 {
		t.Fatalf("parsed x/y as %v/%v, want 0.2/0.8", metadata.ZoomX, metadata.ZoomY)
	}
	if metadata.ZoomLevel != 2 || metadata.ZoomDuration != 1500 {
		t.Fatalf("parsed level/duration as %v/%v, want 2/1500", metadata.ZoomLevel, metadata.ZoomDuration)
	}

	// And what it collected must be what Map writes.
	out := metadata.Map()
	if out["zoom_x"] != "0.2" || out["zoom_y"] != "0.8" || out["zoom_level"] != "2" || out["zoom_duration"] != "1500" {
		t.Fatalf("Map wrote %v", out)
	}
}

func TestPromptSkipsZoomWhenAlreadySet(t *testing.T) {
	withStdin(t, answers)

	existing := &exif.Metadata{Exif: map[string]string{}, ZoomX: 0.3, ZoomY: 0.4, ZoomLevel: 3, ZoomDuration: 800}
	metadata, err := ForMetadata("x.jpg", existing, "")
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.ZoomLevel != 3 || metadata.ZoomX != 0.3 {
		t.Fatalf("an inherited zoom was overwritten: %v/%v", metadata.ZoomX, metadata.ZoomLevel)
	}
}
