package main

import (
	"strings"
	"testing"
)

type bucket struct {
	masters    []string
	compressed []string
}

func (b bucket) conflicts(files ...string) []conflict {
	return detectConflicts(files, "art/master", "art/compressed", b.masters, b.compressed)
}

func keysOfConflicts(conflicts []conflict) (overwrites, owners []string) {
	for _, c := range conflicts {
		overwrites = append(overwrites, c.overwrites...)
		owners = append(owners, c.stemOwners...)
	}
	return overwrites, owners
}

func TestAFreeStemConflictsWithNothing(t *testing.T) {
	existing := bucket{
		masters:    []string{"art/master/home-1.webp", "art/master/home-2.webp"},
		compressed: []string{"art/compressed/home-1.avif", "art/compressed/home-2.avif"},
	}

	if got := existing.conflicts("home-3.png"); len(got) != 0 {
		t.Fatalf("a free stem reported %+v", got)
	}
}

// The failure this check exists for: the master key is free, but the compressed key
// derived from the stem belongs to another photograph.
func TestADifferentlyNamedFileOwningTheStemIsRefused(t *testing.T) {
	existing := bucket{
		masters:    []string{"art/master/home-1.webp"},
		compressed: []string{"art/compressed/home-1.avif"},
	}

	conflicts := existing.conflicts("home-1.png")
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %+v", conflicts)
	}

	overwrites, owners := keysOfConflicts(conflicts)
	if want := "art/compressed/home-1.avif"; len(overwrites) != 1 || overwrites[0] != want {
		t.Errorf("overwrites = %v, want [%s]", overwrites, want)
	}
	if want := "art/master/home-1.webp"; len(owners) != 1 || owners[0] != want {
		t.Errorf("stem owners = %v, want [%s]", owners, want)
	}
}

func TestRepublishingTheSameNameIsAConflict(t *testing.T) {
	existing := bucket{
		masters:    []string{"art/master/home-1.webp"},
		compressed: []string{"art/compressed/home-1.avif"},
	}

	conflicts := existing.conflicts("home-1.webp")
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %+v", conflicts)
	}

	overwrites, owners := keysOfConflicts(conflicts)
	if len(overwrites) != 2 {
		t.Errorf("overwrites = %v, want both the master and its sibling", overwrites)
	}
	if len(owners) != 0 {
		t.Errorf("stem owners = %v, want none: this file is the stem's own master", owners)
	}
}

func TestAnOrphanCompressedSiblingIsAConflict(t *testing.T) {
	existing := bucket{compressed: []string{"art/compressed/home-9.avif"}}

	conflicts := existing.conflicts("home-9.png")
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %+v", conflicts)
	}

	overwrites, _ := keysOfConflicts(conflicts)
	if want := "art/compressed/home-9.avif"; len(overwrites) != 1 || overwrites[0] != want {
		t.Errorf("overwrites = %v, want [%s]", overwrites, want)
	}
}

func TestUnrelatedObjectsDoNotConflict(t *testing.T) {
	existing := bucket{
		masters:    []string{"art/master/home-1.webp", "art/master/pauline-2.webp"},
		compressed: []string{"art/compressed/home-1.avif", "art/compressed/pauline-2.avif"},
	}

	if got := existing.conflicts("home-4.png"); len(got) != 0 {
		t.Fatalf("an unrelated stem reported %+v", got)
	}
	if got := existing.conflicts("home-10.png"); len(got) != 0 {
		t.Fatalf("home-10 reported %+v against home-1", got)
	}
}

func TestADottedStemIsMatchedWhole(t *testing.T) {
	existing := bucket{
		masters:    []string{"art/master/ceres.and.kimi-6.png"},
		compressed: []string{"art/compressed/ceres.and.kimi-6.avif"},
	}

	conflicts := existing.conflicts("ceres.and.kimi-6.webp")
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %+v", conflicts)
	}

	_, owners := keysOfConflicts(conflicts)
	if want := "art/master/ceres.and.kimi-6.png"; len(owners) != 1 || owners[0] != want {
		t.Errorf("stem owners = %v, want [%s]", owners, want)
	}
}

func TestABatchReportsOnlyTheCollidingFiles(t *testing.T) {
	existing := bucket{
		masters:    []string{"art/master/home-1.webp"},
		compressed: []string{"art/compressed/home-1.avif"},
	}

	conflicts := existing.conflicts("/exports/home-1.png", "/exports/home-4.png", "/exports/home-5.png")
	if len(conflicts) != 1 {
		t.Fatalf("expected one conflict, got %+v", conflicts)
	}
	if !strings.HasSuffix(conflicts[0].file, "home-1.png") {
		t.Errorf("flagged %q, want home-1.png", conflicts[0].file)
	}
}

func TestAnEmptyBucketNeverConflicts(t *testing.T) {
	if got := (bucket{}).conflicts("home-1.png", "home-2.png"); len(got) != 0 {
		t.Fatalf("an empty bucket reported %+v", got)
	}
}
