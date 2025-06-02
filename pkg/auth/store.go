package auth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	env "github.com/telecter/cmd-launcher/pkg"
)

// Store is the global authentication store.
var Store AuthStore

type msaAuthStore struct {
	AccessToken  string    `json:"access_token"`
	Expires      time.Time `json:"expires"`
	RefreshToken string    `json:"refresh_token"`
}

func (store *msaAuthStore) isValid() bool {
	return store.AccessToken != "" && store.Expires.After(time.Now())
}
func (store *msaAuthStore) refresh(clientID string) error {
	resp, err := MSA.authenticate(url.Values{
		"client_id":     {clientID},
		"scope":         {scope},
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

// An AuthStore is an authentication store which stores necessary information to log in.
type AuthStore struct {
	MSA       msaAuthStore       `json:"msa"`
	XBL       xblAuthStore       `json:"xbl"`
	XSTS      xstsAuthStore      `json:"xsts"`
	Minecraft minecraftAuthStore `json:"minecraft"`
}

// WriteToCache writes the store to the file specified in env.AuthStorePath
func (store *AuthStore) WriteToCache() error {
	data, _ := json.MarshalIndent(store, "", "    ")
	return os.WriteFile(env.AuthStorePath, data, 0644)
}

// Clear clears the store and writes it to the file specified in env.AuthStorePath
func (store *AuthStore) Clear() error {
	store = &AuthStore{}
	return store.WriteToCache()
}

// ReadFromCache reads an AuthStore into the global store from the file specified in env.AuthStorePath
//
// This function should be run in order to load the authentication info from the cache. If it is not, the global AuthStore will be blank.
func ReadFromCache() error {
	cache, err := os.ReadFile(env.AuthStorePath)
	if err != nil {
		if _, err := os.Create(env.AuthStorePath); err != nil {
			return fmt.Errorf("create auth store: %w", err)
		}
		cache = []byte{}
	}

	var store AuthStore
	if err := json.Unmarshal(cache, &store); err != nil {
		store = AuthStore{}
	}
	Store = store
	return nil
}
