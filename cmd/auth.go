package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
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

const (
	modePublic = "public"
	modeBoth   = "both"
)

func runAuthLogin(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	username, err := chooseOrAddAccount(a)
	if err != nil {
		return err
	}

	a.Config.Activate(username)
	if err := a.Save(); err != nil {
		return err
	}

	fmt.Printf("Logged in as %s\n", username)
	return nil
}

// chooseOrAddAccount lets the user pick one of the already-saved accounts (or
// "Other" to enter new keys). With no saved accounts it goes straight to the
// add flow. Returns the username to activate.
func chooseOrAddAccount(a *app.App) (string, error) {
	if len(a.Config.Accounts) == 0 {
		return addAccount(a)
	}

	choice, err := selectAccount(a, "Log in as", huh.NewOption("Other — log in with keys", ""))
	if err != nil {
		return "", err
	}
	if _, ok := a.Config.Accounts[choice]; ok {
		return choice, nil // existing account chosen
	}
	return addAccount(a) // "Other" selected (or nothing matched)
}

// addAccount runs the key form, resolves the username from the API (which also
// validates the keys), and stores the account in the config. It does not save —
// the caller decides when to persist. Shared by `login` and `accounts add`.
func addAccount(a *app.App) (string, error) {
	keys, err := collectKeys()
	if err != nil {
		return "", err
	}

	username, err := resolveUsername(a.Config.Host, keys)
	if err != nil {
		return "", err
	}

	a.Config.Accounts[username] = keys
	return username, nil
}

// collectKeys runs the interactive form gathering a public (and optional
// private) key.
func collectKeys() (app.Keys, error) {
	mode := modePublic
	var pub, prv string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("How do you want to authenticate?").
				Options(
					huh.NewOption("Public key only", modePublic),
					huh.NewOption("Public and private key", modeBoth),
				).
				Value(&mode),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Public key").
				Value(&pub).
				Validate(required),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Private key").
				EchoMode(huh.EchoModeNone). // hide the secret entirely
				Value(&prv).
				Validate(required),
		).WithHideFunc(func() bool { return mode != modeBoth }),
	)

	if err := form.Run(); err != nil {
		return app.Keys{}, err
	}

	keys := app.Keys{Public: pub}
	if mode == modeBoth {
		keys.Private = prv
	}
	return keys, nil
}

// resolveUsername authenticates with the given keys and returns the account's
// username. It doubles as key validation.
func resolveUsername(host string, keys app.Keys) (string, error) {
	jsb := jsonbank.Init(jsonbank.Config{
		Host: host,
		Keys: jsonbank.Keys{Public: keys.Public, Private: keys.Private},
	})

	data, reqErr := jsb.Authenticate()
	if reqErr != nil {
		return "", fmt.Errorf("authentication failed: %s", reqErr.Message)
	}
	return data.Username, nil
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

	username, err := resolveUsername(a.Config.Host, a.Config.Keys)
	if err != nil {
		return err
	}

	fmt.Printf("Logged As: %s\n", username)
	return nil
}

// required is a huh validator that rejects blank input.
func required(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("this field is required")
	}
	return nil
}
