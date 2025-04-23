package cmd

import (
	"context"
	"log"

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/urfave/cli/v3"
)

func logout(ctx context.Context, c *cli.Command) error {
	if err := auth.Logout(); err != nil {
		return cli.Exit(err, 1)
	}
	log.Println("Logged out")
	return nil
}
func login(ctx context.Context, c *cli.Command) error {
	if auth.IsLoggedIn() {
		return cli.Exit("Already logged in", 0)
	}
	_, err := auth.LoginWithMicrosoft()
	if err != nil {
		return cli.Exit(err, 1)
	}
	log.Println("Logged in")
	return nil
}

var Auth = &cli.Command{
	Name:  "auth",
	Usage: "Manage account authentication",
	Commands: []*cli.Command{
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
