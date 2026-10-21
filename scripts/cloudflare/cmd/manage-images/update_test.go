package main

import (
	"testing"
	"time"
)

// An object whose copy replaced its metadata has no cache header left; the tier
// its key belongs to has to put one back.
func TestCacheControlFor(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		current string
		want    string
	}{
		{"master without a header", "art/master/sam-1.webp", "", masterCacheControl},
		{"compressed without a header", "art/compressed/sam-1.avif", "", compressedCacheControl},
		{"existing header wins", "art/compressed/sam-1.avif", "public, max-age=60", "public, max-age=60"},
		{"a key outside the tiers keeps none", "www/app.js", "", ""},
	}

	for _, test := range tests {
		if got := cacheControlFor(test.key, test.current); got != test.want {
			t.Errorf("%s: cacheControlFor(%q, %q) = %q, want %q", test.name, test.key, test.current, got, test.want)
		}
	}
}

// The copy resets LastModified, which the gallery reads as the photograph's
// date, so an edit pins the date the object already had.
func TestPinUploaded(t *testing.T) {
	lastModified := time.Date(2026, 10, 19, 18, 44, 3, 652000000, time.UTC)

	pinned := pinUploaded(map[string]string{"title": "BMW E30"}, lastModified)
	if got, want := pinned["uploaded"], "2026-10-19T18:44:03.652Z"; got != want {
		t.Errorf("pinUploaded() uploaded = %q, want %q", got, want)
	}
	if got := pinned["title"]; got != "BMW E30" {
		t.Errorf("pinUploaded() dropped an existing field: title = %q", got)
	}

	held := pinUploaded(map[string]string{"uploaded": "2025-01-02T03:04:05.000Z"}, lastModified)
	if got, want := held["uploaded"], "2025-01-02T03:04:05.000Z"; got != want {
		t.Errorf("pinUploaded() overrode a curated date: got %q, want %q", got, want)
	}
}
