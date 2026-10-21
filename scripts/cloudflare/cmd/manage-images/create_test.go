package main

import (
	"os"
	"os/exec"
	"testing"

	"manage-images/internal/ffmpeg"
)

func TestMasterKeyFor(t *testing.T) {
	cases := []struct {
		prefix   string
		filePath string
		want     string
	}{
		{"art/master", "pauline-1.webp", "art/master/pauline-1.webp"},
		{"art/master", "/Users/farquaad/exports/ceres-and-kimi-6.webp", "art/master/ceres-and-kimi-6.webp"},
		{"masters", "DSC_0222.JPG", "masters/DSC_0222.JPG"},
	}

	for _, c := range cases {
		if got := masterKeyFor(c.prefix, c.filePath); got != c.want {
			t.Errorf("masterKeyFor(%q, %q) = %q, want %q", c.prefix, c.filePath, got, c.want)
		}
	}
}

func TestCompressedKeyFor(t *testing.T) {
	cases := []struct {
		prefix    string
		masterKey string
		want      string
	}{
		{"art/compressed", "art/master/pauline-1.webp", "art/compressed/pauline-1.avif"},
		{"art/compressed", "art/master/ceres-and-kimi-6.webp", "art/compressed/ceres-and-kimi-6.avif"},
		{"compressed", "masters/DSC_0222.JPG", "compressed/DSC_0222.avif"},
	}

	for _, c := range cases {
		if got := compressedKeyFor(c.prefix, c.masterKey); got != c.want {
			t.Errorf("compressedKeyFor(%q, %q) = %q, want %q", c.prefix, c.masterKey, got, c.want)
		}
	}
}

func TestCompressedKeyForKeepsStemOfADottedName(t *testing.T) {
	got := compressedKeyFor("art/compressed", "art/master/ceres.and.kimi-6.webp")
	if want := "art/compressed/ceres.and.kimi-6.avif"; got != want {
		t.Errorf("compressedKeyFor dropped part of the stem: got %q, want %q", got, want)
	}
}

func TestAvailableReportsAMissingEncoder(t *testing.T) {
	t.Setenv("PATH", "")

	if err := ffmpeg.Available(); err == nil {
		t.Fatal("expected a missing encoder to be reported, so nothing uploads a master without its sibling")
	}
}

func TestCompressFailsOnAMissingInput(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is not installed")
	}

	err := ffmpeg.Compress("no-such-master.webp", t.TempDir()+"/out.avif", 1600, 22, 6)
	if err == nil {
		t.Fatal("expected an encode failure for a missing input")
	}
}

func TestAskForMetadataNumbersTheWholeBatch(t *testing.T) {
	withPromptStdin(t, "One\n\nFirst\nsam\n\nn\n"+"Two\n\nSecond\n\n\n")

	// A nil client is honest here: --inherit is empty, so nothing is fetched.
	uploads := askForMetadata(nil, []string{"a.jpg", "b.jpg"}, "", "art/master")
	if len(uploads) != 2 {
		t.Fatalf("collected %d files, want 2", len(uploads))
	}
	if got := uploads[0].metadata.Number; got != 1 {
		t.Fatalf("the first file took number %d, want the offered 1", got)
	}
	if got := uploads[1].metadata.Set; got != "sam" {
		t.Fatalf("the second file's set is %q, want the batch default sam", got)
	}
	if got := uploads[1].metadata.Number; got != 2 {
		t.Fatalf("the second file took number %d, want the offered 2", got)
	}
}

// Two files in one set cannot land on the same position: the repeat is refused and
// the file has to answer again.
func TestAskForMetadataRefusesASharedNumberWithinTheBatch(t *testing.T) {
	withPromptStdin(t, "One\n\nFirst\nsam\n1\nn\n"+"Two\n\nSecond\n\n1\n2\n")

	uploads := askForMetadata(nil, []string{"a.jpg", "b.jpg"}, "", "art/master")
	if len(uploads) != 2 {
		t.Fatalf("collected %d files, want 2", len(uploads))
	}
	if uploads[0].metadata.Number != 1 || uploads[1].metadata.Number != 2 {
		t.Fatalf("numbers are %d and %d, want 1 and 2", uploads[0].metadata.Number, uploads[1].metadata.Number)
	}
}

// withPromptStdin feeds the create prompts a script and keeps their chatter out of
// the test log: the prompt reads os.Stdin directly.
func withPromptStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdin pipe: %v", err)
	}
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("devnull: %v", err)
	}

	originalIn, originalOut := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = r, devNull
	t.Cleanup(func() {
		os.Stdin, os.Stdout = originalIn, originalOut
		devNull.Close()
	})

	go func() {
		defer w.Close()
		_, _ = w.WriteString(input)
	}()
}
