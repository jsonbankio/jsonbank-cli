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

	mode := modePublic
	var pub, prv string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("How would you like to log in?").
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
		return err
	}

	keys := app.Keys{Public: pub}
	if mode == modeBoth {
		keys.Private = prv
	}

	a.Config.Keys = keys
	if err := a.Save(); err != nil {
		return err
	}

	fmt.Println("Saved. You are now logged in.")
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

// required is a huh validator that rejects blank input.
func required(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("this field is required")
	}
	return nil
}
