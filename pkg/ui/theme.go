package ui

import "github.com/charmbracelet/lipgloss"

// ─── Design Tokens ────────────────────────────────────────────────────────────

var (
	// Brand palette
	ColorOrange  = lipgloss.Color("#FF9900")
	ColorBlue    = lipgloss.Color("#146EB4")
	ColorSuccess = lipgloss.Color("#22C55E")
	ColorError   = lipgloss.Color("#EF4444")
	ColorWarning = lipgloss.Color("#F59E0B")
	ColorInfo    = lipgloss.Color("#38BDF8")
	ColorMuted   = lipgloss.Color("#6B7280")
	ColorSubtle  = lipgloss.Color("#374151")
	ColorWhite   = lipgloss.Color("#F9FAFB")
	ColorDimText = lipgloss.Color("#9CA3AF")

	// Surfaces
	SurfaceDark   = lipgloss.Color("#0F0F0F")
	SurfaceCard   = lipgloss.Color("#1A1A1A")
	SurfaceActive = lipgloss.Color("#1E293B")
)

// ─── Status Badge ─────────────────────────────────────────────────────────────

// StatusBadge renders a colored status pill.
func StatusBadge(status string) string {
	var bg lipgloss.Color
	switch status {
	case "DELIVERED":
		bg = ColorSuccess
	case "UNDELIVERABLE", "FAILED":
		bg = ColorError
	case "PENDING", "":
		bg = ColorWarning
	default:
		bg = ColorBlue
	}
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#000000")).
		Background(bg).
		Padding(0, 1).
		Render(status)
}

// ─── Divider ──────────────────────────────────────────────────────────────────

// Divider renders a horizontal rule using the given width.
func Divider(width int) string {
	line := ""
	for i := 0; i < width; i++ {
		line += "─"
	}
	return lipgloss.NewStyle().Foreground(ColorSubtle).Render(line)
}

// ─── Spinner frames (used for async hints) ────────────────────────────────────

var SpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
