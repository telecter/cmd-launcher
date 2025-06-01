package auth

import (
	"encoding/json"
	"net/url"
	"os"
	"time"

	"github.com/telecter/cmd-launcher/internal"
)

var Store AuthStore

type msaAuthStore struct {
	AccessToken  string    `json:"access_token"`
	Expires      time.Time `json:"expires"`
	RefreshToken string    `json:"refresh_token"`
}

func (store *msaAuthStore) isValid() bool {
	return store.AccessToken != "" && store.Expires.After(time.Now())
}
func (store *msaAuthStore) refresh() error {
	resp, err := MSA.authenticate(url.Values{
		"client_id":     {clientID},
		"scope":         {scope},
		"redirect_uri":  {redirectURL},
		"grant_type":    {"refresh_token"},
		"refresh_token": {Store.MSA.RefreshToken},
	})
	if err != nil {
		return err
	}
	store.write(resp)
	return nil
}
func (store *msaAuthStore) write(resp msaResponse) {
	store.AccessToken = resp.AccessToken
	store.Expires = time.Now().Add(time.Second * time.Duration(resp.ExpiresIn))
	store.RefreshToken = resp.RefreshToken
}

type xblAuthStore struct {
	Userhash string    `json:"uhs"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires"`
}

func (store *xblAuthStore) isValid() bool {
	return store.Token != "" && store.Userhash != "" && store.Expires.After(time.Now())
}
func (store *xblAuthStore) refresh() error {
	resp, err := XBL.authenticate(Store.MSA.AccessToken)
	if err != nil {
		return err
	}
	store.write(resp)
	return nil
}
func (store *xblAuthStore) write(resp xblResponse) {
	store.Userhash = resp.DisplayClaims.Xui[0].Uhs
	store.Token = resp.Token
	store.Expires = resp.NotAfter
}

type xstsAuthStore struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func (store *xstsAuthStore) isValid() bool {
	return store.Token != "" && store.Expires.After(time.Now())
}
func (store *xstsAuthStore) refresh() error {
	resp, err := XSTS.authenticate(Store.XBL.Token)
	if err != nil {
		return err
	}
	store.write(resp)
	return nil
}
func (store *xstsAuthStore) write(resp xstsResponse) {
	store.Token = resp.Token
	store.Expires = resp.NotAfter
}

type minecraftAuthStore struct {
	AccessToken string    `json:"access_token"`
	Expires     time.Time `json:"expires"`
	Username    string    `json:"name"`
	UUID        string    `json:"id"`
}

func (store *minecraftAuthStore) isValid() bool {
	return store.AccessToken != "" && store.Expires.After(time.Now())
}
func (store *minecraftAuthStore) refresh() error {
	resp, profile, err := Minecraft.authenticate(Store.XSTS.Token, Store.XBL.Userhash)
	if err != nil {
		return err
	}
	store.write(resp, profile)
	return nil
}
func (store *minecraftAuthStore) write(resp minecraftResponse, profile minecraftProfile) {
	store.AccessToken = resp.AccessToken
	store.Expires = time.Now().Add(time.Second * time.Duration(resp.ExpiresIn))
	store.Username = profile.Name
	store.UUID = profile.ID
}

type AuthStore struct {
	MSA       msaAuthStore       `json:"msa"`
	XBL       xblAuthStore       `json:"xbl"`
	XSTS      xstsAuthStore      `json:"xsts"`
	Minecraft minecraftAuthStore `json:"minecraft"`
}

func (store *AuthStore) Write() error {
	data, _ := json.MarshalIndent(store, "", "    ")
	return os.WriteFile(internal.AuthStorePath, data, 0644)
}
func (store *AuthStore) Clear() error {
	store = &AuthStore{}
	return store.Write()
}
