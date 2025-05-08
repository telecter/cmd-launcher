package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/browser"
	"github.com/telecter/cmd-launcher/internal/network"
)

var (
	clientID    = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
	scope       = "XboxLive.SignIn,offline_access"
	redirectURL = "http://localhost:8000/signin"
)

func fetchMSACode() (string, error) {
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
		return "", fmt.Errorf("open browser: %w", err)
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

type msaAuthResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func authenticateMSA(code string, refresh bool) (msaAuthResult, error) {
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

	var result msaAuthResult
	resp, err := http.Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err := network.CheckResponse(resp, err); err != nil {
		return msaAuthResult{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	return result, nil
}

type xboxAuthData struct {
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

func authenticateXbox(msaAuthToken string) (xboxAuthData, error) {
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
		return xboxAuthData{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var data xboxAuthData
	json.Unmarshal(body, &data)
	return data, nil
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
type xstsAuthData struct {
	Token string `json:"Token"`
}

func authenticateXSTS(xblToken string) (xstsAuthData, error) {
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
		return xstsAuthData{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var data xstsAuthData
	json.Unmarshal(body, &data)
	return data, nil
}

type minecraftAuthRequest struct {
	IdentityToken string `json:"identityToken"`
}
type minecraftAuthData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
}

func authenticateMinecraft(xstsToken string, userhash string) (minecraftAuthData, error) {
	req, _ := json.Marshal(minecraftAuthRequest{
		IdentityToken: fmt.Sprintf("XBL3.0 x=%s;%s", userhash, xstsToken),
	})
	resp, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox", "application/json", strings.NewReader(string(req)))
	if err := network.CheckResponse(resp, err); err != nil {
		return minecraftAuthData{}, err
	}
	var data minecraftAuthData
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	return data, nil
}

type minecraftProfile struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func fetchMinecraftProfile(jwtToken string) (minecraftProfile, error) {
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Add("Authorization", "Bearer "+jwtToken)
	resp, err := http.DefaultClient.Do(req)
	if err := network.CheckResponse(resp, err); err != nil {
		return minecraftProfile{}, err
	}
	var data minecraftProfile
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &data)
	return data, nil
}

type MinecraftLoginData struct {
	Token    string
	UUID     string
	Username string
}

func LoginWithMicrosoft() (MinecraftLoginData, error) {
	refreshToken := GetRefreshToken()
	var msaAuthResult msaAuthResult
	if refreshToken == "" {
		log.Println("No refresh token found, opening browser for authentication...")
		code, err := fetchMSACode()
		if err != nil {
			return MinecraftLoginData{}, fmt.Errorf("retrieve Microsoft authentication code: %w", err)
		}
		msaAuthResult, err = authenticateMSA(code, false)
		if err != nil {
			return MinecraftLoginData{}, fmt.Errorf("authenticate with MSA: %w", err)
		}
	} else {
		var err error
		msaAuthResult, err = authenticateMSA(refreshToken, true)
		if err != nil {
			return MinecraftLoginData{}, fmt.Errorf("re-authenticate with MSA: %w", err)
		}
	}
	xboxAuthData, err := authenticateXbox(msaAuthResult.AccessToken)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("authenticate with Xbox: %w", err)
	}
	xstsAuthData, err := authenticateXSTS(xboxAuthData.Token)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("authenticate with Xbox: %w", err)
	}

	minecraftAuthData, err := authenticateMinecraft(xstsAuthData.Token, xboxAuthData.DisplayClaims.Xui[0].Uhs)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("authenticate with Minecraft: %w", err)
	}
	profile, err := fetchMinecraftProfile(minecraftAuthData.AccessToken)
	if err != nil {
		return MinecraftLoginData{}, fmt.Errorf("get Minecraft profile: %w", err)
	}

	if err := SetRefreshToken(msaAuthResult.RefreshToken); err != nil {
		return MinecraftLoginData{}, fmt.Errorf("set refresh token: %w", err)
	}
	log.Println("Authenticated successfully")
	return MinecraftLoginData{
		Username: profile.Name,
		UUID:     profile.ID,
		Token:    minecraftAuthData.AccessToken,
	}, nil
}
