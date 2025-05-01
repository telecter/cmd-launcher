package auth

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/telecter/cmd-launcher/internal"
)

type AuthStoreData struct {
	Refresh string `json:"refresh"`
}

func GetRefreshToken() string {
	data, err := os.ReadFile(internal.AccountDataCache)
	if err != nil {
		return ""
	}

	var store AuthStoreData
	if err := json.Unmarshal(data, &store); err != nil {
		return ""
	}
	return store.Refresh
}
func SetRefreshToken(token string) error {
	data, _ := json.Marshal(AuthStoreData{
		Refresh: token,
	})
	if err := os.WriteFile(internal.AccountDataCache, data, 0644); err != nil {
		return err
	}
	return nil
}

func Logout() error {
	if err := os.Remove(internal.AccountDataCache); os.IsNotExist(err) {
		return fmt.Errorf("already logged out")
	} else if err != nil {
		return fmt.Errorf("remove account store: %w", err)
	}
	return nil
}

func IsLoggedIn() bool {
	refresh := GetRefreshToken()
	return refresh != ""
}
