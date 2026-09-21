package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func newReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read [key...]",
		Short: "Print the CDN URL of every matching image",
		Run:   runRead,
	}

	addTargetFlags(cmd)

	return cmd
}

func runRead(cmd *cobra.Command, args []string) {
	dir, _ := cmd.Flags().GetString("dir")
	pattern, _ := cmd.Flags().GetString("pattern")

	client := newClient()

	if client.CDNBase() == "" {
		log.Fatal("No CDN base configured: set r2.cdn_base in config.yaml or CDN_BASE in the environment")
	}

	keys, err := selectKeys(client, dir, pattern, args, true)
	if err != nil {
		log.Fatal(err)
	}

	for _, key := range keys {
		fmt.Println(client.URL(key))
	}
}
