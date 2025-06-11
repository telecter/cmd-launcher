package cli

import (
	"fmt"

	"github.com/fatih/color"
)

// Map of translation strings to human-readable text.
var translations = map[string]string{
	"auth.code.fetching":        "Loading device code...",
	"auth.code.display":         "Use the code %s at %s to sign in",
	"auth.browser.opening":      "A web browser will be opened to continue authenticatication.",
	"auth.complete":             "Logged in as %s",
	"auth.logout":               "Logged out from account.",
	"auth.redirect":             "Logged in! You can close this window and return to the launcher.",
	"auth.fail":                 "Failed to log in. An error occurred during authentication: %s",
	"cmd.login":                 "Login in to an account",
	"cmd.logout":                "Log out of an account",
	"cmd.create":                "Create a new instance",
	"cmd.delete":                "Delete an instance",
	"cmd.rename":                "Rename an instance",
	"cmd.list":                  "List all instances",
	"cmd.search":                "Search versions",
	"cmd.start":                 "Start the specified instance",
	"cmd.instance":              "Manage Minecraft instances",
	"cmd.auth":                  "Manage account authentication",
	"cmd.version":               "Display launcher version and about",
	"arg.dir":                   "Root directory to use for launcher",
	"arg.nocolor":               "Disable all color output. The NO_COLOR environment variable is also supported.",
	"cmd.search.query":          "Search query",
	"cmd.search.kind":           "What to search for",
	"cmd.search.reverse":        "Reverse the listing",
	"cmd.start.id":              "Instance to launch",
	"cmd.start.verbose":         "Increase verbosity",
	"cmd.start.username":        "Set your username to the provided value (launches game in offline mode)",
	"cmd.start.server":          "Join a server immediately upon starting the game",
	"cmd.start.demo":            "Start the game in demo mode",
	"cmd.start.disablemp":       "Disable multiplayer",
	"cmd.start.disablechat":     "Disable chat",
	"cmd.start.width":           "Game window width",
	"cmd.start.height":          "Game window height",
	"cmd.start.jvm":             "Path to the JVM to use",
	"cmd.start.minmemory":       "Minimum memory",
	"cmd.start.maxmemory":       "Maximum memory",
	"cmd.start.opts":            "Game Options",
	"cmd.start.overrides":       "Configuration Overrides",
	"cmd.start.downloading":     "Downloading files",
	"cmd.delete.confirm":        "Are you sure you want to delete this instance?",
	"cmd.delete.warning":        "'%s' will be lost forever (a long time!) [y/n] ",
	"cmd.delete.complete":       "Deleted instance '%s'",
	"cmd.delete.abort":          "Operation aborted",
	"cmd.delete.id":             "Instance to delete",
	"cmd.delete.yes":            "Assume yes to all questions",
	"cmd.rename.id":             "Instance to rename",
	"cmd.rename.new":            "New name for instance",
	"cmd.rename.complete":       "Renamed instance.",
	"cmd.create.id":             "Instance name",
	"cmd.create.loader":         "Mod loader to use",
	"cmd.create.version":        "Game version",
	"cmd.create.loaderversion":  "Mod loader version",
	"cmd.create.complete":       "Created instance '%s' with Minecraft %s",
	"cmd.create.complete.extra": "(%s %s)",
	"cmd.create.nojvm":          "Java could not be found. This means you will need to manually set its path in the instance configuration.",
	"cmd.auth.nobrowser":        "Use device code instead of browser for authentication",
	"start.assets":              "Identified %d assets",
	"start.libraries":           "Identified %d libraries",
	"start.metadata":            "Version metadata retrieved",
	"start.launching":           "Launching game with username '%s'",
	"start.debug.jvmargs":       "JVM arguments: %s",
	"start.debug.gameargs":      "Game arguments: %s",
	"start.debug.info":          "Starting main class %q. Game directory is %q.",
	"search.table.version":      "Version",
	"search.table.type":         "Type",
	"search.table.date":         "Release Date",
	"search.table.name":         "Name",
}

// Translate takes a translation string and looks up its human-readable text. If not available, it returns the same translation string.
func Translate(key string, args ...any) string {
	t, ok := translations[key]
	if !ok {
		return key
	}
	return fmt.Sprintf(t, args...)
}

func Translations() map[string]string {
	return translations
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

// Error prints an error message.
//
// Indicates a fatal error.
func Error(format string, a ...any) {
	color.New(color.Bold, color.FgRed).Print("| Error: ")
	fmt.Printf(format+"\n", a...)
}

// Debug prints a debug message.
//
// Used to print information messages useful for debugging the launcher.
func Debug(format string, a ...any) {
	color.New(color.Bold, color.FgMagenta).Print("| Debug: ")
	fmt.Printf(format+"\n", a...)
}
