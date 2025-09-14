package main

import (
	"github.com/telecter/cmd-launcher/internal/cli"
)

func main() {
	// Parse and run the main CLI.
	exiter, code := cli.Run()
	exiter(code)
}
