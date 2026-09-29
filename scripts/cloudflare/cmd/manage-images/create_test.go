package main

import (
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
