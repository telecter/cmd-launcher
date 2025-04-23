package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/telecter/cmd-launcher/internal/env"
)

func getAccountDataFilePath() string {
	return filepath.Join(env.RootDir, "account.txt")
}

func GetRefreshToken() string {
	data, err := os.ReadFile(getAccountDataFilePath())

	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		panic(err)
	}

	return string(data)
}
func SetRefreshToken(token string) {
	if err := os.WriteFile(getAccountDataFilePath(), []byte(token), 0644); err != nil {
		panic(err)
	}
}

func Logout() error {
	if err := os.Remove(getAccountDataFilePath()); os.IsNotExist(err) {
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
