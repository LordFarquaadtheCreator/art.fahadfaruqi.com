package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func testFlagTree() *cobra.Command {
	root := &cobra.Command{Use: "manage-images"}
	root.AddCommand(newCreateCmd(), newReadCmd(), newUpdateCmd(), newDeleteCmd())

	return root
}

func TestValidateShorthandsRejectsClusters(t *testing.T) {
	root := testFlagTree()

	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			"-dir is a cluster of d, i and r",
			[]string{"create", "-dir", "~/cars/edited", "-pattern", "*.png"},
			`unknown shorthand flag: 'i' in -dir`,
		},
		{
			"-pattern is a cluster starting with p",
			[]string{"read", "-pattern", "*.webp"},
			`unknown shorthand flag: 'a' in -pattern`,
		},
		{
			"two value-taking shorthands cannot be combined",
			[]string{"read", "-dp", "*.webp"},
			`shorthand -d takes a value and cannot be combined`,
		},
	}

	for _, c := range cases {
		err := validateShorthands(root, c.args)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: validateShorthands(%q) = %v, want %q", c.name, c.args, err, c.want)
		}
	}
}

func TestValidateShorthandsAccepts(t *testing.T) {
	root := testFlagTree()
	root.Flags().BoolP("alpha", "a", false, "")
	root.Flags().BoolP("beta", "b", false, "")

	accepted := [][]string{
		{"create", "-d", "masters", "-p", "*.webp"},
		{"read", "-d=art/master"},
		{"read", "--dir", "art/master", "--pattern", "*.webp"},
		{"read", "-d", "art/master"},
		{"create", "-ab"},
		{"create", "--", "-dir"},
		{"create", "sam-1.webp"},
	}

	for _, args := range accepted {
		if err := validateShorthands(root, args); err != nil {
			t.Errorf("validateShorthands(%q) = %v, want no error", args, err)
		}
	}
}
