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
func getMsaAuthToken(code string, refresh bool) string {
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
	resp, _ := http.Post("https://login.microsoftonline.com/consumers/oauth2/v2.0/token", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	body, _ := io.ReadAll(resp.Body)
	data := struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{}
	json.Unmarshal(body, &data)
	return data.AccessToken
}
func getXboxAuthData(msaAuthToken string) (string, string, error) {
	var token string
	var userhash string
	type Properties struct {
		AuthMethod string `json:"AuthMethod"`
		SiteName   string `json:"SiteName"`
		RpsTicket  string `json:"RpsTicket"`
	}
	data := struct {
		Properties   Properties `json:"Properties"`
		RelyingParty string     `json:"RelyingParty"`
		TokenType    string     `json:"TokenType"`
	}{
		Properties: Properties{
			AuthMethod: "RPS",
			SiteName:   "user.auth.xboxlive.com",
			RpsTicket:  "d=" + msaAuthToken,
		},
		RelyingParty: "http://auth.xboxlive.com",
		TokenType:    "JWT",
	}
	json, _ := json.Marshal(data)
	resp, err := http.Post("https://user.auth.xboxlive.com/user/authenticate", "application/json", strings.NewReader(string(json)))
	if err := util.CheckResponse(resp, err); err != nil {
		return token, userhash, err
	}
	return token, userhash, err
}

func GetAuthData() string {
	code := getMsaAuthCode()
	token := getMsaAuthToken(code, false)
	return token
}
