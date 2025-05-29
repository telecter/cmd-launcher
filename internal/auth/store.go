package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/telecter/cmd-launcher/internal"
)

type msaAuthStore struct {
	AccessToken  string    `json:"access_token"`
	Expires      time.Time `json:"expires"`
	RefreshToken string    `json:"refresh_token"`
}
type xblAuthStore struct {
	Userhash string    `json:"uhs"`
	Token    string    `json:"token"`
	Expires  time.Time `json:"expires"`
}
type xstsAuthStore struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}
type minecraftAuthStore struct {
	AccessToken string    `json:"access_token"`
	Expires     time.Time `json:"expires"`
}
type AuthStore struct {
	MSA       msaAuthStore       `json:"msa"`
	XBL       xblAuthStore       `json:"xbl"`
	XSTS      xstsAuthStore      `json:"xsts"`
	Minecraft minecraftAuthStore `json:"minecraft"`
}

func GetStore() (AuthStore, error) {
	cache, err := os.ReadFile(internal.AccountDataCache)
	if err != nil {
		return AuthStore{}, fmt.Errorf("read auth store: %w", err)
	}

	var store AuthStore
	if err := json.Unmarshal(cache, &store); err != nil {
		return AuthStore{}, fmt.Errorf("parse auth store: %w", err)
	}
	return store, nil
}

func SetStore(store AuthStore) error {
	data, _ := json.MarshalIndent(store, "", "    ")
	return os.WriteFile(internal.AccountDataCache, data, 0644)
}

func Logout() error {
	return SetStore(AuthStore{})
}

func IsLoggedIn() bool {
	store, err := GetStore()
	if err != nil {
		return false
	}
	return store.MSA.RefreshToken != ""
}
