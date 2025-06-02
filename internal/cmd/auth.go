package cmd

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

type Login struct {
	NoBrowser bool `name:"no-browser" help:"Use device code instead of browser for authentication"`
}

func (c *Login) Run(ctx *kong.Context) error {
	var session auth.Session

	session, err := auth.Authenticate()
	if err != nil {
		if c.NoBrowser {
			resp, err := auth.FetchDeviceCode(auth.ClientID)
			if err != nil {
				return fmt.Errorf("fetch device code: %w", err)
			}
			fmt.Printf("Use the code %s at %s to sign in\n", resp.UserCode, resp.VerificationURI)
			fmt.Println("Waiting for authentication....")
			session, err = auth.AuthenticateWithCode(auth.ClientID, resp)
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		} else {
			fmt.Println("Opening browser for authentication...")
			url := auth.GetBrowserAuthURL(auth.ClientID, auth.RedirectURL)
			if err := browser.OpenURL(url.String()); err != nil {
				return fmt.Errorf("open browser: %w", err)
			}
			var err error
			session, err = auth.AuthenticateWithRedirect(auth.ClientID, auth.RedirectURL)

			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		}
	}
	fmt.Printf("Logged in as %s\n", session.Username)
	return nil
}

type Logout struct{}

func (c *Logout) Run(ctx *kong.Context) error {
	auth.Store.Clear()
	fmt.Println("Logged out from account")
	return nil
}

type Auth struct {
	Login  Login  `cmd:"" name:"login" help:"Log in to an account"`
	Logout Logout `cmd:"" name:"logout" help:"Log out of an account"`
}
