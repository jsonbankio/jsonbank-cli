package cmd

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/jsonbankio/jsonbank-cli/internal/app"
)

var authSwitchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch the active account",
	RunE:  runAuthSwitch,
}

var authAccountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage saved JSONBank accounts",
}

var authAccountsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved accounts",
	RunE:  runAccountsList,
}

var authAccountsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add an account without making it active",
	RunE:  runAccountsAdd,
}

var authAccountsRemoveCmd = &cobra.Command{
	Use:   "remove <username>",
	Short: "Remove a saved account",
	Args:  cobra.ExactArgs(1),
	RunE:  runAccountsRemove,
}

func init() {
	authAccountsCmd.AddCommand(authAccountsListCmd, authAccountsAddCmd, authAccountsRemoveCmd)
	authCmd.AddCommand(authSwitchCmd, authAccountsCmd)
}

const noAccountsHint = "No accounts yet. Add one with: jsb auth accounts add"

func runAccountsList(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	if len(a.Config.Accounts) == 0 {
		fmt.Println(noAccountsHint)
		return nil
	}

	active := a.Config.ActiveUsername()
	for _, username := range sortedUsernames(a.Config.Accounts) {
		marker := "  "
		if username == active {
			marker = "* "
		}
		fmt.Printf("%s%s\n", marker, username)
	}
	return nil
}

func runAccountsAdd(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	username, err := addAccount(a)
	if err != nil {
		return err
	}

	if err := a.Save(); err != nil {
		return err
	}

	fmt.Printf("Added %s (not active)\n", username)
	return nil
}

func runAccountsRemove(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	username := args[0]
	if !a.Config.RemoveAccount(username) {
		return fmt.Errorf("no account named %q", username)
	}

	if err := a.Save(); err != nil {
		return err
	}

	fmt.Printf("Removed %s\n", username)
	return nil
}

func runAuthSwitch(cmd *cobra.Command, args []string) error {
	a, err := app.Init()
	if err != nil {
		return err
	}

	if len(a.Config.Accounts) == 0 {
		fmt.Println(noAccountsHint)
		return nil
	}

	choice, err := selectAccount(a, "Switch to which account?")
	if err != nil {
		return err
	}

	a.Config.Activate(choice)
	if err := a.Save(); err != nil {
		return err
	}

	fmt.Printf("Switched to %s\n", choice)
	return nil
}

// selectAccount shows a select over the saved accounts (plus any extra options)
// and returns the chosen value. The active account is pre-selected.
func selectAccount(a *app.App, title string, extra ...huh.Option[string]) (string, error) {
	options := make([]huh.Option[string], 0, len(a.Config.Accounts)+len(extra))
	for _, username := range sortedUsernames(a.Config.Accounts) {
		options = append(options, huh.NewOption(username, username))
	}
	options = append(options, extra...)

	choice := a.Config.ActiveUsername()
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Value(&choice),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return choice, nil
}

// sortedUsernames returns the account usernames in stable alphabetical order.
func sortedUsernames(accounts map[string]app.Keys) []string {
	names := make([]string, 0, len(accounts))
	for u := range accounts {
		names = append(names, u)
	}
	sort.Strings(names)
	return names
}
