package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/telecter/cmd-launcher/internal"
)

type AuthStoreData struct {
	Refresh string `json:"refresh"`
}

func GetRefreshToken() string {
	cache, err := os.ReadFile(internal.AccountDataCache)
	if err != nil {
		return ""
	}

	var store AuthStoreData
	if err := json.Unmarshal(cache, &store); err != nil {
		return ""
	}
	return store.Refresh
}
func SetRefreshToken(token string) error {
	data, _ := json.Marshal(AuthStoreData{Refresh: token})
	return os.WriteFile(internal.AccountDataCache, data, 0644)
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
