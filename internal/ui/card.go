package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"ohio/internal/api"
)

var (
	CardLabelStyle       = lipgloss.NewStyle().Foreground(ColorGrayMuted)
	CardValueStyle       = lipgloss.NewStyle().Foreground(ColorWhite)
	CardLinkStyle        = lipgloss.NewStyle().Foreground(ColorCyanBright)
	CardMarkdownStyle    = lipgloss.NewStyle().Foreground(ColorPurplePrimary)
	CardChecksumStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#a1a1aa"))
	CardCheckmarkStyle   = lipgloss.NewStyle().Foreground(ColorEmeraldAccent).Bold(true)
	CardDotStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#52525b"))
	CardDimmedStyle      = lipgloss.NewStyle().Foreground(ColorGrayMuted)
	CardSessionKeyStyle  = lipgloss.NewStyle().Foreground(ColorCyanBright).Bold(true)
)

// FormatKV formats an aligned key-value line for minimalist cards.
func FormatKV(label, value string) string {
	return fmt.Sprintf("  %s %s", CardLabelStyle.Render(fmt.Sprintf("%-12s", label)), value)
}

// ResolveWebRoot converts an API URL into the public frontend web URL.
func ResolveWebRoot(apiURL string) string {
	if apiURL == "" {
		return "https://ohiofiles.cloud"
	}
	u := strings.TrimRight(apiURL, "/")
	u = strings.TrimSuffix(u, "/api")
	if strings.Contains(u, "://api.") {
		u = strings.Replace(u, "://api.", "://", 1)
	}
	if strings.Contains(u, "ohfs.app") {
		u = strings.ReplaceAll(u, "ohfs.app", "ohiofiles.cloud")
	}
	return u
}

// RenderUploadCard renders the modern minimalist upload completion card.
func RenderUploadCard(f *api.FileItem, duration time.Duration, showQR bool) string {
	var b strings.Builder

	// Title line: ✔ Uploaded filename.ext in 1.4s (12.4 MB)
	check := CardCheckmarkStyle.Render("✔")
	title := TitleStyle.Render("Uploaded " + f.Filename)
	dur := CardDimmedStyle.Render("in " + FormatDurationShort(duration))
	size := CardDimmedStyle.Render("(" + FormatBytes(f.GetEffectiveSize()) + ")")
	b.WriteString(fmt.Sprintf("%s %s %s %s\n\n", check, title, dur, size))

	// Link
	linkURL := f.URL
	if linkURL == "" {
		linkURL = fmt.Sprintf("https://ohiofiles.cloud/%s", f.Slug)
	}
	b.WriteString(FormatKV("Link", CardLinkStyle.Render(linkURL)) + "\n")

	// Slug
	b.WriteString(FormatKV("Slug", CardValueStyle.Render(f.Slug)) + "\n")

	// Visibility & Badges
	var badges []string
	if f.IsPublic {
		badges = append(badges, PillPublicStyle.Render("Public"))
	} else {
		badges = append(badges, PillPrivateStyle.Render("Private"))
	}
	if f.IsOneTime {
		badges = append(badges, PillOneTimeStyle.Render("One-Time"))
	}
	if f.PasswordRequired {
		badges = append(badges, PillPasswordStyle.Render("Password Protected"))
	}
	b.WriteString(FormatKV("Visibility", strings.Join(badges, CardDotStyle.Render("  •  "))) + "\n")

	// Location
	b.WriteString(FormatKV("Location", CardValueStyle.Render(api.GetFileLocation(f))) + "\n")

	// Markdown Embed: ![[ohiofile:my-slug]]
	b.WriteString(FormatKV("Markdown", CardMarkdownStyle.Render(fmt.Sprintf("![[ohiofile:%s]]", f.Slug))) + "\n")

	// Checksum
	if f.ChecksumSHA256 != nil && *f.ChecksumSHA256 != "" {
		b.WriteString(FormatKV("Checksum", CardChecksumStyle.Render(FormatChecksum(*f.ChecksumSHA256))) + "\n")
	}

	// Terminal QR Code enclosed in subtle frame if requested
	if showQR {
		qrBox, err := RenderFramedQRCode(linkURL)
		if err == nil {
			b.WriteString("\n" + qrBox + "\n")
		}
	}

	return b.String()
}

