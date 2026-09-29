package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBar manages smooth, animated, real-time terminal progress for uploads and downloads.
type ProgressBar struct {
	label      string
	filename   string
	totalBytes int64
	disabled   bool
	startTime  time.Time
	lastTime   time.Time
	lastBytes  int64
	currSpeed  float64
	mu         sync.Mutex
	spinIdx    int
	finalBytes int64
}

var spinFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewProgressBar initializes a new progress tracker.
func NewProgressBar(label, filename string, totalBytes int64, disabled bool) *ProgressBar {
	now := time.Now()
	return &ProgressBar{
		label:      label,
		filename:   TruncateString(filename, 28),
		totalBytes: totalBytes,
		disabled:   disabled,
		startTime:  now,
		lastTime:   now,
	}
}

// Update renders an animated progress frame with written and total bytes.
func (p *ProgressBar) Update(written, total int64) {
	if p.disabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.finalBytes = written
	if total > 0 {
		p.totalBytes = total
	}

	now := time.Now()
	dt := now.Sub(p.lastTime).Seconds()
	if dt >= 0.15 {
		db := float64(written - p.lastBytes)
		if dt > 0 {
			instantSpeed := db / dt
			if p.currSpeed == 0 {
				p.currSpeed = instantSpeed
			} else {
				p.currSpeed = p.currSpeed*0.65 + instantSpeed*0.35 // Smooth moving average
			}
		}
		p.lastBytes = written
		p.lastTime = now
		p.spinIdx = (p.spinIdx + 1) % len(spinFrames)
	}

	spinnerChar := lipgloss.NewStyle().Foreground(ColorCyanBright).Render(spinFrames[p.spinIdx])
	labelStyled := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Render(p.label)
	fileStyled := lipgloss.NewStyle().Foreground(ColorGrayLight).Render(p.filename)

	if p.totalBytes <= 0 {
		speedStr := ""
		if p.currSpeed > 0 {
			speedStr = fmt.Sprintf(" • %s/s", FormatBytes(int64(p.currSpeed)))
		}
		fmt.Printf("\r\033[K  %s %s %s  %s%s", spinnerChar, labelStyled, fileStyled, FormatBytes(written), speedStr)
		return
	}

	pct := float64(written) / float64(p.totalBytes)
	if pct > 1.0 {
		pct = 1.0
	}

	barWidth := 22
	filled := int(pct * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	fillStr := strings.Repeat("━", filled)
	fillStyled := lipgloss.NewStyle().Foreground(ColorCyanBright).Render(fillStr)

	head := ""
	if filled > 0 && filled < barWidth {
		head = lipgloss.NewStyle().Foreground(ColorEmeraldAccent).Render("╸")
		empty--
	}
	if empty < 0 {
		empty = 0
	}
	emptyStr := strings.Repeat("━", empty)
	emptyStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("#27272a")).Render(emptyStr)

	speedStr := ""
	if p.currSpeed > 0 {
		speedStr = fmt.Sprintf(" %s/s", FormatBytes(int64(p.currSpeed)))
	}

	fmt.Printf("\r\033[K  %s %s %s  [%s%s%s] %5.1f%%  (%s / %s)%s",
		spinnerChar,
		labelStyled,
		fileStyled,
		fillStyled,
		head,
		emptyStyled,
		pct*100.0,
		FormatBytes(written),
		FormatBytes(p.totalBytes),
		lipgloss.NewStyle().Foreground(ColorGrayMuted).Render(speedStr),
	)
}

// UpdateChunk updates chunked upload progress.
func (p *ProgressBar) UpdateChunk(chunk, totalChunks int, totalBytes int64) {
	if p.disabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.spinIdx = (p.spinIdx + 1) % len(spinFrames)
	spinnerChar := lipgloss.NewStyle().Foreground(ColorCyanBright).Render(spinFrames[p.spinIdx])
	labelStyled := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Render(p.label)
	fileStyled := lipgloss.NewStyle().Foreground(ColorGrayLight).Render(p.filename)

	pct := float64(chunk) / float64(totalChunks)
	if pct > 1.0 {
		pct = 1.0
	}

	barWidth := 22
	filled := int(pct * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled

	fillStr := strings.Repeat("━", filled)
	fillStyled := lipgloss.NewStyle().Foreground(ColorCyanBright).Render(fillStr)

	head := ""
	if filled > 0 && filled < barWidth {
		head = lipgloss.NewStyle().Foreground(ColorEmeraldAccent).Render("╸")
		empty--
	}
	if empty < 0 {
		empty = 0
	}
	emptyStr := strings.Repeat("━", empty)
	emptyStyled := lipgloss.NewStyle().Foreground(lipgloss.Color("#27272a")).Render(emptyStr)

	sizeNote := ""
	if totalBytes > 0 {
		sizeNote = fmt.Sprintf(" (%s)", FormatBytes(totalBytes))
	}

	fmt.Printf("\r\033[K  %s %s %s  [%s%s%s] %5.1f%%  (chunk %d/%d)%s",
		spinnerChar,
		labelStyled,
		fileStyled,
		fillStyled,
		head,
		emptyStyled,
		pct*100.0,
		chunk,
		totalChunks,
		lipgloss.NewStyle().Foreground(ColorGrayMuted).Render(sizeNote),
	)
}

// UpdateIndeterminate updates indeterminate streaming progress.
func (p *ProgressBar) UpdateIndeterminate(written int64) {
	if p.disabled {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.finalBytes = written
	p.spinIdx = (p.spinIdx + 1) % len(spinFrames)
	spinnerChar := lipgloss.NewStyle().Foreground(ColorCyanBright).Render(spinFrames[p.spinIdx])
	labelStyled := lipgloss.NewStyle().Bold(true).Foreground(ColorWhite).Render(p.label)
	fileStyled := lipgloss.NewStyle().Foreground(ColorGrayLight).Render(p.filename)

	fmt.Printf("\r\033[K  %s %s %s  %s", spinnerChar, labelStyled, fileStyled, FormatBytes(written))
}

// Finish clears the progress line completely.
func (p *ProgressBar) Finish() {
	if p.disabled {
		return
	}
	fmt.Printf("\r\033[2K\r")
}

// FinalBytes returns the last byte count recorded by the progress bar.
func (p *ProgressBar) FinalBytes() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.finalBytes
}

// Elapsed returns the time elapsed since the start of the progress bar.
func (p *ProgressBar) Elapsed() time.Duration {
	return time.Since(p.startTime)
}
