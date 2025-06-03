package cmd

import (
	"fmt"
	"net/url"

	"github.com/alecthomas/kong"
	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

const clientID = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"

var redirectURL, _ = url.Parse("http://localhost:8000/signin")

func init() {
	auth.MSA.ClientID = clientID
	auth.MSA.RedirectURI = redirectURL
}

type Login struct {
	NoBrowser bool `name:"no-browser" help:"Use device code instead of browser for authentication"`
}

func (c *Login) Run(ctx *kong.Context) error {
	var session auth.Session

	session, err := auth.Authenticate()
	if err != nil {
		if c.NoBrowser {
			resp, err := auth.MSA.FetchDeviceCode()
			if err != nil {
				return fmt.Errorf("fetch device code: %w", err)
			}
			fmt.Printf("Use the code %s at %s to sign in\n", resp.UserCode, resp.VerificationURI)
			fmt.Println("Waiting for authentication....")
			session, err = auth.AuthenticateWithCode(resp)
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		} else {
			fmt.Println("Opening browser for authentication...")
			url := auth.MSA.AuthCodeURL()
			if err := browser.OpenURL(url.String()); err != nil {
				return fmt.Errorf("open browser: %w", err)
			}
			var err error
			session, err = auth.AuthenticateWithRedirect()

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