// RenderDownloadSuccess formats the download completion output.
func RenderDownloadSuccess(filename string, size int64, destPath string, duration time.Duration) string {
	check := CardCheckmarkStyle.Render("✔")
	title := TitleStyle.Render("Downloaded " + filename)
	sizeText := CardDimmedStyle.Render("(" + FormatBytes(size) + ")")
	durText := CardDimmedStyle.Render("in " + FormatDurationShort(duration))
	return fmt.Sprintf("%s %s %s %s\n%s\n%s",
		check, title, sizeText, durText,
		FormatKV("Destination", CardValueStyle.Render(destPath)),
		FormatKV("Size", CardValueStyle.Render(FormatBytes(size))),
	)
}

// RenderSessionCard renders active or newly created session details.
func RenderSessionCard(sess *api.PublicSession, title string, deviceId, apiUrl string) string {
	var b strings.Builder

	check := CardCheckmarkStyle.Render("✔")
	header := fmt.Sprintf("%s %s %s\n\n", check, TitleStyle.Render(title+":"), CardSessionKeyStyle.Render(sess.Key))
	b.WriteString(header)

	// Nickname
	if sess.Nickname != "" {
		b.WriteString(FormatKV("Nickname", CardValueStyle.Render(sess.Nickname)) + "\n")
	}

	// Profile link
	if sess.ProfileSlug != "" {
		webRoot := ResolveWebRoot(apiUrl)
		profileURL := fmt.Sprintf("%s/profiles/%s", webRoot, sess.ProfileSlug)
		b.WriteString(FormatKV("Profile", CardLinkStyle.Render(profileURL)) + "\n")
	}

	// Files
	if sess.FileCount > 0 {
		b.WriteString(FormatKV("Files", CardValueStyle.Render(fmt.Sprintf("%d files", sess.FileCount))) + "\n")
	}

	// Expires
	if sess.ExpiresAt != nil && !sess.ExpiresAt.IsZero() {
		b.WriteString(FormatKV("Expires", CardDimmedStyle.Render(FormatRelativeAndDate(*sess.ExpiresAt))) + "\n")
	}

	// Device
	if deviceId != "" {
		b.WriteString(FormatKV("Device", CardDimmedStyle.Render(deviceId)) + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// RenderSessionNicknameSuccess renders nickname update confirmation.
func RenderSessionNicknameSuccess(sess *api.PublicSession, apiUrl string) string {
	var b strings.Builder

	check := CardCheckmarkStyle.Render("✔")
	header := fmt.Sprintf("%s %s %s\n\n", check, TitleStyle.Render("Nickname updated:"), CardValueStyle.Render(sess.Nickname))
	b.WriteString(header)

	b.WriteString(FormatKV("Session", CardSessionKeyStyle.Render(sess.Key)) + "\n")
	if sess.ProfileSlug != "" {
		webRoot := ResolveWebRoot(apiUrl)
		profileURL := fmt.Sprintf("%s/profiles/%s", webRoot, sess.ProfileSlug)
		b.WriteString(FormatKV("Profile", CardLinkStyle.Render(profileURL)) + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// RenderShareCard formats share links and embeds with terminal QR code.
func RenderShareCard(f *api.FileItem, directURL string) string {
	var b strings.Builder

	check := CardCheckmarkStyle.Render("✔")
	header := fmt.Sprintf("%s %s\n\n", check, TitleStyle.Render("Share file: "+f.Filename))
	b.WriteString(header)

	linkURL := f.URL
	if linkURL == "" {
		linkURL = fmt.Sprintf("https://ohiofiles.cloud/%s", f.Slug)
	}

	b.WriteString(FormatKV("Link", CardLinkStyle.Render(linkURL)) + "\n")
	if directURL != "" {
		b.WriteString(FormatKV("Direct", CardLinkStyle.Render(directURL)) + "\n")
	}
	b.WriteString(FormatKV("Markdown", CardMarkdownStyle.Render(fmt.Sprintf("![[ohiofile:%s]]", f.Slug))) + "\n")
	b.WriteString(FormatKV("HTML", CardDimmedStyle.Render(fmt.Sprintf("<a href=\"%s\">%s</a>", linkURL, f.Filename))) + "\n")

	qrBox, err := RenderFramedQRCode(linkURL)
	if err == nil {
		b.WriteString("\n" + qrBox + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// RenderFolderCard formats folder creation or inspection cards.
func RenderFolderCard(folder *api.FolderItem, title string) string {
	var b strings.Builder

	check := CardCheckmarkStyle.Render("✔")
	header := fmt.Sprintf("%s %s\n\n", check, TitleStyle.Render(fmt.Sprintf("%s: %s", title, folder.Name)))
	b.WriteString(header)

	b.WriteString(FormatKV("ID", CardDimmedStyle.Render(folder.ID)) + "\n")
	b.WriteString(FormatKV("Slug", CardValueStyle.Render(folder.Slug)) + "\n")

	if folder.IsPublic {
		b.WriteString(FormatKV("Visibility", PillPublicStyle.Render("Public")) + "\n")
	} else {
		b.WriteString(FormatKV("Visibility", PillPrivateStyle.Render("Private")) + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

// RenderFolderZipSuccess formats folder archive download completion.
func RenderFolderZipSuccess(archivePath string, size int64, duration time.Duration) string {
	check := CardCheckmarkStyle.Render("✔")
	title := TitleStyle.Render("Downloaded archive")
	sizeText := CardDimmedStyle.Render("(" + FormatBytes(size) + ")")
	durText := CardDimmedStyle.Render("in " + FormatDurationShort(duration))
	return fmt.Sprintf("%s %s %s %s\n%s\n%s",
		check, title, sizeText, durText,
		FormatKV("Archive", CardValueStyle.Render(archivePath)),
		FormatKV("Size", CardValueStyle.Render(FormatBytes(size))),
	)
}

// RenderDeleteSuccess formats deletion confirmation.
func RenderDeleteSuccess(itemType, idOrSlug string) string {
	check := CardCheckmarkStyle.Render("✔")
	return fmt.Sprintf("%s %s %s", check, TitleStyle.Render(fmt.Sprintf("Deleted %s:", itemType)), CardDimmedStyle.Render(idOrSlug))
}

// RenderInfoCard renders file details for ohio info.
func RenderInfoCard(f *api.FileItem) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("File Details: "+f.Filename) + "\n\n")

	linkURL := f.URL
	if linkURL == "" {
		linkURL = fmt.Sprintf("https://ohiofiles.cloud/%s", f.Slug)
	}

	b.WriteString(FormatKV("Link", CardLinkStyle.Render(linkURL)) + "\n")
	b.WriteString(FormatKV("Slug", CardValueStyle.Render(f.Slug)) + "\n")
	b.WriteString(FormatKV("Location", CardValueStyle.Render(api.GetFileLocation(f))) + "\n")
	b.WriteString(FormatKV("Size", CardValueStyle.Render(fmt.Sprintf("%s (%d B)", FormatBytes(f.GetEffectiveSize()), f.GetEffectiveSize()))) + "\n")
	b.WriteString(FormatKV("MIME Type", CardDimmedStyle.Render(f.MimeType)) + "\n")

	var badges []string
	if f.IsPublic {
		badges = append(badges, PillPublicStyle.Render("Public"))
	} else {
		badges = append(badges, PillPrivateStyle.Render("Private"))
	}
	if f.IsOneTime {
		badges = append(badges, PillOneTimeStyle.Render("One-Time"))
	}
	if f.PasswordRequired {
		badges = append(badges, PillPasswordStyle.Render("Password Protected"))
	}
	b.WriteString(FormatKV("Visibility", strings.Join(badges, CardDotStyle.Render("  •  "))) + "\n")

	uploadedVal := fmt.Sprintf("%s %s", CardValueStyle.Render(FormatRelativeTime(f.UploadedAt)), CardDimmedStyle.Render("("+f.UploadedAt.Format("2006-01-02 15:04:05")+")"))
	b.WriteString(FormatKV("Uploaded", uploadedVal) + "\n")
	b.WriteString(FormatKV("Downloads", CardValueStyle.Render(fmt.Sprintf("%d", f.Downloads))) + "\n")

	if f.ChecksumSHA256 != nil && *f.ChecksumSHA256 != "" {
		b.WriteString(FormatKV("Checksum", CardChecksumStyle.Render(FormatChecksum(*f.ChecksumSHA256))) + "\n")
	}

	qrBox, err := RenderFramedQRCode(linkURL)
	if err == nil {
		b.WriteString("\n" + qrBox + "\n")
	}

	return strings.TrimRight(b.String(), "\n")
}
