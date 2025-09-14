package output

import (
	"errors"
	"fmt"
	"net"

	"github.com/fatih/color"
	"github.com/telecter/cmd-launcher/internal/meta"
	"github.com/telecter/cmd-launcher/internal/network"
	"github.com/telecter/cmd-launcher/pkg/auth"
)

// Info prints an general informational message.
func Info(format string, a ...any) {
	color.New(color.Bold, color.FgBlue).Print("| ")
	fmt.Printf(format+"\n", a...)
}

// Success prints a success information message.
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
	color.New(color.Bold, color.FgYellow).Printf("| %s: ", Translate("launcher.warning"))
	fmt.Printf(format+"\n", a...)
}

// Debug prints a debug message.
//
// Used to print information messages useful for debugging the launcher.
func Debug(format string, a ...any) {
	color.New(color.Bold, color.FgMagenta).Printf("| %s: ", Translate("launcher.debug"))
	fmt.Printf(format+"\n", a...)
}

// Error prints an error message.
//
// Indicates a fatal error.
func Error(format string, a ...any) {
	color.New(color.Bold, color.FgRed).Printf("| %s: ", Translate("launcher.error"))
	fmt.Printf(format+"\n", a...)
}

// Tip prints a tip message.
//
// Indicates an action that should be performed.
func Tip(format string, a ...any) {
	color.New(color.Bold, color.FgYellow).Printf("| %s: ", Translate("launcher.tip"))
	fmt.Printf(format+"\n", a...)
}

// Tips prints a tip message based on an error, if any are available.
func Tips(err error) {
	// General internet connection related issues
	if errors.Is(err, &net.OpError{}) {
		Tip(Translate("tip.internet"))
	}
	// A cache couldn't be updated from the remote source
	if errors.Is(err, network.ErrNotCached) {
		Tip(Translate("tip.cache"))
	}
	// Mojang-provided JVM isn't working
	if errors.Is(err, meta.ErrJavaBadSystem) || errors.Is(err, meta.ErrJavaNoVersion) {
		Tip(Translate("tip.nojvm"))
	}
	// Not logged in
	if errors.Is(err, auth.ErrNoAccount) {
		Tip(Translate("tip.noaccount"))
	}
}
