package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/internal/cli/cmd"

	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

const (
	NAME        = "cmd-launcher"
	DESCRIPTION = "A minimal command line Minecraft launcher."
	VERSION     = "1.1.0"
	LICENSE     = "Licensed MIT"
	COPYRIGHT   = "Copyright (c) 2024-2025 telecter"
)

type version struct{}

func (version) Run(ctx *kong.Context) error {
	fmt.Printf("%s %s\n%s\n%s\n%s\n", NAME, VERSION, DESCRIPTION, COPYRIGHT, LICENSE)
	return nil
}

type CLI struct {
	Start     cmd.Start    `cmd:"" help:"${cmd_start}"`
	Instance  cmd.Instance `cmd:"" help:"${cmd_instance}" aliases:"inst"`
	Auth      cmd.Auth     `cmd:"" help:"${cmd_auth}"`
	Search    cmd.Search   `cmd:"" help:"${cmd_search}"`
	Version   version      `cmd:"" help:"${cmd_version}"`
	Verbosity int          `short:"v" type:"counter"`
	Dir       string       `help:"${arg_dir}" type:"path" placeholder:"PATH"`
	NoColor   bool         `help:"${arg_nocolor}"`
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
	if c.NoColor {
		color.NoColor = true
	}
	return nil
}

func main() {
	vars := make(kong.Vars)
	for k, v := range cli.Translations() {
		vars[strings.ReplaceAll(k, ".", "_")] = v
	}

	c := &CLI{}
	parser := kong.Must(c,
		kong.UsageOnError(),
		kong.Name(NAME),
		kong.Description(DESCRIPTION),
		kong.ConfigureHelp(kong.HelpOptions{
			NoExpandSubcommands: true,
		}),
		kong.Groups{
			"overrides": cli.Translate("cmd.start.overrides"),
			"opts":      cli.Translate("cmd.start.opts"),
		},
		vars,
	)
	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		exitCode := 1
		var parseErr *kong.ParseError
		if errors.As(err, &parseErr) {
			parseErr.Context.PrintUsage(false)
			exitCode = parseErr.ExitCode()
		}
		cli.Error("%s", err)
		parser.Exit(exitCode)
	}

	if err := ctx.Run(c.Verbosity); err != nil {
		cli.Error("%s", err)
		var coder kong.ExitCoder
		if errors.As(err, &coder) {
			ctx.Exit(coder.ExitCode())
		}
		ctx.Exit(1)
	}
}
