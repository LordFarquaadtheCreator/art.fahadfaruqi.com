package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"manage-images/internal/r2"
)

// selectKeys resolves the -dir/-pattern filter, or positional keys, to the keys
// that exist in the bucket. A pattern is globbed with the same semantics the
// create command uses for local files, but against bucket keys, so -dir acts as
// the key prefix: '-d paintings -p *.jpg' selects paintings/*.jpg. An empty
// filter selects every object only when all is true, which is what keeps a
// pattern-less delete from wiping the bucket.
func selectKeys(client *r2.Client, dir, pattern string, args []string, all bool) ([]string, error) {
	if pattern == "" && len(args) == 0 {
		if !all {
			return nil, fmt.Errorf("no objects selected: pass -pattern or one or more keys")
		}

		objects, err := client.List("")
		if err != nil {
			return nil, err
		}

		keys := make([]string, 0, len(objects))
		for _, object := range objects {
			keys = append(keys, object.Key)
		}

		sort.Strings(keys)

		return keys, nil
	}

	if pattern != "" {
		glob := filepath.Join(dir, pattern)

		objects, err := client.List(literalPrefix(glob))
		if err != nil {
			return nil, err
		}

		var keys []string
		for _, object := range objects {
			matched, err := filepath.Match(glob, object.Key)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %q: %w", glob, err)
			}
			if matched {
				keys = append(keys, object.Key)
			}
		}

		if len(keys) == 0 {
			return nil, fmt.Errorf("no objects match %s", glob)
		}

		sort.Strings(keys)

		return keys, nil
	}

	objects, err := client.List("")
	if err != nil {
		return nil, err
	}

	existing := make(map[string]bool, len(objects))
	for _, object := range objects {
		existing[object.Key] = true
	}

	keys := make([]string, 0, len(args))
	missing := make([]string, 0, len(args))
	seen := make(map[string]bool, len(args))

	for _, arg := range args {
		key := filepath.Join(dir, arg)
		if seen[key] {
			continue
		}
		seen[key] = true

		if existing[key] {
			keys = append(keys, key)
		} else {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("no such object: %s", strings.Join(missing, ", "))
	}

	return keys, nil
}

// literalPrefix returns the part of a glob that has to match exactly, so a
// listing can be narrowed to the keys the glob could possibly match.
func literalPrefix(glob string) string {
	index := strings.IndexAny(glob, "*?[")
	if index == -1 {
		return glob
	}

	prefix := glob[:index]
	if slash := strings.LastIndex(prefix, "/"); slash != -1 {
		return prefix[:slash+1]
	}

	return ""
}

// renamedKey applies -name to a key, keeping the original extension when the new
// name carries none. A name without a slash stays in the key's own prefix.
func renamedKey(key, name string) string {
	if name == "" {
		return key
	}

	if filepath.Ext(name) == "" {
		name += filepath.Ext(key)
	}

	return filepath.Join(filepath.Dir(key), name)
}

// mergeMetadata layers changes over the metadata an object already holds. R2
// lowercases metadata keys, so a change replaces any key that differs only in case.
func mergeMetadata(existing, changes map[string]string) map[string]string {
	merged := make(map[string]string, len(existing)+len(changes))
	for key, value := range existing {
		merged[key] = value
	}

	for key, value := range changes {
		for held := range merged {
			if strings.EqualFold(held, key) {
				delete(merged, held)
			}
		}
		merged[key] = value
	}

	return merged
}
