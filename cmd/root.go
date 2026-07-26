package cmd

import (
	"os"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is set at release time via:
//
//	-ldflags "-X github.com/jsonbankio/jsonbank-cli/cmd.version=x.y.z"
//
// When empty, resolveVersion falls back to the module version Go stamps
// into the binary, so `go install ...@vX.Y.Z` builds report correctly.
var version = ""

func resolveVersion() string {
	if version != "" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

var rootCmd = &cobra.Command{
	Use:     "jsb",
	Short:   "JSONBank CLI",
	Long:    "Command-line interface for JSONBank — store, fetch, and manage JSON documents.",
	Version: resolveVersion(),
	// Don't print usage after a runtime error (only the error itself).
	SilenceUsage: true,
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
