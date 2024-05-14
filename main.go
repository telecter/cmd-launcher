package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/telecter/cmd-launcher/pkg/api"
	"github.com/telecter/cmd-launcher/pkg/launcher"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: cmd-launcher <version>")
		os.Exit(1)
	}
	var modLoader string
	if len(os.Args) > 2 {
		modLoader = os.Args[2]
	}

	homeDir, _ := os.UserHomeDir()
	rootDir := homeDir + "/.minecraft"
	accountFile, err := os.ReadFile(rootDir + "/account.txt")

	var refresh string
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println("No account data file found")
		refresh = ""
	} else if err == nil {
		refresh = string(accountFile)
	}
	authData, err := api.GetAuthData(refresh)
	refresh = authData.Refresh

	err = os.WriteFile(rootDir+"/account.txt", []byte(refresh), os.ModePerm)
	if err != nil {
		fmt.Println("Couldn't save auth data file. Authentication may not work.")
	}

	if err != nil {
		fmt.Println("Authentication could not be completed. Using offline mode.", err)
	}

	if err := launcher.Launch(os.Args[1], rootDir, launcher.LaunchOptions{
		ModLoader: modLoader,
	}, authData); err != nil {
		fmt.Println(err)
	}
}
