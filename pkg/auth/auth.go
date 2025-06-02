// Package auth provides functions related to game authentication.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/telecter/cmd-launcher/internal/network"
)

var (
	// Client ID of the launcher. You probably should not use this, as it will make it difficult to later set up your own redirect URL.
	ClientID = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
	// Redirect URL of the launcher. If you have your own client, you will need your own redirect URL.
	RedirectURL = "http://localhost:8000/signin"
	scope       = "XboxLive.SignIn,offline_access"
	scopeSpaced = "XboxLive.signin offline_access"
)

// All required auth information to launch Minecraft.
type Session struct {
	UUID        string
	Username    string
	AccessToken string
}

type msa struct{}

var MSA msa

type msaResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	IDToken      string `json:"id_token"`
	// error response
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (msa) authenticate(payload url.Values) (msaResponse, error) {
	var data msaResponse
	resp, err := http.Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", "application/x-www-form-urlencoded", strings.NewReader(payload.Encode()))
	if err != nil {
		return msaResponse{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	return data, nil
}

type xbl struct{}

var XBL xbl

type xblResponse struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		Xui []struct {
			Uhs string `string:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
	IssueInstant time.Time `json:"IssueInstant"`
	NotAfter     time.Time `json:"NotAfter"`
}

func (xbl) authenticate(msaAccessToken string) (xblResponse, error) {
	type properties struct {
		AuthMethod string `json:"AuthMethod"`
		SiteName   string `json:"SiteName"`
		RpsTicket  string `json:"RpsTicket"`
	}
	type request struct {
		Properties   properties `json:"Properties"`
		TokenType    string     `json:"TokenType"`
		RelyingParty string     `json:"RelyingParty"`
	}
	req, _ := json.Marshal(
		request{
			Properties: properties{
				AuthMethod: "RPS",
				SiteName:   "user.auth.xboxlive.com",
				RpsTicket:  "d=" + msaAccessToken,
			},
			TokenType:    "JWT",
			RelyingParty: "http://auth.xboxlive.com",
		})
	resp, err := http.Post("https://user.auth.xboxlive.com/user/authenticate", "application/json", strings.NewReader(string(req)))
	if err != nil {
		return xblResponse{}, err
	}
	if err := network.CheckResponse(resp); err != nil {
		return xblResponse{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data xblResponse
	json.Unmarshal(body, &data)
	return data, nil
}

type xsts struct{}

var XSTS xsts

type xstsResponse struct {
	Token        string    `json:"Token"`
	IssueInstant time.Time `json:"IssueInstant"`
	NotAfter     time.Time `json:"NotAfter"`
	// error response
	XErr int `json:"XErr"`
}

func (xsts) authenticate(xblToken string) (xstsResponse, error) {
	type properties struct {
		SandboxID  string   `json:"SandboxId"`
		UserTokens []string `json:"UserTokens"`
	}
	type request struct {
		Properties   properties `json:"Properties"`
		RelyingParty string     `json:"RelyingParty"`
		TokenType    string     `json:"TokenType"`
	}

	req, _ := json.Marshal(request{
		Properties: properties{
			SandboxID:  "RETAIL",
			UserTokens: []string{xblToken},
		},
		RelyingParty: "rp://api.minecraftservices.com/",
		TokenType:    "JWT",
	})
	resp, err := http.Post("https://xsts.auth.xboxlive.com/xsts/authorize", "application/json", strings.NewReader(string(req)))
	if err != nil {
		return xstsResponse{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data xstsResponse
	json.Unmarshal(body, &data)

	if err := network.CheckResponse(resp); err != nil {
		if data.XErr != 0 {
			return xstsResponse{}, fmt.Errorf("got error %d", data.XErr)
		}
		return xstsResponse{}, err
	}
	return data, nil
}

type minecraft struct{}

var Minecraft minecraft

type minecraftResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
type minecraftProfile struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	// error response
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

func (minecraft) authenticate(xstsToken string, userhash string) (minecraftResponse, minecraftProfile, error) {
	type request struct {
		IdentityToken string `json:"identityToken"`
	}

	reqBody, _ := json.Marshal(request{
		IdentityToken: fmt.Sprintf("XBL3.0 x=%s;%s", userhash, xstsToken),
	})
	resp, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox", "application/json", strings.NewReader(string(reqBody)))
	if err != nil {
		return minecraftResponse{}, minecraftProfile{}, err
	}
	if err := network.CheckResponse(resp); err != nil {
		return minecraftResponse{}, minecraftProfile{}, err
	}
	var data minecraftResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)

	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Add("Authorization", "Bearer "+data.AccessToken)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return minecraftResponse{}, minecraftProfile{}, err
	}
	var profile minecraftProfile
	body, _ = io.ReadAll(resp.Body)
	json.Unmarshal(body, &profile)
	if err := network.CheckResponse(resp); err != nil {
		if profile.Error != "" && profile.ErrorMessage != "" {
			return minecraftResponse{}, minecraftProfile{}, errors.New(profile.Error)
		}
		return minecraftResponse{}, minecraftProfile{}, err
	}
	return data, profile, nil
}
func Authenticate() (Session, error) {
	if Store.MSA.RefreshToken == "" {
		return Session{}, fmt.Errorf("no account found")
	}
	if !Store.MSA.isValid() {
		if err := Store.MSA.refresh(); err != nil {
			return Session{}, fmt.Errorf("authenticate with MSA: %w", err)
		}
	}
	if !Store.XBL.isValid() {
		if err := Store.XBL.refresh(); err != nil {
			return Session{}, fmt.Errorf("authenticate with XBL: %w", err)
		}
	}
	if !Store.XSTS.isValid() {
		if err := Store.XSTS.refresh(); err != nil {
			return Session{}, fmt.Errorf("authenticate with XSTS: %w", err)
		}
	}
	if !Store.Minecraft.isValid() {
		if err := Store.Minecraft.refresh(); err != nil {
			return Session{}, fmt.Errorf("authenticate with Minecraft: %w", err)
		}
	}
	if err := Store.WriteToCache(); err != nil {
		return Session{}, fmt.Errorf("write auth store: %w", err)
	}
	return Session{
		Username:    Store.Minecraft.Username,
		UUID:        Store.Minecraft.UUID,
		AccessToken: Store.Minecraft.AccessToken,
	}, nil
}

// Return the URL for the user to navigate to for the OAuth2 Code flow.
func GetBrowserAuthURL(clientID string, redirectURL string) *url.URL {
	query := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURL},
		"scope":         {scope},
		"response_mode": {"query"},
	}
	loc, _ := url.Parse("https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize")
	loc.RawQuery = query.Encode()
	return loc
}

// Authenticate using the OAuth2 Code flow.
//
// This function blocks until a response has been recieved on the local authentication server.
func AuthenticateWithRedirect(clientID string, redirectURL string) (Session, error) {
	var code string
	var err error
	server := &http.Server{Addr: ":8000", Handler: nil}
	http.HandleFunc("/signin", func(w http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()

		if params.Get("error") != "" {
			fmt.Fprintf(w, "Failed to log in. An error occurred during authentication: %s", params.Get("error_description"))
			err = fmt.Errorf("get MSA code interactively: %s", params.Get("error_description"))
		} else {
			fmt.Fprintf(w, "Logged in! You can close this window.")
		}
		code = params.Get("code")
		go server.Shutdown(context.Background())
	})
	server.ListenAndServe()
	if err != nil {
		return Session{}, err
	}

	resp, err := MSA.authenticate(url.Values{
		"client_id":    {clientID},
		"scope":        {scope},
		"redirect_uri": {redirectURL},
		"grant_type":   {"authorization_code"},
		"code":         {code},
	})
	if err != nil {
		return Session{}, fmt.Errorf("authenticate with MSA: %w", err)
	}
	Store.MSA.write(resp)

	return Authenticate()
}

type deviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

// Fetch a device code that can be used to authenticate the user.
// Used in the OAuth2 Device Code flow.
func FetchDeviceCode(clientID string) (deviceCodeResponse, error) {
	params := url.Values{
		"client_id": {clientID},
		"scope":     {scopeSpaced},
	}
	resp, err := http.Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/devicecode", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return deviceCodeResponse{}, err
	}
	defer resp.Body.Close()
	if err := network.CheckResponse(resp); err != nil {
		return deviceCodeResponse{}, err
	}
	var data deviceCodeResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)

	return data, nil
}

// Authenticate with a device code.
//
// This function blocks until the user has been authenticated, or another error has occurred.
func AuthenticateWithCode(clientID string, codeResp deviceCodeResponse) (Session, error) {
	for {
		resp, err := MSA.authenticate(url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"client_id":   {clientID},
			"device_code": {codeResp.DeviceCode},
		})
		if err != nil {
			return Session{}, fmt.Errorf("authenticate with MSA: %w", err)
		}
		if resp.Error == "authorization_pending" {
			time.Sleep(time.Second * time.Duration(codeResp.Interval))
			continue
		} else if resp.Error == "" {
			Store.MSA.write(resp)
			break
		} else {
			return Session{}, errors.New(resp.Error)
		}
	}
	return Authenticate()
}
