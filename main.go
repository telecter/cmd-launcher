package main

import (
	"fmt"

	"github.com/alecthomas/kong"
	"github.com/telecter/cmd-launcher/internal/cmd"
	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

const (
	NAME        = "cmd-launcher"
	DESCRIPTION = "A minimal command line Minecraft launcher."
	VERSION     = "1.0.0"
	LICENSE     = "Licensed MIT"
	COPYRIGHT   = "Copyright (c) 2024-2025 telecter"
)

type version struct{}

func (version) Run(ctx *kong.Context) error {
	fmt.Printf("%s %s\n%s\n%s\n%s\n", NAME, VERSION, DESCRIPTION, COPYRIGHT, LICENSE)
	return nil
}

type CLI struct {
	Start    cmd.Start    `cmd:"" help:"Start the specified instance"`
	Instance cmd.Instance `cmd:"" help:"Manage Minecraft instances" aliases:"inst"`
	Auth     cmd.Auth     `cmd:"" help:"Manage account authentication"`
	Search   cmd.Search   `cmd:"" help:"Search versions"`
	Version  version      `cmd:"" help:"Display launcher version and about"`
	Dir      string       `name:"dir" help:"Root directory to use for launcher" type:"path" placeholder:"PATH"`
}

func (c *CLI) AfterApply() error {
	if c.Dir != "" {
		if err := env.SetDirs(c.Dir); err != nil {
			return err
		}
	}
	if err := auth.ReadFromCache(); err != nil {
		return fmt.Errorf("read auth store: %w", err)
	}
	return nil
}

func main() {
	ctx := kong.Parse(&CLI{}, kong.UsageOnError(), kong.Name(NAME), kong.Description(DESCRIPTION))
	err := ctx.Run()

	ctx.FatalIfErrorf(err)
}
