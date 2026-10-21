package prompt

import (
	"bufio"
	"os"
	"reflect"
	"strings"
	"testing"

	"manage-images/internal/exif"
)

func TestNumbersDefaultIsOnePastTheHighestAcknowledged(t *testing.T) {
	numbers := NewNumbers()

	if got := numbers.Default("sam"); got != 1 {
		t.Fatalf("a set with no number yet offers %d, want 1", got)
	}

	if !numbers.Acknowledge("sam", 4) {
		t.Fatal("a free number was refused")
	}
	if got := numbers.Default("sam"); got != 5 {
		t.Fatalf("after 4 the set offers %d, want 5", got)
	}

	// A gap the batch skipped stays a gap: the next default is not the hole.
	numbers.Acknowledge("sam", 9)
	if got := numbers.Default("sam"); got != 10 {
		t.Fatalf("after 4 and 9 the set offers %d, want 10", got)
	}

	if got := numbers.Default("pauline"); got != 1 {
		t.Fatalf("another set offers %d, want its own 1", got)
	}
}

func TestNumbersRefuseARepeatWithinOneSet(t *testing.T) {
	numbers := NewNumbers()

	if !numbers.Acknowledge("sam", 2) {
		t.Fatal("2 was refused in a set with no numbers")
	}
	if numbers.Acknowledge("sam", 2) {
		t.Fatal("a repeated number was accepted in the same set")
	}
	if !numbers.Acknowledge("pauline", 2) {
		t.Fatal("2 must be free in another set: numbers only collide within one")
	}
	if !numbers.Acknowledge("sam", 3) {
		t.Fatal("3 was refused in a set holding only 2")
	}

	if got, want := numbers.AcknowledgedIn("sam"), []int{2, 3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("AcknowledgedIn(sam) = %v, want %v", got, want)
	}
	if got := numbers.AcknowledgedIn("nothing"); len(got) != 0 {
		t.Fatalf("an untouched set reports %v, want nothing", got)
	}
}

// The first item of a run is offered 1, and an empty answer takes it.
func TestPromptOffersOneForTheFirstItem(t *testing.T) {
	capture := withStdin(t, "One\n\nDesc one\nsam\n\n")

	metadata, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", NewNumbers())
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.Number != 1 {
		t.Fatalf("accepted %d, want the offered 1", metadata.Number)
	}
	if printed := capture(); !strings.Contains(printed, "Number in sam [1]: ") {
		t.Fatalf("the prompt did not offer 1:\n%s", printed)
	}
}

// One book, two files: the section default is the set, the number default continues it.
func TestPromptOffersTheNextNumberForTheSecondFile(t *testing.T) {
	numbers := NewNumbers()

	withStdin(t, "One\n\nDesc one\nsam\n6\n")()

	first, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", numbers)
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if first.Number != 6 {
		t.Fatalf("accepted %d, want 6", first.Number)
	}

	capture := withStdin(t, "Two\n\nDesc two\n\n\n")

	second, err := ForMetadata(bufio.NewReader(os.Stdin), "b.jpg", &exif.Metadata{Exif: map[string]string{}}, "sam", numbers)
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if second.Set != "sam" {
		t.Fatalf("the set default was not taken: %q", second.Set)
	}
	if second.Number != 7 {
		t.Fatalf("accepted %d, want the offered 7", second.Number)
	}
	if printed := capture(); !strings.Contains(printed, "Number in sam [7]: ") {
		t.Fatalf("the prompt did not offer 7:\n%s", printed)
	}
}

// A repeat is refused, the refusal names what the set holds, and the answer is asked again.
func TestPromptRefusesANumberTheRunAlreadyTook(t *testing.T) {
	numbers := NewNumbers()
	numbers.Acknowledge("sam", 3)

	capture := withStdin(t, "One\n\nDesc\nsam\n3\n4\n")

	metadata, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", numbers)
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.Number != 4 {
		t.Fatalf("accepted %d, want 4 after 3 was refused", metadata.Number)
	}
	if printed := capture(); !strings.Contains(printed, "Number 3 is already taken in sam (3).") {
		t.Fatalf("the refusal did not name the taken number:\n%s", printed)
	}
}

// Gaps are legal: a free number below the set's highest is accepted, and the default
// still follows the highest rather than the number just typed.
func TestPromptFillsAGapWithoutMovingTheDefault(t *testing.T) {
	numbers := NewNumbers()
	numbers.Acknowledge("sam", 7)

	capture := withStdin(t, "One\n\nDesc\nsam\n4\n")

	metadata, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", numbers)
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.Number != 4 {
		t.Fatalf("a gap was refused: accepted %d, want 4", metadata.Number)
	}
	if printed := capture(); !strings.Contains(printed, "Number in sam [8]: ") {
		t.Fatalf("the gap moved the default off 8:\n%s", printed)
	}
}

// Numbers stay positive: 0 is the not-answered sentinel, and a negative position
// would sort ahead of the set's first photograph.
func TestPromptRejectsZeroAndNegativeNumbers(t *testing.T) {
	capture := withStdin(t, "One\n\nDesc\nsam\n0\n-2\n1\n")

	metadata, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", NewNumbers())
	if err != nil {
		t.Fatalf("ForMetadata: %v", err)
	}
	if metadata.Number != 1 {
		t.Fatalf("accepted %d, want 1", metadata.Number)
	}
	if printed := capture(); strings.Count(printed, "Number must be 1 or more.") != 2 {
		t.Fatalf("0 and -2 were not both refused:\n%s", printed)
	}
}

// A script that runs out before a required answer is an error, not an endless
// re-ask: the loop used to print "required" at full speed until it was killed.
func TestPromptFailsWhenTheScriptEndsEarly(t *testing.T) {
	withStdin(t, "A title\n\n")

	_, err := ForMetadata(bufio.NewReader(os.Stdin), "a.jpg", &exif.Metadata{Exif: map[string]string{}}, "", NewNumbers())
	if err == nil {
		t.Fatal("a script ending before the description was accepted")
	}
	if !strings.Contains(err.Error(), "stdin ended before the description") {
		t.Fatalf("error = %v, want it to name the missing answer", err)
	}
}
