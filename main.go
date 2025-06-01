package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/telecter/cmd-launcher/cmd"
	"github.com/telecter/cmd-launcher/internal"
	"github.com/telecter/cmd-launcher/internal/auth"
)

type CLI struct {
	Start  cmd.Start  `cmd:"" help:"Start the specified instance"`
	Auth   cmd.Auth   `cmd:"" help:"Manage account authentication"`
	Create cmd.Create `cmd:"" help:"Create a new Minecraft instance"`
	Search cmd.Search `cmd:"" help:"Search versions and instances"`
	Delete cmd.Delete `cmd:"" help:"Delete a Minecraft instance"`
	Dir    string     `name:"dir" help:"Root directory to use for launcher" type:"path" placeholder:"PATH"`
}

func (c *CLI) AfterApply() error {
	if c.Dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get user home directory: %w", err)
		}
		c.Dir = filepath.Join(home, ".minecraft")
	}
	if err := internal.SetDirs(c.Dir); err != nil {
		return err
	}

	if err := auth.ReadFromCache(); err != nil {
		return fmt.Errorf("read auth store: %w", err)
	}

	return nil
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.UsageOnError(), kong.Description("A minimal command line Minecraft launcher."))
	err := ctx.Run()

	ctx.FatalIfErrorf(err)
}
