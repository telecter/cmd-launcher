package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/internal/network"
)

type MinecraftLoginData struct {
	Token    string
	UUID     string
	Username string
}

type msaTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type xboxAuthResponse struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		Xui []struct {
			Uhs string `string:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}
type xboxAuthProperties struct {
	AuthMethod string `json:"AuthMethod"`
	SiteName   string `json:"SiteName"`
	RpsTicket  string `json:"RpsTicket"`
}
type xboxAuthRequest struct {
	Properties struct {
		AuthMethod string `json:"AuthMethod"`
		SiteName   string `json:"SiteName"`
		RpsTicket  string `json:"RpsTicket"`
	} `json:"Properties"`
	TokenType    string `json:"TokenType"`
	RelyingParty string `json:"RelyingParty"`
}
type xstsProperties struct {
	SandboxID  string   `json:"SandboxId"`
	UserTokens []string `json:"UserTokens"`
}
type xstsRequest struct {
	Properties   xstsProperties `json:"Properties"`
	RelyingParty string         `json:"RelyingParty"`
	TokenType    string         `json:"TokenType"`
}
type xstsResponse struct {
	Token string `json:"Token"`
}
type minecraftAuthTokenRequest struct {
	IdentityToken string `json:"identityToken"`
}
type minecraftAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}
type minecraftProfileResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

const clientID string = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
const scope string = "XboxLive.SignIn,offline_access"
const redirectURL string = "http://localhost:8000/signin"

func fetchMSAOAuthCode() (string, error) {
	query := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURL},
		"scope":         {scope},
		"response_mode": {"query"},
	}
	url, _ := url.Parse("https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize")
	url.RawQuery = query.Encode()

	err := browser.OpenURL(url.String())
	if err != nil {
		return "", fmt.Errorf("couldn't open browser: %w", err)
	}

	var code string

	server := &http.Server{Addr: ":8000", Handler: nil}
	http.HandleFunc("/signin", func(w http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()
		code = params.Get("code")
		fmt.Fprintf(w, "Response recorded. You can close this tab.")

		go server.Shutdown(context.Background())
	})
	server.ListenAndServe()
	return code, nil
}

func fetchMSAToken(code string, refresh bool) (string, string, error) {
	params := url.Values{
		"client_id":    {clientID},
		"scope":        {scope},
		"redirect_uri": {redirectURL},
	}
	grantType := "authorization_code"
	if refresh {
		grantType = "refresh_token"
	}
	params.Add("grant_type", grantType)
	paramName := "code"
	if refresh {
		paramName = "refresh_token"
	}
	params.Add(paramName, code)

	var data msaTokenResponse
	resp, err := http.Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err := network.CheckResponse(resp, err); err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	return data.AccessToken, data.RefreshToken, nil
}

func fetchXboxAuthData(msaAuthToken string) (string, string, error) {
	req, _ := json.Marshal(
		xboxAuthRequest{
			Properties: xboxAuthProperties{
				AuthMethod: "RPS",
				SiteName:   "user.auth.xboxlive.com",
				RpsTicket:  "d=" + msaAuthToken,
			},
			TokenType:    "JWT",
			RelyingParty: "http://auth.xboxlive.com",
		})
	resp, err := http.Post("https://user.auth.xboxlive.com/user/authenticate", "application/json", strings.NewReader(string(req)))
	if err := network.CheckResponse(resp, err); err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data xboxAuthResponse
	err = json.Unmarshal(body, &data)
	return data.Token, data.DisplayClaims.Xui[0].Uhs, err
}

func fetchXboxXSTSToken(xblToken string) (string, error) {
	req, _ := json.Marshal(xstsRequest{
		Properties: xstsProperties{
			SandboxID:  "RETAIL",
			UserTokens: []string{xblToken},
		},
		RelyingParty: "rp://api.minecraftservices.com/",
		TokenType:    "JWT",
	})
	resp, err := http.Post("https://xsts.auth.xboxlive.com/xsts/authorize", "application/json", strings.NewReader(string(req)))
	if err := network.CheckResponse(resp, err); err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data xstsResponse
	json.Unmarshal(body, &data)
	return data.Token, nil
}

func fetchMinecraftAuthToken(xstsToken string, userhash string) (string, error) {
	req, _ := json.Marshal(minecraftAuthTokenRequest{
		IdentityToken: fmt.Sprintf("XBL3.0 x=%s;%s", userhash, xstsToken),
	})
	resp, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox", "application/json", strings.NewReader(string(req)))
	if err := network.CheckResponse(resp, err); err != nil {
		return "", err
	}
	var data minecraftAuthTokenResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	return data.AccessToken, nil
}
func fetchMinecraftProfile(jwtToken string) (string, string, error) {
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Add("Authorization", "Bearer "+jwtToken)
	resp, err := http.DefaultClient.Do(req)
	if err := network.CheckResponse(resp, err); err != nil {
		return "", "", err
	}
	var data minecraftProfileResponse
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	return data.Name, data.ID, nil
}

func LoginWithMicrosoft() (MinecraftLoginData, error) {
	var err error
	refreshToken := GetRefreshToken()
	var msaToken string
	if refreshToken == "" {
		oAuthCode, err := fetchMSAOAuthCode()
		if err != nil {
			return MinecraftLoginData{}, fmt.Errorf("failed to retrieve Microsoft authentication code: %w", err)
		}
		msaToken, refreshToken, err = fetchMSAToken(oAuthCode, false)
	} else {
		msaToken, refreshToken, err = fetchMSAToken(refreshToken, true)
	}
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("failed to retrieve Microsoft authentication token: %w", err)
	}

	xblToken, userhash, err := fetchXboxAuthData(msaToken)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("failed to authenticate with Xbox: %w", err)
	}
	xstsToken, err := fetchXboxXSTSToken(xblToken)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("failed to authenticate with Xbox: %w", err)
	}

	minecraftToken, err := fetchMinecraftAuthToken(xstsToken, userhash)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("failed to get Minecraft authentication token: %w", err)
	}
	username, uuid, err := fetchMinecraftProfile(minecraftToken)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("failed to get Minecraft profile: %w", err)
	}

	SetRefreshToken(refreshToken)
	return MinecraftLoginData{
		Username: username,
		UUID:     uuid,
		Token:    minecraftToken,
	}, nil
}
