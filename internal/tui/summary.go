package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// Field is a single labelled value rendered inside the summary panel.
type Field struct {
	Label string
	Value string
}

// SummaryBox renders the closing panel that confirms the manifest was written
// and recaps the key choices. The title is shown with an emerald check mark.
func SummaryBox(title string, fields []Field) string {
	check := lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Render("✓")
	head := check + " " + lipgloss.NewStyle().Bold(true).Foreground(colorSuccess).Render(title)

	labelStyle := lipgloss.NewStyle().Foreground(colorSubtle).Width(14)
	valueStyle := lipgloss.NewStyle().Foreground(colorPrimary)

	rows := []string{head}
	for _, f := range fields {
		rows = append(rows, labelStyle.Render(f.Label)+valueStyle.Render(f.Value))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSuccess).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// NextSteps renders the numbered "what to do next" list shown after init.
func NextSteps(steps []string) string {
	numStyle := lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	heading := lipgloss.NewStyle().Bold(true).Render("Next steps")

	var b strings.Builder
	b.WriteString(heading)
	b.WriteString("\n")
	for i, s := range steps {
		fmt.Fprintf(&b, "  %s %s\n", numStyle.Render(fmt.Sprintf("%d.", i+1)), s)
	}
	return strings.TrimRight(b.String(), "\n")
}
