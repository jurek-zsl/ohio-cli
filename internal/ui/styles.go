package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Theme Colors - Sleek Modern Dark Palette
	ColorBgDark        = lipgloss.Color("#09090b") // Zinc 950
	ColorCardBg        = lipgloss.Color("#121217") // Dark slate background
	ColorBorder        = lipgloss.Color("#27272a") // Zinc 800 subtle border
	ColorBorderFocused = lipgloss.Color("#38bdf8") // Sky 400
	ColorRowSelected   = lipgloss.Color("#1e1e2e") // Catppuccin / modern dark selection

	ColorPurplePrimary = lipgloss.Color("#a855f7") // Modern violet
	ColorPurpleDark    = lipgloss.Color("#581c87")
	ColorCyanAccent    = lipgloss.Color("#06b6d4")
	ColorCyanBright    = lipgloss.Color("#38bdf8") // Sky bright
	ColorEmeraldAccent = lipgloss.Color("#10b981") // Mint / emerald
	ColorRoseAccent    = lipgloss.Color("#f43f5e") // Rose
	ColorAmberAccent   = lipgloss.Color("#f59e0b") // Amber
	ColorGrayMuted     = lipgloss.Color("#71717a") // Zinc 500
	ColorGrayDark      = lipgloss.Color("#1f2937") // Slate 800
	ColorGrayLight     = lipgloss.Color("#e4e4e7") // Zinc 200
	ColorWhite         = lipgloss.Color("#fafafa") // Zinc 50

	// Header & Logos
	TopBarBulletStyle = lipgloss.NewStyle().
				Foreground(ColorCyanBright).
				Bold(true)

	TopBarLogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	TopBarVersionStyle = lipgloss.NewStyle().
				Foreground(ColorGrayMuted).
				Background(ColorGrayDark).
				Padding(0, 1)

	HeaderLogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyanBright)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorWhite)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorGrayMuted)

	LivePulseStyle = lipgloss.NewStyle().
			Foreground(ColorEmeraldAccent).
			Bold(true)

	// Navigation Tabs - Sleek, understated minimalist styling
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyanBright).
			Background(lipgloss.Color("#1e293b")).
			Padding(0, 2).
			MarginRight(1)

	TabInactiveStyle = lipgloss.NewStyle().
				Foreground(ColorGrayMuted).
				Padding(0, 2).
				MarginRight(1)

	// Badges & Pills - Sleek, subtle dark badge aesthetic
	PillPublicStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#34d399")).
			Background(lipgloss.Color("#064e3b")).
			Bold(true).
			Padding(0, 1)

	PillPrivateStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a1a1aa")).
				Background(lipgloss.Color("#27272a")).
				Padding(0, 1)

	PillOneTimeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fb7185")).
				Background(lipgloss.Color("#4c0519")).
				Bold(true).
				Padding(0, 1)

	PillPasswordStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#fbbf24")).
				Background(lipgloss.Color("#451a03")).
				Bold(true).
				Padding(0, 1)

	PillSessionStyle = lipgloss.NewStyle().
				Foreground(ColorCyanBright).
				Background(lipgloss.Color("#18181b")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	// Cards & Containers - Sleek subtle borders
	BoxCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)

	BoxCardActiveStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorBorderFocused).
				Padding(0, 1)

	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorCyanBright).
			Background(ColorCardBg).
			Padding(1, 2)

	// Table Styles - Dimmed headers, clean selection indicator
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorGrayMuted).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ColorBorder).
				Padding(0, 1)

	CliTableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorGrayMuted)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorGrayLight).
			Padding(0, 1)

	TableSelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorWhite).
				Background(ColorRowSelected).
				Padding(0, 1)

	// Status Messages
	StatusSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorEmeraldAccent)

	StatusErrorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorRoseAccent)

	StatusWarningStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAmberAccent)

	StatusInfoStyle = lipgloss.NewStyle().
			Foreground(ColorCyanBright)

	// Helper & Keybinding Hints
	KeyHelpStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorCyanBright)

	DescHelpStyle = lipgloss.NewStyle().
			Foreground(ColorGrayMuted)
)
