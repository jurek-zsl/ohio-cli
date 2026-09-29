package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/skip2/go-qrcode"
)

// GenerateTerminalQRCode creates a compact Unicode terminal representation of a QR code.
// The output uses half-blocks to render square QR codes directly in terminal windows.
func GenerateTerminalQRCode(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	// ToSmallString(false) renders standard black-on-white blocks suitable for dark backgrounds
	qrStr := qr.ToSmallString(false)
	lines := strings.Split(qrStr, "\n")

	var b strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

// GenerateInvertedTerminalQRCode creates an inverted QR code if user terminal requires it.
func GenerateInvertedTerminalQRCode(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	qrStr := qr.ToSmallString(true)
	lines := strings.Split(qrStr, "\n")

	var b strings.Builder
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

// RenderFramedQRCode creates an ASCII QR code neatly framed in a subtle rounded border.
func RenderFramedQRCode(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	qrStr := strings.TrimSpace(qr.ToSmallString(false))
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1).
		MarginLeft(2).
		Render(qrStr)

	return box, nil
}

