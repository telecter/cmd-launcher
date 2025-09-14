package cmd

import (
	"fmt"
	"net/url"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/internal/cli/output"
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

// LoginCmd authenticates and logs into an account.
type LoginCmd struct {
	NoBrowser bool `help:"${login_arg_nobrowser}"`
}

func (c *LoginCmd) Run(ctx *kong.Context) error {
	var session auth.Session

	session, err := auth.Authenticate()
	if err != nil {
		if c.NoBrowser {
			output.Info(output.Translate("login.code.fetching"))
			resp, err := auth.FetchDeviceCode()
			if err != nil {
				return fmt.Errorf("fetch device code: %w", err)
			}
			output.Info(output.Translate("login.code"), color.BlueString(resp.UserCode), color.BlueString(resp.VerificationURI))
			session, err = auth.AuthenticateWithCode(resp)
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		} else {
			output.Info(output.Translate("login.browser"))
			url := auth.AuthCodeURL()
			output.Info(output.Translate("login.url"), url.String())

			browser.OpenURL(url.String())
			var err error
			session, err = auth.AuthenticateWithRedirect(output.Translate("login.redirect"), output.Translate("login.redirectfail"))
			if err != nil {
				return fmt.Errorf("add account: %w", err)
			}
		}
	}
	output.Success(output.Translate("login.complete"), color.New(color.Bold).Sprint(session.Username))
	return nil
}

// LogoutCmd logs out of the current account.
type LogoutCmd struct{}

func (c *LogoutCmd) Run(ctx *kong.Context) error {
	auth.Store.Clear()
	output.Info(output.Translate("logout.complete"))
	return nil
}

// AuthCmd enables management of an account.
type AuthCmd struct {
	Login  LoginCmd  `cmd:"" help:"${login}"`
	Logout LogoutCmd `cmd:"" help:"${logout}"`
}
