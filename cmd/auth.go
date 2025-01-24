package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/urfave/cli/v2"
)

func logout(ctx *cli.Context) error {
	if err := os.Remove(filepath.Join(ctx.String("dir"), "account.txt")); errors.Is(err, fs.ErrNotExist) {
		fmt.Println("Already logged out")
		return nil
	} else if err != nil {
		return cli.Exit(fmt.Errorf("error removing account information: %s", err), 1)
	}
	fmt.Println("Logged out")
	return nil
}
func login(ctx *cli.Context) error {
	if _, err := os.Stat(filepath.Join(ctx.String("dir"), "account.txt")); err == nil {
		return cli.Exit("Already logged in", 0)
	}
	authData, err := auth.LoginWithMicrosoft("")
	if err != nil {
		return cli.Exit(fmt.Errorf("error logging in with MSA: %s", err), 1)
	}
	if err = os.WriteFile(filepath.Join(ctx.String("dir"), "account.txt"), []byte(authData.Refresh), 0644); err != nil {
		return cli.Exit(fmt.Errorf("error writing auth data: %s", err), 1)
	}
	fmt.Println("Logged in")
	return nil
}

var Auth = &cli.Command{
	Name:  "auth",
	Usage: "Manage account authentication",
	Subcommands: []*cli.Command{
		{
			Name:   "logout",
			Usage:  "Log out from account",
			Action: logout,
		},
		{
			Name:   "login",
			Usage:  "Log in to an account",
			Action: login,
		},
	},
}
