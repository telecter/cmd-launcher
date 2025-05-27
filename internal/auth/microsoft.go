package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/internal/network"
)

var (
	clientID    = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
	scope       = "XboxLive.SignIn,offline_access"
	redirectURL = "http://localhost:8000/signin"
)

func getMSACodeInteractive() (string, error) {
	query := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURL},
		"scope":         {scope},
		"response_mode": {"query"},
	}
	url, _ := url.Parse("https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize")
	url.RawQuery = query.Encode()

	if err := browser.OpenURL(url.String()); err != nil {
		return "", fmt.Errorf("open browser: %w", err)
	}

	var code string
	var err error
	server := &http.Server{Addr: ":8000", Handler: nil}
	http.HandleFunc("/signin", func(w http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()

		if params.Get("error") != "" {
			fmt.Fprintf(w, "An error occurred during authentication: %s", params.Get("error_description"))
			err = fmt.Errorf("get MSA code interactively: %s", params.Get("error_description"))
		} else {
			fmt.Fprintf(w, "Response recorded. You can close this tab and return to the launcher.")
		}
		code = params.Get("code")
		go server.Shutdown(context.Background())
	})
	server.ListenAndServe()
	if err != nil {
		return "", err
	}
	return code, nil
}

func authenticateMSA(code string, refresh bool) (msaAuthStore, error) {
	type response struct {
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
	params := url.Values{
		"client_id":    {clientID},
		"scope":        {scope},
		"redirect_uri": {redirectURL},
	}
	grant := "authorization_code"
	if refresh {
		grant = "refresh_token"
	}
	params.Add("grant_type", grant)
	param := "code"
	if refresh {
		param = "refresh_token"
	}
	params.Add(param, code)

	var data response
	resp, err := http.Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		return msaAuthStore{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)

	if err := network.CheckResponse(resp); err != nil {
		if data.Error != "" && data.ErrorDescription != "" {
			return msaAuthStore{}, errors.New(data.ErrorDescription)
		}
		return msaAuthStore{}, err
	}
	return msaAuthStore{
		AccessToken:  data.AccessToken,
		Expires:      time.Now().Add(time.Second * time.Duration(data.ExpiresIn)),
		RefreshToken: data.RefreshToken,
	}, nil
}

func authenticateXbox(msaAuthToken string) (xblAuthStore, error) {
	type response struct {
		Token         string `json:"Token"`
		DisplayClaims struct {
			Xui []struct {
				Uhs string `string:"uhs"`
			} `json:"xui"`
		} `json:"DisplayClaims"`
		IssueInstant time.Time `json:"IssueInstant"`
		NotAfter     time.Time `json:"NotAfter"`
	}
	type authProperties struct {
		AuthMethod string `json:"AuthMethod"`
		SiteName   string `json:"SiteName"`
		RpsTicket  string `json:"RpsTicket"`
	}
	type request struct {
		Properties   authProperties `json:"Properties"`
		TokenType    string         `json:"TokenType"`
		RelyingParty string         `json:"RelyingParty"`
	}
	req, _ := json.Marshal(
		request{
			Properties: authProperties{
				AuthMethod: "RPS",
				SiteName:   "user.auth.xboxlive.com",
				RpsTicket:  "d=" + msaAuthToken,
			},
			TokenType:    "JWT",
			RelyingParty: "http://auth.xboxlive.com",
		})
	resp, err := http.Post("https://user.auth.xboxlive.com/user/authenticate", "application/json", strings.NewReader(string(req)))
	if err != nil {
		return xblAuthStore{}, err
	}
	if err := network.CheckResponse(resp); err != nil {
		return xblAuthStore{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data response
	json.Unmarshal(body, &data)
	return xblAuthStore{
		Userhash: data.DisplayClaims.Xui[0].Uhs,
		Token:    data.Token,
		Expires:  data.NotAfter,
	}, nil
}

func authenticateXSTS(xblToken string) (xstsAuthStore, error) {
	type response struct {
		Token        string    `json:"Token"`
		IssueInstant time.Time `json:"IssueInstant"`
		NotAfter     time.Time `json:"NotAfter"`
		// error response
		XErr int `json:"XErr"`
	}
	type authProperties struct {
		SandboxID  string   `json:"SandboxId"`
		UserTokens []string `json:"UserTokens"`
	}
	type request struct {
		Properties   authProperties `json:"Properties"`
		RelyingParty string         `json:"RelyingParty"`
		TokenType    string         `json:"TokenType"`
	}

	req, _ := json.Marshal(request{
		Properties: authProperties{
			SandboxID:  "RETAIL",
			UserTokens: []string{xblToken},
		},
		RelyingParty: "rp://api.minecraftservices.com/",
		TokenType:    "JWT",
	})
	resp, err := http.Post("https://xsts.auth.xboxlive.com/xsts/authorize", "application/json", strings.NewReader(string(req)))
	if err != nil {
		return xstsAuthStore{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data response
	json.Unmarshal(body, &data)

	if err := network.CheckResponse(resp); err != nil {
		if data.XErr != 0 {
			return xstsAuthStore{}, fmt.Errorf("XSTS error %d", data.XErr)
		}
		return xstsAuthStore{}, err
	}
	return xstsAuthStore{
		Token:   data.Token,
		Expires: data.NotAfter,
	}, nil
}

func authenticateMinecraft(xstsToken string, userhash string) (minecraftAuthStore, error) {
	type request struct {
		IdentityToken string `json:"identityToken"`
	}
	type response struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	req, _ := json.Marshal(request{
		IdentityToken: fmt.Sprintf("XBL3.0 x=%s;%s", userhash, xstsToken),
	})
	resp, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox", "application/json", strings.NewReader(string(req)))
	if err != nil {
		return minecraftAuthStore{}, err
	}
	if err := network.CheckResponse(resp); err != nil {
		return minecraftAuthStore{}, err
	}
	var data response
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	return minecraftAuthStore{
		AccessToken: data.AccessToken,
		Expires:     time.Now().Add(time.Second * time.Duration(data.ExpiresIn)),
	}, nil
}

type minecraftProfile struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	// error response
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

func fetchMinecraftProfile(jwtToken string) (minecraftProfile, error) {
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Add("Authorization", "Bearer "+jwtToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return minecraftProfile{}, err
	}
	var data minecraftProfile
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	if err := network.CheckResponse(resp); err != nil {
		if data.Error != "" && data.ErrorMessage != "" {
			return minecraftProfile{}, fmt.Errorf("got error %s: %s", data.Error, data.ErrorMessage)
		}
		return minecraftProfile{}, err
	}
	return data, nil
}

type LoginSession struct {
	Token    string
	UUID     string
	Username string
}

func LoginWithMicrosoft() (LoginSession, error) {
	var err error

	store, _ := GetStore()
	now := time.Now()
	if store.MSA.AccessToken == "" || store.MSA.Expires.Before(now) {
		if store.MSA.RefreshToken != "" {
			store.MSA, err = authenticateMSA(store.MSA.RefreshToken, true)
			if err != nil {
				return LoginSession{}, fmt.Errorf("re-authenticate with MSA: %w", err)
			}
		} else {
			log.Println("No refresh token found, opening browser for authentication...")
			code, err := getMSACodeInteractive()
			if err != nil {
				return LoginSession{}, fmt.Errorf("retrieve Microsoft authentication code: %w", err)
			}
			store.MSA, err = authenticateMSA(code, false)
			if err != nil {
				return LoginSession{}, fmt.Errorf("authenticate with MSA: %w", err)
			}
		}
	}

	if store.XBL.Token == "" || store.XBL.Userhash == "" || store.XBL.Expires.Before(now) {
		store.XBL, err = authenticateXbox(store.MSA.AccessToken)
		if err != nil {
			return LoginSession{}, fmt.Errorf("authenticate with XBL: %w", err)
		}
	}

	if store.XSTS.Token == "" || store.XSTS.Expires.Before(now) {
		store.XSTS, err = authenticateXSTS(store.XBL.Token)
		if err != nil {
			return LoginSession{}, fmt.Errorf("authenticate with XSTS: %w", err)
		}
	}

	if store.Minecraft.AccessToken == "" || store.Minecraft.Expires.Before(now) {
		store.Minecraft, err = authenticateMinecraft(store.XSTS.Token, store.XBL.Userhash)
		if err != nil {
			return LoginSession{}, fmt.Errorf("authenticate with Minecraft: %w", err)
		}
	}
	profile, err := fetchMinecraftProfile(store.Minecraft.AccessToken)
	if err != nil {
		return LoginSession{}, fmt.Errorf("get Minecraft profile: %w", err)
	}

	if err := SetStore(store); err != nil {
		return LoginSession{}, fmt.Errorf("set account store: %w", err)
	}
	return LoginSession{
		Username: profile.Name,
		UUID:     profile.ID,
		Token:    store.Minecraft.AccessToken,
	}, nil
}
