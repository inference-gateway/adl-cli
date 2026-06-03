// Package tui provides the styling and interactive-form layer for the adl CLI.
// It wraps charmbracelet/huh (forms) and charmbracelet/lipgloss (styling) behind
// a small, brand-themed API so command code stays thin.
//
// Color is handled entirely by lipgloss/termenv underneath, which honors
// NO_COLOR and CLICOLOR automatically and downsamples truecolor to whatever the
// terminal supports. No extra handling is required for monochrome terminals.
package tui

import (
	"image/color"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// Brand palette. Violet primary (the recognizable Charm look) with emerald for
// success. The subtle gray is chosen to read on both light and dark terminals.
var (
	colorPrimary color.Color = lipgloss.Color("#7D56F4") // violet
	colorAccent  color.Color = lipgloss.Color("#B69CFF") // light lavender
	colorSuccess color.Color = lipgloss.Color("#10B981") // emerald
	colorError   color.Color = lipgloss.Color("#EF4444") // red
	colorOnBrand color.Color = lipgloss.Color("#FAFAFA") // near-white text on a brand background
	colorSubtle  color.Color = lipgloss.Color("#9CA3AF") // gray, readable on dark and light
)

// BrandTheme returns the huh theme used by every adl init form: violet titles
// and selectors, a violet left-border accent on the focused field, and emerald
// multi-select check marks. It is built from huh's base theme so it inherits
// sensible defaults and adapts to the terminal's light/dark background.
func BrandTheme() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		t := huh.ThemeBase(isDark)

		t.Focused.Base = t.Focused.Base.BorderForeground(colorPrimary)
		t.Focused.Title = t.Focused.Title.Foreground(colorPrimary).Bold(true)
		t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(colorPrimary).Bold(true)
		t.Focused.Description = t.Focused.Description.Foreground(colorSubtle)
		t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(colorPrimary)
		t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(colorPrimary).Bold(true)
		t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(colorPrimary)
		t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(colorSuccess)
		t.Focused.FocusedButton = t.Focused.FocusedButton.Background(colorPrimary).Foreground(colorOnBrand)
		t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(colorPrimary)
		t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(colorPrimary)
		t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(colorError)
		t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(colorError)

		t.Blurred.Title = t.Blurred.Title.Foreground(colorSubtle)
		t.Blurred.Description = t.Blurred.Description.Foreground(colorSubtle)

		return t
	})
}
