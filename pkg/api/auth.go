package api

import (
	util "cmd-launcher/internal"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type MSATokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
type XboxAuthResponse struct {
	Token         string `json:"Token"`
	DisplayClaims struct {
		Xui []struct {
			Uhs string `string:"uhs"`
		} `json:"xui"`
	} `json:"DisplayClaims"`
}
type XboxAuthReqProps struct {
	AuthMethod string `json:"AuthMethod"`
	SiteName   string `json:"SiteName"`
	RpsTicket  string `json:"RpsTicket"`
}
type XboxAuthRequest struct {
	Properties   XboxAuthReqProps `json:"Properties"`
	TokenType    string           `json:"TokenType"`
	RelyingParty string           `json:"RelyingParty"`
}
type XboxXSTSReqProps struct {
	SandboxID  string   `json:"SandboxId"`
	UserTokens []string `json:"UserTokens"`
}
type XboxXSTSRequest struct {
	Properties   XboxXSTSReqProps `json:"Properties"`
	RelyingParty string           `json:"RelyingParty"`
	TokenType    string           `json:"TokenType"`
}
type XboxXSTSResponse struct {
	Token string `json:"Token"`
}
type MCAuthTokenRequest struct {
	IdentityToken string `json:"identityToken"`
}
type MCAuthTokenResponse struct {
	AccessToken string `json:"access_token"`
}
type MCProfileResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

const clientID string = "6a533aa3-afbf-45a4-91bc-8c35a37e35c7"
const scope string = "XboxLive.SignIn,offline_access"
const redirectURL string = "http://localhost:8000/signin"

func getMsaAuthCode() string {
	params := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"redirect_uri":  {redirectURL},
		"scope":         {scope},
		"response_mode": {"query"},
	}
	url, _ := url.Parse("https://login.microsoftonline.com/consumers/oauth2/v2.0/authorize")
	url.RawQuery = params.Encode()

	cmd := exec.Command("open", url.String())
	cmd.Start()

	var code string

	server := &http.Server{Addr: ":8000", Handler: nil}
	http.HandleFunc("/signin", func(w http.ResponseWriter, req *http.Request) {
		params := req.URL.Query()
		code = params.Get("code")
		fmt.Fprintf(w, "Response recorded. You can close this tab.")
		go func() {
			time.Sleep(1)
			server.Shutdown(nil)
		}()
	})
	server.ListenAndServe()
	return code
}
func getMsaAuthToken(code string, refresh bool) (string, error) {
	data := MSATokenResponse{}
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
		return data.AccessToken, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	json.Unmarshal(body, &data)
	return data.AccessToken, nil
}
func getXboxAuthData(msaAuthToken string) (string, string, error) {
	data := XboxAuthRequest{
		Properties: XboxAuthReqProps{
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
	respData := XboxAuthResponse{}
	err = json.Unmarshal(body, &respData)
	return respData.Token, respData.DisplayClaims.Xui[0].Uhs, err
}
func getXSTSToken(xblToken string) (string, error) {
	data := XboxXSTSRequest{
		Properties: XboxXSTSReqProps{
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
	respData := XboxAuthResponse{}
	json.Unmarshal(body, &respData)
	return respData.Token, nil
}

func getMinecraftAuthToken(xstsToken string, userhash string) (string, error) {
	var token string
	data := MCAuthTokenRequest{
		IdentityToken: fmt.Sprintf("XBL3.0 x=%v;%v", userhash, xstsToken),
	}
	req, _ := json.Marshal(data)
	resp, err := http.Post("https://api.minecraftservices.com/authentication/login_with_xbox", "application/json", strings.NewReader(string(req)))
	if err := util.CheckResponse(resp, err); err != nil {
		return token, err
	}
	respData := MCAuthTokenResponse{}
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
	respData := MCProfileResponse{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &respData)
	return respData.Name, respData.ID, nil
}

func GetAuthData() (string, string, string, error) {
	var authToken string
	var uuid string
	var username string

	code := getMsaAuthCode()
	msaToken, err := getMsaAuthToken(code, false)
	if err != nil {
		return authToken, uuid, username, fmt.Errorf("Failed to retrieve Microsoft authentication token (%v)", err)
	}
	token, userhash, err := getXboxAuthData(msaToken)
	if err != nil {
		return authToken, uuid, username, fmt.Errorf("Failed to authenticate with Xbox (%v)", err)
	}
	xstsToken, err := getXSTSToken(token)
	if err != nil {
		return authToken, uuid, username, fmt.Errorf("Failed to authenticate with Xbox (%v)", err)
	}
	authToken, err = getMinecraftAuthToken(xstsToken, userhash)
	if err != nil {
		return authToken, uuid, username, fmt.Errorf("Couldn't get Minecraft authentication token (%v)", err)
	}
	username, uuid, err = getMinecraftProfile(authToken)
	if err != nil {
		return authToken, uuid, username, fmt.Errorf("Couldn't get Minecraft profile (%v)", err)
	}
	return authToken, uuid, username, nil
}
