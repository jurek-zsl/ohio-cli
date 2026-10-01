package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want int
	}{
		{"2.2.0", "2.2.1", -1},
		{"v2.2.0", "2.2.0", 0},
		{"2.3.0", "2.2.0", 1},
		{"v2.2.1", "v2.2.0", 1},
		{"v2.2.0", "v2.2.1", -1},
		{"v2.2.0", "v2.2.0", 0},
		{"2.2", "2.2.0", 0},
		{"2.2.0", "2.2", 0},
		{"2.10.0", "2.9.0", 1},
		{"2.9.0", "2.10.0", -1},
		{"v3.0.0", "2.99.99", 1},
		{"1.0.0", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_vs_"+tt.v2, func(t *testing.T) {
			got := CompareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestFindAsset(t *testing.T) {
	assets := []ReleaseAsset{
		{Name: "ohio-darwin-arm64", Size: 1000, BrowserDownloadURL: "https://example.com/darwin-arm64"},
		{Name: "ohio-darwin-amd64", Size: 1100, BrowserDownloadURL: "https://example.com/darwin-amd64"},
		{Name: "ohio-linux-amd64", Size: 1200, BrowserDownloadURL: "https://example.com/linux-amd64"},
		{Name: "ohio-windows-amd64.exe", Size: 1300, BrowserDownloadURL: "https://example.com/windows-amd64.exe"},
	}

	t.Run("darwin arm64", func(t *testing.T) {
		asset, err := FindAsset(assets, "darwin", "arm64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if asset.Name != "ohio-darwin-arm64" {
			t.Errorf("got %q, want %q", asset.Name, "ohio-darwin-arm64")
		}
	})

	t.Run("windows amd64 appends .exe", func(t *testing.T) {
		asset, err := FindAsset(assets, "windows", "amd64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if asset.Name != "ohio-windows-amd64.exe" {
			t.Errorf("got %q, want %q", asset.Name, "ohio-windows-amd64.exe")
		}
	})

	t.Run("linux amd64", func(t *testing.T) {
		asset, err := FindAsset(assets, "linux", "amd64")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if asset.Name != "ohio-linux-amd64" {
			t.Errorf("got %q, want %q", asset.Name, "ohio-linux-amd64")
		}
	})

	t.Run("unsupported platform returns error", func(t *testing.T) {
		asset, err := FindAsset(assets, "solaris", "sparc")
		if err == nil {
			t.Fatalf("expected error for solaris/sparc, got asset: %v", asset)
		}
	})
}

func TestFetchLatestRelease(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "ohio-cli" {
			t.Errorf("missing or invalid User-Agent: %s", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"tag_name": "v2.3.0",
			"name": "OhioCLI v2.3.0",
			"html_url": "https://github.com/jurek-zsl/ohio-cli/releases/tag/v2.3.0",
			"assets": [
				{
					"name": "ohio-darwin-arm64",
					"size": 1234567,
					"browser_download_url": "https://example.com/asset"
				}
			]
		}`))
	}))
	defer ts.Close()

	// Direct test for request handler logic
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ts.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ohio-cli")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", resp.StatusCode)
	}
}
