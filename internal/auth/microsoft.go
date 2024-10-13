package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	util "codeberg.org/telecter/cmd-launcher/internal"
	"github.com/pkg/browser"
)

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
type xboxAuthReqProps struct {
	AuthMethod string `json:"AuthMethod"`
	SiteName   string `json:"SiteName"`
	RpsTicket  string `json:"RpsTicket"`
}
type xboxAuthRequest struct {
	Properties   xboxAuthReqProps `json:"Properties"`
	TokenType    string           `json:"TokenType"`
	RelyingParty string           `json:"RelyingParty"`
}
type xboxXSTSReqProps struct {
	SandboxID  string   `json:"SandboxId"`
	UserTokens []string `json:"UserTokens"`
}
type xboxXSTSRequest struct {
	Properties   xboxXSTSReqProps `json:"Properties"`
	RelyingParty string           `json:"RelyingParty"`
	TokenType    string           `json:"TokenType"`
}
type xboxXSTSResponse struct {
	Token string `json:"Token"`
}
type mcAuthTokenRequest struct {
	IdentityToken string `json:"identityToken"`
}
type mcAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}
type mcProfileResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}
type MinecraftLoginData struct {
	Token    string
	Refresh  string
	UUID     string
	Username string
}

const clientID string = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
const scope string = "XboxLive.SignIn,offline_access"
const redirectURL string = "http://localhost:8000/signin"

func getMsaAuthCode() (string, error) {
	params := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURL},
		"scope":         {scope},
		"response_mode": {"query"},
	}
	url, _ := url.Parse("https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize")
	url.RawQuery = params.Encode()

	err := browser.OpenURL(url.String())
	if err != nil {
		return "", fmt.Errorf("couldn't open browser: %s", err)
	}

	var code string

	server := &http.Server{Addr: ":8000", Handler: nil}
	http.HandleFunc("/signin", func(w http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()
		code = params.Get("code")
		fmt.Fprintf(w, "Response recorded. You can close this tab.")
		// what the hell is this? TODO: learn how to actually use go
		go func() {
			time.Sleep(10000)
			server.Shutdown(context.TODO())
		}()
	})
	server.ListenAndServe()
	return code, nil
}
func getMsaAuthToken(code string, refresh bool) (string, string, error) {
	data := msaTokenResponse{}
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
	resp, err := http.Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err := util.CheckResponse(resp, err); err != nil {
		return data.AccessToken, data.RefreshToken, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	json.Unmarshal(body, &data)
	return data.AccessToken, data.RefreshToken, nil
}
func getXboxAuthData(msaAuthToken string) (string, string, error) {
	data := xboxAuthRequest{
		Properties: xboxAuthReqProps{
			AuthMethod: "RPS",
			SiteName:   "user.auth.xboxlive.com",
			RpsTicket:  "d=" + msaAuthToken,
		},
		TokenType:    "JWT",
		RelyingParty: "http://auth.xboxlive.com",
	}
	req, _ := json.Marshal(data)
	resp, err := http.Post("https://user.auth.xboxlive.com/user/authenticate", "application/json", strings.NewReader(string(req)))
	if err := util.CheckResponse(resp, err); err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	respData := xboxAuthResponse{}
	err = json.Unmarshal(body, &respData)
	return respData.Token, respData.DisplayClaims.Xui[0].Uhs, err
}
func getXSTSToken(xblToken string) (string, error) {
	data := xboxXSTSRequest{
		Properties: xboxXSTSReqProps{
			SandboxID:  "RETAIL",
			UserTokens: []string{xblToken},
		},
		RelyingParty: "rp://api.minecraftservices.com/",
		TokenType:    "JWT",
	}
	req, _ := json.Marshal(data)
	resp, err := http.Post("https://xsts.auth.xboxlive.com/xsts/authorize", "application/json", strings.NewReader(string(req)))
	if err := util.CheckResponse(resp, err); err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	respData := xboxXSTSResponse{}
	json.Unmarshal(body, &respData)
	return respData.Token, nil
}

func getMinecraftAuthToken(xstsToken string, userhash string) (string, error) {
	var token string
	data := mcAuthTokenRequest{
		IdentityToken: fmt.Sprintf("XBL3.0 x=%s;%s", userhash, xstsToken),
	}
	req, _ := json.Marshal(data)
	resp, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox", "application/json", strings.NewReader(string(req)))
	if err := util.CheckResponse(resp, err); err != nil {
		return token, err
	}
	respData := mcAuthTokenResponse{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respData)
	return respData.AccessToken, nil
}
func getMinecraftProfile(jwtToken string) (string, string, error) {
	var username string
	var uuid string
	req, _ := http.NewRequest("GET", "https://api.minecraftservices.com/minecraft/profile", nil)
	req.Header.Add("Authorization", "Bearer "+jwtToken)
	resp, err := http.DefaultClient.Do(req)
	if err := util.CheckResponse(resp, err); err != nil {
		return username, uuid, err
	}
	respData := mcProfileResponse{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respData)
	return respData.Name, respData.ID, nil
}

func LoginWithMicrosoft(refreshToken string) (MinecraftLoginData, error) {
	var loginData MinecraftLoginData
	var code string
	var msaToken string
	var err error

	if refreshToken == "" {
		code, err = getMsaAuthCode()
		if err != nil {
			return loginData, fmt.Errorf("failed to retrieve Microsoft authentication code (%s)", err)
		}
		msaToken, refreshToken, err = getMsaAuthToken(code, false)
	} else {
		msaToken, refreshToken, err = getMsaAuthToken(refreshToken, true)
	}
	if err != nil {
		return loginData, fmt.Errorf("failed to retrieve Microsoft authentication token (%s)", err)
	}
	loginData.Refresh = refreshToken
	token, userhash, err := getXboxAuthData(msaToken)
	if err != nil {
		return loginData, fmt.Errorf("failed to authenticate with Xbox (%s)", err)
	}
	xstsToken, err := getXSTSToken(token)
	if err != nil {
		return loginData, fmt.Errorf("failed to authenticate with Xbox (%s)", err)
	}
	authToken, err := getMinecraftAuthToken(xstsToken, userhash)
	loginData.Token = authToken
	if err != nil {
		return loginData, fmt.Errorf("couldn't get Minecraft authentication token (%s)", err)
	}
	username, uuid, err := getMinecraftProfile(authToken)
	loginData.Username = username
	loginData.UUID = uuid
	if err != nil {
		return loginData, fmt.Errorf("couldn't get Minecraft profile (%s)", err)
	}
	return loginData, nil
}
