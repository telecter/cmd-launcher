package main

import (
	"cmd-launcher/pkg/launcher"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: cmd-launcher <version>")
		os.Exit(1)
	}
	if err := launcher.Launch(os.Args[1]); err != nil {
		fmt.Println(err)
	}
}
