package prompt

import (
	"bufio"
	"os"
	"strings"
	"testing"

	"manage-images/internal/exif"
)

func TestSetDescriptionsForgetNothingButNeverInvent(t *testing.T) {
	descriptions := NewSetDescriptions()

	if got := descriptions.Default("sam"); got != "" {
		t.Fatalf("an undescribed set offers %q, want nothing", got)
	}

	descriptions.Record("sam", "Night streets, lit by the last shop closing.")
	if got := descriptions.Default("sam"); got != "Night streets, lit by the last shop closing." {
		t.Fatalf("sam offers %q", got)
	}
	if got := descriptions.Default("pauline"); got != "" {
		t.Fatalf("pauline offers %q from another set", got)
	}

	// A blank answer says nothing about the set, so it must not erase what is known.
	descriptions.Record("sam", "")
	if got := descriptions.Default("sam"); got == "" {
		t.Fatal("a blank answer erased the set's description")
	}
}

// The description belongs to the set: the next file of the same set is offered it.
func TestPromptOffersTheSetsDescriptionToTheNextFile(t *testing.T) {
	descriptions := NewSetDescriptions()

	withStdin(t, "One\n\nDesc one\nsam\nNight streets, lit by the last shop closing.\n1\n")()

	first, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", descriptions, NewNumbers())
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if first.SetDescription != "Night streets, lit by the last shop closing." {
		t.Fatalf("first file took %q", first.SetDescription)
	}

	capture := withStdin(t, "Two\n\nDesc two\n\n\n\n")

	second, err := ForMetadata(bufio.NewReader(os.Stdin), "b.jpg", &exif.Metadata{Exif: map[string]string{}}, "sam", descriptions, NewNumbers())
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if second.SetDescription != first.SetDescription {
		t.Fatalf("the blank answer did not take the set's description: %q", second.SetDescription)
	}
	if got := second.Map()["set_description"]; got != first.SetDescription {
		t.Fatalf("Map wrote set_description as %q", got)
	}
	if printed := capture(); !strings.Contains(printed, "Set description [Night streets, lit by the last shop closing.]: ") {
		t.Fatalf("the prompt did not offer the set's description:\n%s", printed)
	}
}

// Another set starts blank rather than borrowing the description of the last one.
func TestPromptDoesNotCarryADescriptionToAnotherSet(t *testing.T) {
	descriptions := NewSetDescriptions()
	descriptions.Record("sam", "Night streets, lit by the last shop closing.")

	capture := withStdin(t, "One\n\nDesc\npauline\nA different street entirely.\n1\n")

	metadata, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", descriptions, NewNumbers())
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.SetDescription != "A different street entirely." {
		t.Fatalf("accepted %q", metadata.SetDescription)
	}
	if printed := capture(); strings.Contains(printed, "Set description [") {
		t.Fatalf("another set was offered sam's description:\n%s", printed)
	}
}

// A blank answer for a set the run has not described leaves the field empty, which is
// what the UI reads as "no description": nothing is invented to fill it.
func TestPromptLeavesAnUnansweredSetDescriptionEmpty(t *testing.T) {
	capture := withStdin(t, "One\n\nDesc\nsam\n\n1\n")

	metadata, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", NewSetDescriptions(), NewNumbers())
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.SetDescription != "" {
		t.Fatalf("a blank answer produced %q", metadata.SetDescription)
	}
	if got, present := metadata.Map()["set_description"]; !present || got != "" {
		t.Fatalf("Map wrote set_description as %q (present %v)", got, present)
	}
	if printed := capture(); !strings.Contains(printed, "Set description: ") {
		t.Fatalf("the prompt did not ask for a description:\n%s", printed)
	}
}
