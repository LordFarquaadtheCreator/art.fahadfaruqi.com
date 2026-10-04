package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"

	"manage-images/internal/r2"
)

// A compressed sibling is named after the stem alone (art/compressed/<stem>.avif),
// so the stem is the slot rather than the file name, and two masters sharing one are
// merged by the Worker into a single entry.

type conflict struct {
	file       string
	overwrites []string
	stemOwners []string
}

func detectConflicts(files []string, masterPrefix, compressedPrefix string, masterKeys, compressedKeys []string) []conflict {
	masters := make(map[string]bool, len(masterKeys))
	for _, key := range masterKeys {
		masters[key] = true
	}

	compressed := make(map[string]bool, len(compressedKeys))
	for _, key := range compressedKeys {
		compressed[key] = true
	}

	byStem := make(map[string][]string, len(masterKeys))
	for _, key := range masterKeys {
		byStem[stemOf(key)] = append(byStem[stemOf(key)], key)
	}

	var found []conflict

	for _, file := range files {
		masterKey := masterKeyFor(masterPrefix, file)
		compressedKey := compressedKeyFor(compressedPrefix, masterKey)

		var overwrites []string
		if masters[masterKey] {
			overwrites = append(overwrites, masterKey)
		}
		if compressed[compressedKey] {
			overwrites = append(overwrites, compressedKey)
		}

		var owners []string
		for _, existing := range byStem[stemOf(masterKey)] {
			if existing != masterKey {
				owners = append(owners, existing)
			}
		}

		if len(overwrites) == 0 && len(owners) == 0 {
			continue
		}

		sort.Strings(overwrites)
		sort.Strings(owners)
		found = append(found, conflict{file: file, overwrites: overwrites, stemOwners: owners})
	}

	return found
}

func stemOf(key string) string {
	return strings.TrimSuffix(filepath.Base(key), filepath.Ext(key))
}

func keysOf(objects []r2.Object) []string {
	keys := make([]string, 0, len(objects))
	for _, object := range objects {
		keys = append(keys, object.Key)
	}

	return keys
}

// guardAgainstOverwrites refuses a publish that would replace a photograph already
// in the bucket. --inherit is treated as declared intent because re-encoding an item
// in place is what it is for; it is not a general licence to overwrite.
func guardAgainstOverwrites(client *r2.Client, files []string, masterPrefix, compressedPrefix, inherit string, replace bool) {
	masters, err := client.List(masterPrefix + "/")
	if err != nil {
		log.Fatalf("Cannot check the bucket for collisions: %v", err)
	}

	compressed, err := client.List(compressedPrefix + "/")
	if err != nil {
		log.Fatalf("Cannot check the bucket for collisions: %v", err)
	}

	conflicts := detectConflicts(files, masterPrefix, compressedPrefix, keysOf(masters), keysOf(compressed))
	if len(conflicts) == 0 {
		return
	}

	if replace {
		fmt.Println("Overwriting existing objects:")
		for _, c := range conflicts {
			targets := append(append([]string{}, c.overwrites...), c.stemOwners...)
			fmt.Printf("  %s -> %s\n", filepath.Base(c.file), strings.Join(targets, ", "))
		}
		return
	}

	// --inherit supersedes only the sibling it reads, so a shared stem survives it.
	if inherit != "" {
		for _, c := range conflicts {
			if len(c.stemOwners) > 0 {
				fmt.Fprintf(log.Writer(),
					"Note: %s leaves the stem shared with %s; delete the superseded file, or the gallery merges the two into one entry.\n",
					filepath.Base(c.file), strings.Join(c.stemOwners, ", "))
			}
		}
		return
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Refusing to publish %d file(s) that would overwrite existing photographs:\n", len(conflicts))
	for _, c := range conflicts {
		fmt.Fprintf(&b, "\n  %s\n", filepath.Base(c.file))
		for _, key := range c.stemOwners {
			fmt.Fprintf(&b, "      %s already holds this stem%s\n", key, ownerNote(client, key))
		}
		for _, key := range c.overwrites {
			fmt.Fprintf(&b, "      %s would be replaced%s\n", key, ownerNote(client, key))
		}
	}
	b.WriteString("\n  The compressed sibling is named after the stem, so the stem is the slot however\n")
	b.WriteString("  the file is named: two masters under one stem swap a photograph's rendered copy\n")
	b.WriteString("  while leaving its master and metadata in place.\n")
	b.WriteString("\nPass --replace to overwrite anyway, or give the file a free stem.\n")

	log.Fatal(b.String())
}

// ownerNote names the photograph holding a key. Best effort: an object the bucket
// will not answer for is not a reason to fail the check.
func ownerNote(client *r2.Client, key string) string {
	title, set, number := titleOf(client, key)
	if title == "" {
		return ""
	}

	note := fmt.Sprintf(" — %q", title)
	if set != "" {
		note += fmt.Sprintf(" in %s", set)
		if number != "" {
			note += fmt.Sprintf(" no. %s", number)
		}
	}

	return note
}

// titleOf resolves a compressed key through the master sharing its stem, which is
// where that metadata actually lives.
func titleOf(client *r2.Client, key string) (title, set, number string) {
	if filepath.Ext(key) == ".avif" {
		stem := stemOf(key)
		masters, err := client.List(filepath.Dir(key) + "/")
		if err != nil {
			return "", "", ""
		}
		for _, object := range masters {
			if stemOf(object.Key) == stem {
				key = object.Key
				break
			}
		}
	}

	info, err := client.Head(key)
	if err != nil {
		return "", "", ""
	}

	return info.Metadata["title"], info.Metadata["set"], info.Metadata["number"]
}
