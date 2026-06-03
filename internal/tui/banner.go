package tui

import (
	"charm.land/lipgloss/v2"
)

// Banner returns the rounded, violet-accented intro card shown at the top of
// `adl init`.
func Banner() string {
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorOnBrand).
		Background(colorPrimary).
		Padding(0, 1).
		Render("adl")

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		Render("Agent Definition Language")

	heading := lipgloss.JoinHorizontal(lipgloss.Center, badge, "  ", title)

	subtitle := lipgloss.NewStyle().
		Foreground(colorAccent).
		Render("Scaffold an A2A agent - generate Go, Rust, or TypeScript.")

	body := lipgloss.JoinVertical(lipgloss.Left, heading, "", subtitle)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 3).
		Render(body)
}

// PrintBanner writes the intro banner to stdout, padded with blank lines.
func PrintBanner() {
	Println("")
	Println(Banner())
	Println("")
}
