package main

import (
	"cmd-launcher/pkg/api"
	"cmd-launcher/pkg/launcher"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

func main() {
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

	fmt.Println("Saving auth file...")
	err = os.WriteFile(rootDir+"/account.txt", []byte(refresh), os.ModePerm)
	if err != nil {
		fmt.Println("Couldn't save auth data file. Authentication may not work.")
	}

	if err != nil {
		fmt.Println("Authentication could not be completed. Using offline mode.", err)
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: cmd-launcher <version>")
		os.Exit(1)
	}
	if err := launcher.Launch(os.Args[1], rootDir, authData); err != nil {
		fmt.Println(err)
	}
}
