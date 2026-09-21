package main

import (
	"log"

	"github.com/spf13/cobra"

	"manage-images/internal/config"
	"manage-images/internal/r2"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "manage-images",
		Short: "Create, read, update, and delete images in the R2 bucket",
	}

	rootCmd.AddCommand(newCreateCmd(), newReadCmd(), newUpdateCmd(), newDeleteCmd())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// addTargetFlags registers the flags that select which objects a command works
// on: a local directory or bucket prefix, and a glob to match within it.
func addTargetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("dir", "d", ".", "Working directory for image files, or key prefix within the bucket")
	cmd.Flags().StringP("pattern", "p", "", "Glob to match, e.g. '*.jpg' locally or 'paintings/*.jpg' in the bucket")
}

func newClient() *r2.Client {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := r2.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create R2 client: %v", err)
	}

	return client
}
