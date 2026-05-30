package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// version is the CLI version. Override at build time with:
//
//	-ldflags "-X jsb-cli/cmd.version=x.y.z"
var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "jsb",
	Short:   "JSONBank CLI",
	Long:    "jsb is the command-line interface for JSONBank — store, fetch, and manage JSON documents.",
	Version: version,
	// Drop cobra's auto-generated `completion` command.
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

// Execute runs the root command. It is the single entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints the error; just exit non-zero.
		os.Exit(1)
	}
}
