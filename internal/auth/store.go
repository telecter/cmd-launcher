package auth

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/telecter/cmd-launcher/internal/env"
)

type AuthStoreData struct {
	Refresh string `json:"refresh"`
}

func GetRefreshToken() string {
	data, err := os.ReadFile(env.AccountDataCache)
	if err != nil {
		return ""
	}

	var authStoreData AuthStoreData
	if err := json.Unmarshal(data, &authStoreData); err != nil {
		return ""
	}
	return authStoreData.Refresh
}
func SetRefreshToken(token string) error {
	data, _ := json.Marshal(AuthStoreData{
		Refresh: token,
	})
	if err := os.WriteFile(env.AccountDataCache, data, 0644); err != nil {
		return err
	}
	return nil
}

func Logout() error {
	if err := os.Remove(env.AccountDataCache); os.IsNotExist(err) {
		return fmt.Errorf("already logged out")
	} else if err != nil {
		return fmt.Errorf("error removing account information: %w", err)
	}
	return nil
}

func IsLoggedIn() bool {
	refresh := GetRefreshToken()
	return refresh != ""
}
