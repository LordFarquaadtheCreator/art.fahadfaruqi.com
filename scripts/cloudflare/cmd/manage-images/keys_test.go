package main

import "testing"

func TestLiteralPrefix(t *testing.T) {
	tests := map[string]string{
		"paintings/*.jpg":    "paintings/",
		"d/w800/a*.webp":     "d/w800/",
		"sun*.jpg":           "",
		"*.jpg":              "",
		"paintings/sun.jpg":  "paintings/sun.jpg",
		"paintings/a?b.jpg":  "paintings/",
		"paintings/[ab].jpg": "paintings/",
	}

	for glob, want := range tests {
		if got := literalPrefix(glob); got != want {
			t.Errorf("literalPrefix(%q) = %q, want %q", glob, got, want)
		}
	}
}

func TestRenamedKey(t *testing.T) {
	tests := []struct {
		key  string
		name string
		want string
	}{
		{"sunset.jpg", "", "sunset.jpg"},
		{"sunset.jpg", "beach.jpg", "beach.jpg"},
		{"sunset.jpg", "beach", "beach.jpg"},
		{"paintings/sunset.jpg", "beach", "paintings/beach.jpg"},
		{"paintings/sunset.jpg", "drafts/beach.png", "paintings/drafts/beach.png"},
		{"a/b/sunset.JPG", "beach", "a/b/beach.JPG"},
	}

	for _, test := range tests {
		if got := renamedKey(test.key, test.name); got != test.want {
			t.Errorf("renamedKey(%q, %q) = %q, want %q", test.key, test.name, got, test.want)
		}
	}
}

func TestMergeMetadata(t *testing.T) {
	existing := map[string]string{
		"title":       "Old title",
		"alttext":     "Kept alt text",
		"description": "Old description",
		"Make":        "Nikon",
	}

	changes := map[string]string{
		"title": "New title",
		"set":   "City",
	}

	got := mergeMetadata(existing, changes)

	want := map[string]string{
		"title":       "New title",
		"alttext":     "Kept alt text",
		"description": "Old description",
		"Make":        "Nikon",
		"set":         "City",
	}

	if len(got) != len(want) {
		t.Fatalf("mergeMetadata returned %d keys, want %d: %v", len(got), len(want), got)
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("mergeMetadata[%q] = %q, want %q", key, got[key], value)
		}
	}
}

// R2 lowercases metadata keys, so a change must not leave a stale twin behind.
func TestMergeMetadataReplacesCaseInsensitiveKey(t *testing.T) {
	got := mergeMetadata(map[string]string{"alttext": "Old alt text"}, map[string]string{"altText": "New alt text"})

	if len(got) != 1 {
		t.Fatalf("mergeMetadata returned %d keys, want 1: %v", len(got), got)
	}
	if got["altText"] != "New alt text" {
		t.Errorf("mergeMetadata[altText] = %q, want %q", got["altText"], "New alt text")
	}
}

func TestMergeMetadataDoesNotMutateInputs(t *testing.T) {
	existing := map[string]string{"title": "Old title"}
	changes := map[string]string{"title": "New title"}

	mergeMetadata(existing, changes)

	if existing["title"] != "Old title" {
		t.Errorf("existing was mutated: %v", existing)
	}
	if changes["title"] != "New title" {
		t.Errorf("changes was mutated: %v", changes)
	}
}
