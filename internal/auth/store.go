package auth

import (
	"errors"
	"fmt"
	"os"

	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/network"
)

type AuthStoreData struct {
	Refresh string `json:"refresh"`
}

func GetRefreshToken() string {
	cache := network.JSONCache{Path: internal.AccountDataCache}
	var store AuthStoreData
	if err := cache.Read(&store); err != nil {
		return ""
	}
	return store.Refresh
}
func SetRefreshToken(token string) error {
	cache := network.JSONCache{Path: internal.AccountDataCache}
	if err := cache.Write(AuthStoreData{Refresh: token}); err != nil {
		return err
	}
	return nil
}

func Logout() error {
	if err := os.Remove(internal.AccountDataCache); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("already logged out")
	} else if err != nil {
		return fmt.Errorf("remove account store: %w", err)
	}
	return nil
}

func IsLoggedIn() bool {
	return GetRefreshToken() != ""
}
