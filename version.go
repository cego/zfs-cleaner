package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version can be overridden at compile time using -ldflags '-X main.Version=...'.
var version = "dev"

func init() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information and exit",
		RunE: func(_ *cobra.Command, args []string) error {
			printVersion()

			return nil
		},
	}

	rootCmd.AddCommand(versionCmd)
}

func printVersion() {
	if verbose {
		fmt.Printf("zfs-cleaner %s\n", version)
		fmt.Printf("\nhttps://github.com/cego/zfs-cleaner\n")

		os.Exit(0)
	}

	fmt.Printf("%s\n", version)

	os.Exit(0)
}
