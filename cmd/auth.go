package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	jsonbank "github.com/jsonbankio/go-sdk"
	"github.com/spf13/cobra"

	"jsb-cli/internal/app"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage JSONBank authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in by saving your JSONBank API keys",
	RunE:  runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out by removing saved keys",
	RunE:  runAuthLogout,
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the currently logged-in user",
	RunE:  runAuthWhoami,
}

func init() {
	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authWhoamiCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	in := bufio.NewReader(os.Stdin)

	fmt.Println("How would you like to log in?")
	fmt.Println("  1) Public key only")
	fmt.Println("  2) Public and private key")

	var choice string
	for {
		choice = ask(in, "Select [1-2]: ")
		if choice == "1" || choice == "2" {
			break
		}
		fmt.Println("Please enter 1 or 2.")
	}

	keys := app.Keys{Public: ask(in, "Public key: ")}
	if choice == "2" {
		keys.Private = ask(in, "Private key: ")
	}

	if keys.Public == "" {
		return fmt.Errorf("a public key is required")
	}

	a.Config.Keys = keys
	if err := a.Save(); err != nil {
		return err
	}

	fmt.Printf("Saved. You are now logged in.\nConfig: %s\n", a.ConfigPath())
	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	if a.Config.Keys == (app.Keys{}) {
		fmt.Println("You are not logged in.")
		return nil
	}

	a.Config.Keys = app.Keys{}
	if err := a.Save(); err != nil {
		return err
	}

	fmt.Println("Logged out. Keys removed.")
	return nil
}

func runAuthWhoami(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	if a.Config.Keys.Public == "" {
		return fmt.Errorf("you are not logged in — run: jsb auth login")
	}

	jsb := jsonbank.Init(jsonbank.Config{
		Host: a.Config.Host,
		Keys: jsonbank.Keys{
			Public:  a.Config.Keys.Public,
			Private: a.Config.Keys.Private,
		},
	})

	data, reqErr := jsb.Authenticate()
	if reqErr != nil {
		return fmt.Errorf("authentication failed: %s", reqErr.Message)
	}

	fmt.Printf("Logged As: %s\n", data.Username)
	return nil
}

// ask prints a prompt and reads a trimmed line from the reader.
func ask(r *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}
