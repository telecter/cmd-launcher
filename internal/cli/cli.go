package cli

import (
	"errors"
	"fmt"
	"net"

	"github.com/fatih/color"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

// Map of translation strings to human-readable text.
var translations = map[string]string{
	"instance":    "Manage Minecraft instances",
	"auth":        "Manage account authentication",
	"about":       "Display launcher version and about",
	"list":        "List all instances",
	"completions": "Outputs shell command to install completions",

	"login":               "Login in to an account",
	"login.code.fetching": "Loading device code...",
	"login.code":          "Use the code %s at %s to sign in",
	"login.browser":       "A web browser will be opened to continue authenticatication.",
	"login.url":           "If the browser does not open, please copy and paste this URL into your browser: %s",
	"login.complete":      "Logged in as %s",
	"login.redirect":      "Logged in! You can close this window and return to the launcher.",
	"login.redirectfail":  "Failed to log in: An error occurred during authentication.",
	"login.arg.nobrowser": "Use device code instead of browser for authentication",

	"logout":          "Log out of an account",
	"logout.complete": "Logged out from account.",

	"create":                   "Create a new instance",
	"create.complete":          "Created instance '%s' with Minecraft %s (%s%s)",
	"create.arg.id":            "Instance name",
	"create.arg.loader":        "Mod loader",
	"create.arg.version":       "Game version",
	"create.arg.loaderversion": "Mod loader version",

	"delete":          "Delete an instance",
	"delete.confirm":  "Are you sure you want to delete this instance?",
	"delete.warning":  "'%s' will be lost forever (a long time!) [y/n] ",
	"delete.complete": "Deleted instance '%s'",
	"delete.abort":    "Operation aborted",
	"delete.arg.id":   "Instance to delete",
	"delete.arg.yes":  "Assume yes to all questions",

	"rename":          "Rename an instance",
	"rename.complete": "Renamed instance.",
	"rename.arg.id":   "Instance to rename",
	"rename.arg.new":  "New name for instance",

	"search":               "Search versions",
	"search.table.version": "Version",
	"search.table.type":    "Type",
	"search.table.date":    "Release Date",
	"search.table.name":    "Name",
	"search.arg.query":     "Search query",
	"search.arg.kind":      "What to search for",
	"search.arg.reverse":   "Reverse the listing",

	"start":                    "Start the specified instance",
	"start.arg.id":             "Instance to launch",
	"start.arg.verbose":        "Increase verbosity",
	"start.arg.username":       "Set username (offline mode)",
	"start.arg.server":         "Join a server upon starting the game",
	"start.arg.world":          "Join a world upon starting the game",
	"start.arg.demo":           "Start the game in demo mode",
	"start.arg.disablemp":      "Disable multiplayer",
	"start.arg.disablechat":    "Disable chat",
	"start.arg.width":          "Game window width",
	"start.arg.height":         "Game window height",
	"start.arg.jvm":            "Path to the JVM",
	"start.arg.jvmargs":        "Extra JVM arguments",
	"start.arg.minmemory":      "Minimum memory",
	"start.arg.maxmemory":      "Maximum memory",
	"start.arg.opts":           "Game Options",
	"start.arg.overrides":      "Configuration Overrides",
	"start.launch.downloading": "Downloading files",
	"start.launch.assets":      "Identified %d assets",
	"start.launch.libraries":   "Identified %d libraries",
	"start.launch.metadata":    "Version metadata retrieved",
	"start.launch.jvmargs":     "JVM arguments: %s",
	"start.launch.gameargs":    "Game arguments: %s",
	"start.launch.info":        "Starting main class %q. Game directory is %q.",
	"start.launch":             "Launching game with username '%s'",

	"arg.verbosity": "Increase launcher output verbosity",
	"arg.dir":       "Root directory for launcher files",
	"arg.nocolor":   "Disable all color output. The NO_COLOR environment variable is also supported.",

	"tip.internet":  "Check your internet connection.",
	"tip.cache":     "Remote resources were not cached and were unable to be retrieved. Check your Internet connection.",
	"tip.configure": "Configure this instance with the `instance.json` file within the instance directory.",
	"tip.nojvm":     "If a Mojang-provided JVM is not available, you can install it yourself and set the path to the Java executable in the instance configuration.",
	"tip.noaccount": "To launch in offline mode, use the --username (-u) flag.",
}

func Translations() map[string]string {
	return translations
}

// Translate takes a translation string and looks up its human-readable text. If not available, it returns the same translation string.
func Translate(key string) string {
	t, ok := translations[key]
	if !ok {
		return key
	}
	return t
}

// Info prints an general info message.
func Info(format string, a ...any) {
	color.New(color.Bold, color.FgBlue).Print("| ")
	fmt.Printf(format+"\n", a...)
}

// Success prints a success info message.
//
// Indicates a command or task has successfully completed.
func Success(format string, a ...any) {
	color.New(color.Bold, color.FgGreen).Print("| ")
	fmt.Printf(format+"\n", a...)
}

// Warning prints a cautionary message.
//
// Indicates that there may be an issue.
func Warning(format string, a ...any) {
	color.New(color.Bold, color.FgYellow).Print("| Warning: ")
	fmt.Printf(format+"\n", a...)
}

// Debug prints a debug message.
//
// Used to print information messages useful for debugging the launcher.
func Debug(format string, a ...any) {
	color.New(color.Bold, color.FgMagenta).Print("| Debug: ")
	fmt.Printf(format+"\n", a...)
}

// Error prints an error message.
//
// Indicates a fatal error.
func Error(format string, a ...any) {
	color.New(color.Bold, color.FgRed).Print("| Error: ")
	fmt.Printf(format+"\n", a...)
}

func Tip(format string, a ...any) {
	color.New(color.Bold, color.FgYellow).Print("| Tip: ")
	fmt.Printf(format+"\n", a...)
}

// Tips prints a tip message based on an error, if any are available.
func Tips(err error) {
	if errors.Is(err, &net.OpError{}) {
		Tip(Translate("tip.internet"))
	}
	if errors.Is(err, network.ErrNotCached) {
		Tip(Translate("tip.cache"))
	}
	if errors.Is(err, meta.ErrJavaBadSystem) || errors.Is(err, meta.ErrJavaNoVersion) {
		Tip(Translate("tip.nojvm"))
	}
	if errors.Is(err, auth.ErrNoAccount) {
		Tip(Translate("tip.noaccount"))
	}
}
