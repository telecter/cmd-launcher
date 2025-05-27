package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/telecter/cmd-launcher/internal/auth"
)

type Login struct{}

func (c *Login) Run(ctx *kong.Context) error {
	if auth.IsLoggedIn() {
		return fmt.Errorf("already logged in")
	}
	_, err := auth.LoginWithMicrosoft()
	if err != nil {
		return err
	}
	fmt.Println("Logged in")
	return nil
}

type Logout struct{}

func (c *Logout) Run(ctx *kong.Context) error {
	if err := auth.Logout(); err != nil {
		return err
	}
	fmt.Println("Logged out")
	return nil
}

type Auth struct {
	Login  Login  `cmd:"" name:"login" help:"Log in to an account"`
	Logout Logout `cmd:"" name:"logout" help:"Log out of an account"`
}
