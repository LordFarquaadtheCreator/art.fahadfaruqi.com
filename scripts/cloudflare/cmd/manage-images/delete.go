package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [key...]",
		Short: "Delete matching images from the bucket",
		Run:   runDelete,
	}

	addTargetFlags(cmd)
	cmd.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt")

	return cmd
}

func runDelete(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("dir")
	pattern, _ := cmd.Flags().GetString("pattern")
	yes, _ := cmd.Flags().GetBool("yes")

	client := newClient()

	keys, err := selectKeys(client, dir, pattern, args, false)
	if err != nil {
		log.Fatal(err)
	}

	if !yes && !confirm(keys) {
		fmt.Println("Aborted.")
		return
	}

	failed := false

	for _, key := range keys {
		if err := client.Delete(key); err != nil {
			log.Printf("Delete failed: %v", err)
			failed = true
			continue
		}
		fmt.Printf("Deleted %s\n", key)
	}

	if failed {
		log.Fatal("Some deletes failed")
	}
}

func confirm(keys []string) bool {
	fmt.Printf("About to delete %d object(s):\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  %s\n", key)
	}
	fmt.Print("Continue? [y/N] ")

	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')

	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(answer)), "y")
}
