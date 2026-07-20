package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/dustin/go-humanize"
	jsonbank "github.com/jsonbankio/go-sdk"
	"github.com/jsonbankio/go-sdk/types"
	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	"jsb-cli/internal/app"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "View, download, and inspect JSONBank files",
}

var fileViewCmd = &cobra.Command{
	Use:   "view <path>",
	Short: "Print a file's JSON to the terminal",
	Args:  cobra.ExactArgs(1),
	RunE:  runFileView,
}

var fileDownloadCmd = &cobra.Command{
	Use:   "download <path> <dest>",
	Short: "Download a file and save it locally",
	Args:  cobra.ExactArgs(2),
	RunE:  runFileDownload,
}

var fileMetaCmd = &cobra.Command{
	Use:   "meta <path>",
	Short: "Show a file's metadata",
	Args:  cobra.ExactArgs(1),
	RunE:  runFileMeta,
}

var fileMetaRaw bool

func init() {
	fileMetaCmd.Flags().BoolVar(&fileMetaRaw, "raw", false, "print the metadata as raw JSON instead of a formatted summary")

	fileCmd.AddCommand(fileViewCmd, fileDownloadCmd, fileMetaCmd)
	rootCmd.AddCommand(fileCmd)
}

func runFileView(cmd *cobra.Command, args []string) error {
	jsb, err := authedClient()
	if err != nil {
		return err
	}

	path := cleanPath(args[0])
	content, reqErr := jsb.GetOwnContentAsString(path)
	if reqErr != nil {
		return fmt.Errorf("could not fetch %q: %s", path, reqErr.Message)
	}

	fmt.Print(string(renderJSON([]byte(content))))
	return nil
}

func runFileDownload(cmd *cobra.Command, args []string) error {
	jsb, err := authedClient()
	if err != nil {
		return err
	}

	remote, dest := cleanPath(args[0]), args[1]

	content, reqErr := jsb.GetOwnContentAsString(remote)
	if reqErr != nil {
		return fmt.Errorf("could not fetch %q: %s", remote, reqErr.Message)
	}

	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return err
	}

	fmt.Printf("Saved %s to %s\n", remote, dest)
	return nil
}

func runFileMeta(cmd *cobra.Command, args []string) error {
	jsb, err := authedClient()
	if err != nil {
		return err
	}

	path := cleanPath(args[0])
	meta, reqErr := jsb.GetOwnDocumentMeta(path)
	if reqErr != nil {
		return fmt.Errorf("could not fetch meta for %q: %s", path, reqErr.Message)
	}

	if fileMetaRaw {
		b, err := json.Marshal(meta)
		if err != nil {
			return err
		}
		fmt.Print(string(renderJSON(b)))
		return nil
	}

	printMeta(meta)
	return nil
}

// authedClient initializes the app and returns an SDK client built from the
// active keys (environment or saved account). It errors when no keys are set.
func authedClient() (*jsonbank.Instance, error) {
	a, err := app.Init()
	if err != nil {
		return nil, err
	}

	keys := a.ActiveKeys()
	if keys.Public == "" {
		return nil, fmt.Errorf("you are not logged in — run: jsb auth login")
	}

	jsb := jsonbank.Init(jsonbank.Config{
		Host: a.Config.Host,
		Keys: jsonbank.Keys{Public: keys.Public, Private: keys.Private},
	})
	return &jsb, nil
}

// cleanPath strips surrounding slashes; a leading "/" would otherwise become an
// empty path segment that the API rejects.
func cleanPath(p string) string {
	return strings.Trim(p, "/")
}

// renderJSON indents the JSON, colorizing it when stdout is a terminal so that
// piped output stays plain.
func renderJSON(b []byte) []byte {
	b = pretty.Pretty(b)
	if term.IsTerminal(os.Stdout.Fd()) {
		b = pretty.Color(b, nil)
	}
	return b
}

// printMeta writes document metadata as an aligned label/value list.
func printMeta(m *types.DocumentMeta) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Name\t%s\n", m.Name)
	fmt.Fprintf(w, "Path\t%s\n", m.Path)
	fmt.Fprintf(w, "Project\t%s\n", m.Project)
	fmt.Fprintf(w, "Size\t%s\n", m.ContentSize.String)
	fmt.Fprintf(w, "Updated\t%s\n", humanTime(m.UpdatedAt))
	fmt.Fprintf(w, "Created\t%s\n", humanTime(m.CreatedAt))
	fmt.Fprintf(w, "ID\t%s\n", m.Id)
	w.Flush()
}

// humanTime reformats an ISO 8601 timestamp as an absolute date plus a relative
// age, e.g. "Nov 7, 2023 03:03 UTC (2 years ago)". Unparsable input is returned
// unchanged.
func humanTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%s (%s)", t.UTC().Format("Jan 2, 2006 15:04 MST"), humanize.Time(t))
}
