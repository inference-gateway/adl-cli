package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Print and Println write to stdout through lipgloss's color-profile writer,
// which downsamples truecolor for limited terminals and strips ANSI entirely
// when stdout is not a TTY (CI, pipes, redirects) or NO_COLOR is set. Always
// route styled output through these so non-interactive logs stay clean.
func Print(s string)   { _, _ = lipgloss.Print(s) }
func Println(s string) { _, _ = lipgloss.Println(s) }

// Header renders a section header with the violet accent bar used throughout
// the wizard, so the --defaults run reads as the same product.
func Header(title string) string {
	style := lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	return "\n" + style.Render("▌ "+title)
}

// Row renders a "label › value" line used to echo a resolved default. Empty
// values render as a dim em dash so the column never looks broken.
func Row(label, value string) string {
	labelStyle := lipgloss.NewStyle().Foreground(colorSubtle)
	sep := lipgloss.NewStyle().Foreground(colorPrimary).Render(" › ")
	rendered := value
	if strings.TrimSpace(value) == "" {
		rendered = lipgloss.NewStyle().Foreground(colorSubtle).Render("-")
	}
	return "  " + labelStyle.Render(label) + sep + rendered
}

// Note renders a dimmed, indented helper line.
func Note(text string) string {
	return lipgloss.NewStyle().Foreground(colorSubtle).Render("  " + text)
}

// Bullet renders an emerald check followed by text, confirming an applied
// choice.
func Bullet(text string) string {
	check := lipgloss.NewStyle().Foreground(colorSuccess).Render("✓")
	return "  " + check + " " + text
}
