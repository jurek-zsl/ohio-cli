package ui

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// FormatBytes formats byte sizes into human-readable strings (e.g. 1.2 MB, 450 KB).
func FormatBytes(bytes int64) string {
	if bytes <= 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KB", "MB", "GB", "TB", "PB"}
	val := float64(bytes) / float64(div)
	if val >= 10 || math.Mod(val*10, 10) == 0 {
		return fmt.Sprintf("%.1f %s", val, units[exp])
	}
	return fmt.Sprintf("%.2f %s", val, units[exp])
}

// FormatDuration formats a time.Duration into a concise string (e.g. "12s", "4m 12s").
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	m := d / time.Minute
	s := (d % time.Minute) / time.Second
	if m == 0 {
		return fmt.Sprintf("%ds", s)
	}
	return fmt.Sprintf("%dm %ds", m, s)
}

// FormatRelativeTime returns a human-friendly relative time string (e.g. "2m ago", "1h ago", "in 7 days").
func FormatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	now := time.Now()
	diff := now.Sub(t)
	if diff < 0 {
		diff = -diff
		if diff < time.Minute {
			return "in a few seconds"
		} else if diff < time.Hour {
			return fmt.Sprintf("in %dm", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			return fmt.Sprintf("in %dh", int(diff.Hours()))
		}
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "in 1 day"
		}
		if days < 30 {
			return fmt.Sprintf("in %d days", days)
		}
		months := days / 30
		if months == 1 {
			return "in 1 month"
		}
		return fmt.Sprintf("in %d months", months)
	}

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	} else if diff < 365*24*time.Hour {
		return fmt.Sprintf("%dmo ago", int(diff.Hours()/(24*30)))
	}
	return fmt.Sprintf("%dy ago", int(diff.Hours()/(24*365)))
}

// FormatRelativeAndDate formats a timestamp with relative time and parenthesized date.
func FormatRelativeAndDate(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return fmt.Sprintf("%s (%s)", FormatRelativeTime(t), t.Format("2006-01-02"))
}

// FormatDurationShort formats a time.Duration concisely (e.g. "1.4s", "850ms", "2m 14s").
func FormatDurationShort(d time.Duration) string {
	if d < 100*time.Millisecond {
		return "<0.1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", m, s)
}

// FormatSpeed formats transfer speeds (e.g. "3.2 MB/s", "450.0 KB/s").
func FormatSpeed(bytesPerSec float64) string {
	if bytesPerSec <= 0 {
		return "0 B/s"
	}
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/(1024*1024))
	}
	return fmt.Sprintf("%.2f GB/s", bytesPerSec/(1024*1024*1024))
}

// FormatChecksum formats a sha256 checksum cleanly (e.g. "sha256:4a5e...67e8").
func FormatChecksum(sha string) string {
	clean := strings.TrimSpace(sha)
	clean = strings.TrimPrefix(clean, "sha256:")
	if len(clean) <= 12 {
		return "sha256:" + clean
	}
	return fmt.Sprintf("sha256:%s...%s", clean[:4], clean[len(clean)-4:])
}

// TruncateString truncates a string with ellipsis if it exceeds maxLen.
func TruncateString(s string, maxLen int) string {
	if maxLen <= 3 {
		return s
	}
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen-3]) + "..."
	}
	return s
}

// TruncateMiddle truncates a string preserving start and end (great for filenames with extensions).
func TruncateMiddle(s string, maxLen int) string {
	if len(s) <= maxLen || maxLen <= 6 {
		return s
	}
	ext := ""
	lastDot := strings.LastIndex(s, ".")
	if lastDot > 0 && len(s)-lastDot <= 6 {
		ext = s[lastDot:]
	}
	remaining := maxLen - 3 - len(ext)
	if remaining < 2 {
		return TruncateString(s, maxLen)
	}
	front := (remaining + 1) / 2
	back := remaining - front
	stem := s[:len(s)-len(ext)]
	return stem[:front] + "..." + stem[len(stem)-back:] + ext
}
