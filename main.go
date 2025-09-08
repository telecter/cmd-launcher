package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Xuanwo/go-locale"
	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/telecter/cmd-launcher/internal/cli"
	"github.com/telecter/cmd-launcher/internal/cli/cmd"
	"go.abhg.dev/komplete"

	env "github.com/telecter/cmd-launcher/pkg"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

const (
	name    = "cmd-launcher"
	version = "1.4.0"
)

type aboutCmd struct{}

func (aboutCmd) Run(ctx *kong.Context) error {
	color.New(color.Bold).Println(name, version)
	color.New(color.Underline).Println(cli.Translate("launcher.description"))
	fmt.Println(cli.Translate("launcher.copyright"))
	fmt.Println(cli.Translate("launcher.license"))
	return nil
}

type CLI struct {
	Start       cmd.StartCmd     `cmd:"" help:"${start}"`
	Instance    cmd.InstanceCmd  `cmd:"" help:"${instance}" aliases:"inst"`
	Auth        cmd.AuthCmd      `cmd:"" help:"${auth}"`
	Search      cmd.SearchCmd    `cmd:"" help:"${search}"`
	About       aboutCmd         `cmd:"" help:"${about}"`
	Completions komplete.Command `cmd:"" help:"${completions}"`
	Verbosity   string           `help:"${arg_verbosity}" enum:"info,extra,debug" default:"info"`
	Dir         string           `help:"${arg_dir}" type:"path" placeholder:"PATH"`
	NoColor     bool             `help:"${arg_nocolor}"`
}

func (c *CLI) AfterApply(ctx *kong.Context) error {
	var verbosity int
	switch c.Verbosity {
	case "info":
		verbosity = 0
	case "extra":
		verbosity = 1
	case "debug":
		verbosity = 2
	}
	ctx.Bind(verbosity)
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

func vars() kong.Vars {
	vars := make(kong.Vars)
	for k, v := range cli.Translations() {
		vars[strings.ReplaceAll(k, ".", "_")] = v
	}
	return vars
}

func valueFormatter(value *kong.Value) string {
	if value.Enum != "" {
		return fmt.Sprintf("%s [%s]", value.Help, strings.Join(value.EnumSlice(), ", "))
	}
	return value.Help
}

func groups() kong.Groups {
	return kong.Groups{
		"overrides": cli.Translate("start.arg.overrides"),
		"opts":      cli.Translate("start.arg.opts"),
	}
}

func main() {
	lang, err := locale.Detect()
	if err == nil {
		cli.SetLang(lang)
	}

	parser := kong.Must(&CLI{},
		kong.UsageOnError(),
		kong.Name(name),
		kong.Description(cli.Translate("launcher.description")),
		kong.ConfigureHelp(kong.HelpOptions{
			NoExpandSubcommands: true,
		}),
		kong.ValueFormatter(valueFormatter),
		groups(),
		vars(),
	)
	komplete.Run(parser)

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

	if err := ctx.Run(); err != nil {
		cli.Error("%s", err)
		cli.Tips(err)
		var coder kong.ExitCoder
		if errors.As(err, &coder) {
			ctx.Exit(coder.ExitCode())
		}
		ctx.Exit(1)
	}
}
