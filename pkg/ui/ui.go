package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

// ─── Legacy aliases (kept so callers don't break) ────────────────────────────

var (
	AppColor    = ColorOrange
	AccentColor = ColorBlue
	SuccessColor = ColorSuccess
	ErrorColor   = ColorError
	WarningColor = ColorWarning

	// Shared text styles
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorDimText).
			Width(18).
			Align(lipgloss.Right)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorWhite).
			PaddingLeft(1)

	SuccessStyle = lipgloss.NewStyle().Foreground(ColorSuccess).Bold(true)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError).Bold(true)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorWarning).Bold(true)
	InfoStyle    = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)
)

// ─── Component styles ─────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite).
			Background(ColorOrange).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(0)

	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorBlue).
				Underline(true).
				MarginTop(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBlue).
			Padding(1, 2).
			MarginTop(1)
)

// ─── Print helpers ────────────────────────────────────────────────────────────

// PrintTitle prints a bold orange-background header bar.
func PrintTitle(title string) {
	fmt.Println(titleStyle.Render("  📦 " + strings.ToUpper(title) + "  "))
}

// PrintError prints a red error line with icon.
func PrintError(msg string) {
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorError).
		Render("✖ ERROR")
	text := lipgloss.NewStyle().Foreground(ColorWhite).Render(msg)
	fmt.Println(" " + badge + "  " + text)
}

// PrintSuccess prints a green success line with icon.
func PrintSuccess(msg string) {
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorSuccess).
		Render("✔")
	text := lipgloss.NewStyle().Foreground(ColorWhite).Render(msg)
	fmt.Println(" " + badge + "  " + text)
}

// PrintInfo prints a sky-blue info line.
func PrintInfo(msg string) {
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorInfo).
		Render("ℹ")
	text := lipgloss.NewStyle().Foreground(ColorDimText).Render(msg)
	fmt.Println(" " + badge + "  " + text)
}

// PrintWarning prints an amber warning line.
func PrintWarning(msg string) {
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorWarning).
		Render("⚠")
	text := lipgloss.NewStyle().Foreground(ColorWhite).Render(msg)
	fmt.Println(" " + badge + "  " + text)
}

// PrintKeyValue prints a right-aligned label with a value.
func PrintKeyValue(key, value string) {
	fmt.Printf("%s%s\n", LabelStyle.Render(key+":"), ValueStyle.Render(value))
}

// RenderBox wraps content in a rounded border with a section header.
func RenderBox(title, content string) {
	inner := lipgloss.JoinVertical(lipgloss.Left,
		sectionHeaderStyle.Render(title),
		lipgloss.NewStyle().MarginTop(1).Render(content),
	)
	fmt.Println(boxStyle.Render(inner))
}

// RenderKeyValueBox renders a box with rows of key/value pairs.
func RenderKeyValueBox(title string, rows [][2]string) {
	out := ""
	for _, r := range rows {
		out += fmt.Sprintf("%s%s\n", LabelStyle.Render(r[0]+":"), ValueStyle.Render(r[1]))
	}
	RenderBox(title, out)
}

// ─── Table helper ─────────────────────────────────────────────────────────────

// BuildTable builds a static (non-interactive) lipgloss table view.
func BuildTable(columns []table.Column, rows []table.Row) string {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(false),
		table.WithHeight(len(rows)+1),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(ColorSubtle).
		BorderBottom(true).
		Bold(true).
		Foreground(ColorOrange)
	s.Selected = s.Selected.
		Foreground(ColorWhite).
		Background(ColorBlue).
		Bold(false)
	t.SetStyles(s)

	return t.View()
}

// ─── Banner ───────────────────────────────────────────────────────────────────

// PrintBanner renders the large app banner for the root help screen.
func PrintBanner() {
	logoLines := []string{
		"  ███████╗██╗     ███████╗██╗  ██╗ ██████╗██╗     ██╗",
		"  ██╔════╝██║     ██╔════╝╚██╗██╔╝██╔════╝██║     ██║",
		"  █████╗  ██║     █████╗   ╚███╔╝ ██║     ██║     ██║",
		"  ██╔══╝  ██║     ██╔══╝   ██╔██╗ ██║     ██║     ██║",
		"  ██║     ███████╗███████╗██╔╝ ██╗╚██████╗███████╗██║",
		"  ╚═╝     ╚══════╝╚══════╝╚═╝  ╚═╝ ╚═════╝╚══════╝╚═╝",
	}
	for _, line := range logoLines {
		fmt.Println(lipgloss.NewStyle().Foreground(ColorOrange).Bold(true).Render(line))
	}
	sub := lipgloss.NewStyle().Foreground(ColorDimText).Render("  Amazon Flex Advanced Operations Manager")
	fmt.Println(sub)
	fmt.Println(lipgloss.NewStyle().Foreground(ColorSubtle).Render("  ─────────────────────────────────────────"))
	fmt.Println()
}
