package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// shorthands collects every shorthand in the command tree, and whether it takes a value.
func shorthands(root *cobra.Command) map[byte]bool {
	byLetter := make(map[byte]bool)

	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		visit := func(flag *pflag.Flag) {
			if flag.Shorthand != "" {
				byLetter[flag.Shorthand[0]] = flag.Value.Type() != "bool"
			}
		}

		cmd.Flags().VisitAll(visit)
		cmd.PersistentFlags().VisitAll(visit)

		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)

	return byLetter
}

// validateShorthands reads -dir the way a cluster is meant to read: -d -i -r. pflag
// would take "ir" as -d's value instead, so the letters have to be checked before
// parsing, where the mistake can still be reported.
func validateShorthands(root *cobra.Command, args []string) error {
	byLetter := shorthands(root)
	terminated := false

	for _, arg := range args {
		if terminated {
			continue
		}

		if arg == "--" {
			terminated = true

			continue
		}

		if !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			continue
		}

		body := arg[1:]
		if len(body) < 2 || strings.Contains(body, "=") {
			continue
		}

		for _, letter := range body {
			if _, ok := byLetter[byte(letter)]; !ok {
				return fmt.Errorf("unknown shorthand flag: %q in %s", letter, arg)
			}
		}

		// A cluster of booleans is fine, but a shorthand that wants a value has to
		// stand alone — otherwise its value is silently whatever follows it.
		if byLetter[body[0]] {
			return fmt.Errorf("shorthand -%s takes a value and cannot be combined: write -%s <value> or -%s=<value>", body[:1], body[:1], body[:1])
		}
	}

	return nil
}
