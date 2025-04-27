package cmd

import (
	"context"
	"log"

	"github.com/telecter/cmd-launcher/internal/auth"
	"github.com/urfave/cli/v3"
)

var login = &cli.Command{
	Name:  "logout",
	Usage: "Log out from account",
	Action: func(ctx context.Context, c *cli.Command) error {
		if err := auth.Logout(); err != nil {
			return cli.Exit(err, 1)
		}
		log.Println("Logged out")
		return nil
	},
}
var logout = &cli.Command{
	Name:  "login",
	Usage: "Log in to an account",
	Action: func(ctx context.Context, c *cli.Command) error {
		if auth.IsLoggedIn() {
			return cli.Exit("already logged in", 0)
		}
		_, err := auth.LoginWithMicrosoft()
		if err != nil {
			return cli.Exit(err, 1)
		}
		log.Println("Logged in")
		return nil
	},
}

var Auth = &cli.Command{
	Name:  "auth",
	Usage: "Manage account authentication",
	Commands: []*cli.Command{
		login,
		logout,
	},
}
