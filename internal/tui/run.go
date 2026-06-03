package tui

import (
	"errors"
	"fmt"
	"os"

	"charm.land/huh/v2"
	"golang.org/x/term"
)

// IsTTY reports whether both stdin and stdout are connected to a terminal. The
// interactive wizard is only safe to launch when this is true: huh's Bubble Tea
// program needs a controlling TTY and otherwise errors trying to open /dev/tty
// (e.g. in CI, Docker, or when stdin is piped). Callers fall back to the plain,
// non-interactive flow when this returns false.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd()))
}

// RunForm runs a huh form with the brand theme applied. Ctrl+C / Esc aborts the
// whole wizard cleanly (exit 0), matching the behavior of the previous
// readline-based prompts. Setting ACCESSIBLE=1 switches huh into plain,
// screen-reader friendly prompts.
func RunForm(form *huh.Form) error {
	err := form.
		WithTheme(BrandTheme()).
		WithAccessible(os.Getenv("ACCESSIBLE") != "").
		Run()
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println()
		os.Exit(0)
	}
	return err
}
