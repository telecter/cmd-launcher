package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func logout(ctx *cli.Context) error {
	err := os.Remove(filepath.Join(ctx.String("dir"), "account.txt"))
	if err != nil {
		return cli.Exit(fmt.Errorf("error removing account information: %s", err), 1)
	}
	fmt.Println("Logged out")
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
	},
}
