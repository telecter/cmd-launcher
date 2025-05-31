package main

import (
	"encoding/json"
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
	Dir    string     `name:"dir" help:"Root directory to use for launcher"`
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

	cache, err := os.ReadFile(internal.AuthStorePath)
	if err != nil {
		return fmt.Errorf("read auth store: %w", err)
	}

	var store auth.AuthStore
	if err := json.Unmarshal(cache, &store); err != nil {
		return fmt.Errorf("parse auth store: %w", err)
	}
	auth.Store = store

	return nil
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.UsageOnError(), kong.Description("A minimal command line Minecraft launcher."))
	err := ctx.Run()

	ctx.FatalIfErrorf(err)
}
