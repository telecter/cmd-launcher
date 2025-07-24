package cmd

import (
	"fmt"
	"net/url"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

const (
	clientID    = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
	redirectURI = "http://localhost:8000/signin"
)

func init() {
	auth.ClientID = clientID
	auth.RedirectURI, _ = url.Parse(redirectURI)
}

type Login struct {
	NoBrowser bool `help:"${login_arg_nobrowser}"`
}

func (c *Login) Run(ctx *kong.Context) error {
	var session auth.Session

	session, err := auth.Authenticate()
	if err != nil {
		if c.NoBrowser {
			cli.Info(cli.Translate("login.code.fetching"))
			resp, err := auth.FetchDeviceCode()
			if err != nil {
				return fmt.Errorf("fetch device code: %w", err)
			}
			cli.Info(cli.Translate("login.code"), color.BlueString(resp.UserCode), color.BlueString(resp.VerificationURI))
			session, err = auth.AuthenticateWithCode(resp)
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		} else {
			cli.Info(cli.Translate("login.browser"))
			url := auth.AuthCodeURL()
			cli.Info(cli.Translate("login.url"), url.String())

			browser.OpenURL(url.String())
			var err error
			session, err = auth.AuthenticateWithRedirect(cli.Translate("login.redirect"), cli.Translate("login.redirectfail"))
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		}
	}
	cli.Success(cli.Translate("login.complete"), color.New(color.Bold).Sprint(session.Username))
	return nil
}

type Logout struct{}

func (c *Logout) Run(ctx *kong.Context) error {
	auth.Store.Clear()
	cli.Info(cli.Translate("logout.complete"))
	return nil
}

type Auth struct {
	Login  Login  `cmd:"" help:"${login}"`
	Logout Logout `cmd:"" help:"${logout}"`
}
