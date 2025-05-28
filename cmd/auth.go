package cmd

import (
	"fmt"
	"log"

	"github.com/alecthomas/kong"
	"github.com/telecter/cmd-launcher/internal/auth"
)

type Login struct{}

func (c *Login) Run(ctx *kong.Context) error {
	if auth.IsLoggedIn() {
		return fmt.Errorf("already logged in")
	}
	log.Println("Opening browser for authentication...")
	session, err := auth.LoginMicrosoftInteractive()
	if err != nil {
		return fmt.Errorf("add account: %w", err)
	}
	fmt.Printf("Logged in as %s\n", session.Username)
	return nil
}

type Logout struct{}

func (c *Logout) Run(ctx *kong.Context) error {
	if err := auth.Logout(); err != nil {
		return err
	}
	fmt.Println("Logged out from account")
	return nil
}

type Auth struct {
	Login  Login  `cmd:"" name:"login" help:"Log in to an account"`
	Logout Logout `cmd:"" name:"logout" help:"Log out of an account"`
}
