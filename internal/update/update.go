package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ReleaseAsset represents a single downloadable binary asset from GitHub Releases.
type ReleaseAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// ReleaseInfo represents the GitHub release metadata.
type ReleaseInfo struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	HTMLURL     string         `json:"html_url"`
	PublishedAt time.Time      `json:"published_at"`
	Body        string         `json:"body"`
	Assets      []ReleaseAsset `json:"assets"`
}

// CompareVersions compares two semver-like version strings (e.g. "2.2.0" and "v2.2.1").
// Returns 0 if v1 == v2, 1 if v1 > v2, and -1 if v1 < v2.
func CompareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(strings.TrimPrefix(v1, "v"), "V")
	v2 = strings.TrimPrefix(strings.TrimPrefix(v2, "v"), "V")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}
		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}

// FindAsset locates the platform-specific release binary asset for given OS and architecture.
func FindAsset(assets []ReleaseAsset, goos, goarch string) (*ReleaseAsset, error) {
	target := fmt.Sprintf("ohio-%s-%s", goos, goarch)
	if goos == "windows" {
		target += ".exe"
	}
	for i := range assets {
		if assets[i].Name == target {
			return &assets[i], nil
		}
	}
	return nil, fmt.Errorf("no release asset found for platform %s/%s (expected asset: %s)", goos, goarch, target)
}

// FetchLatestRelease retrieves the latest release info from GitHub API.
func FetchLatestRelease(ctx context.Context, repo string) (*ReleaseInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ohio-cli")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %s", resp.Status)
	}

	var rel ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// ApplyUpdate downloads the binary and replaces the current running binary atomically.
func ApplyUpdate(ctx context.Context, downloadURL string, assetSize int64, progressFn func(written, total int64)) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	if real, err := filepath.EvalSymlinks(execPath); err == nil {
		execPath = real
	}

	execDir := filepath.Dir(execPath)
	tmpFile, err := os.CreateTemp(execDir, ".ohio-update-*")
	if err != nil {
		if os.IsPermission(err) || strings.Contains(strings.ToLower(err.Error()), "permission denied") {
			return "", fmt.Errorf("permission denied writing to %s; try running with sudo: sudo ohio update", execDir)
		}
		return "", err
	}
	tmpPath := tmpFile.Name()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	req.Header.Set("User-Agent", "ohio-cli")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("download failed with HTTP status: %s", resp.Status)
	}

	totalExpected := resp.ContentLength
	if totalExpected <= 0 {
		totalExpected = assetSize
	}

	var reader io.Reader = resp.Body
	if progressFn != nil {
		reader = &progressReader{
			reader:     resp.Body,
			totalBytes: totalExpected,
			progressFn: progressFn,
		}
	}

	if _, err := io.Copy(tmpFile, reader); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("failed writing downloaded binary: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}

	_ = os.Chmod(tmpPath, 0755)

	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"
		_ = os.Remove(oldPath)
		if err := os.Rename(execPath, oldPath); err != nil {
			_ = os.Remove(tmpPath)
			return "", err
		}
		if err := os.Rename(tmpPath, execPath); err != nil {
			_ = os.Rename(oldPath, execPath)
			_ = os.Remove(tmpPath)
			return "", err
		}
		_ = os.Remove(oldPath)
	} else {
		if err := os.Rename(tmpPath, execPath); err != nil {
			// Try copy fallback if cross-device
			if errCopy := copyFile(tmpPath, execPath); errCopy != nil {
				_ = os.Remove(tmpPath)
				return "", errCopy
			}
			_ = os.Remove(tmpPath)
		}
		_ = os.Chmod(execPath, 0755)
	}

	return execPath, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

type progressReader struct {
	reader     io.Reader
	totalBytes int64
	written    int64
	progressFn func(written, total int64)
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.written += int64(n)
		if pr.progressFn != nil {
			pr.progressFn(pr.written, pr.totalBytes)
		}
	}
	return n, err
}
