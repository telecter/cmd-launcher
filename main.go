package main

import (
	"cmd-launcher/pkg/api"
	"cmd-launcher/pkg/launcher"
	"fmt"
	"os"
)

func main() {
	token, uuid, username, err := api.GetAuthData()
	if err != nil {
		fmt.Println("Authentication could not be completed. Using offline mode.", err)
	}
	if len(os.Args) < 2 {
		fmt.Println("usage: cmd-launcher <version>")
		os.Exit(1)
	}
	if err := launcher.Launch(os.Args[1], token, username, uuid); err != nil {
		fmt.Println(err)
	}
}
