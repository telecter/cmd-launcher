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
	auth.MSA.ClientID = clientID
	auth.MSA.RedirectURI, _ = url.Parse(redirectURI)
}

type Login struct {
	NoBrowser bool `help:"${cmd_auth_nobrowser}"`
}

func (c *Login) Run(ctx *kong.Context) error {
	var session auth.Session

	session, err := auth.Authenticate()
	if err != nil {
		if c.NoBrowser {
			cli.Info(cli.Translate("auth.code.fetching"))
			resp, err := auth.MSA.FetchDeviceCode()
			if err != nil {
				return fmt.Errorf("fetch device code: %w", err)
			}
			cli.Info(cli.Translate("auth.code.display"), color.BlueString(resp.UserCode), color.BlueString(resp.VerificationURI))
			session, err = auth.AuthenticateWithCode(resp)
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		} else {
			cli.Info(cli.Translate("auth.browser.opening"))
			url := auth.MSA.AuthCodeURL()
			cli.Info(cli.Translate("auth.browser.url"), url.String())

			browser.OpenURL(url.String())
			var err error
			session, err = auth.AuthenticateWithRedirect()
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		}
	}
	cli.Success(cli.Translate("auth.complete"), color.New(color.Bold).Sprint(session.Username))
	return nil
}

type Logout struct{}

func (c *Logout) Run(ctx *kong.Context) error {
	auth.Store.Clear()
	cli.Info(cli.Translate("auth.logout"))
	return nil
}

type Auth struct {
	Login  Login  `cmd:"" help:"${cmd_login}"`
	Logout Logout `cmd:"" help:"${cmd_logout}"`
}
