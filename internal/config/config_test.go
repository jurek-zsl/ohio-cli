package config

import (
	"os"
	"testing"
)

func TestResolveApiUrlMigration(t *testing.T) {
	cfg := &Config{
		ApiUrl: "https://api.ohfs.app",
	}

	// Case 1: c.ApiUrl with ohfs.app
	if got := cfg.ResolveApiUrl(); got != "https://api.ohiofiles.cloud" {
		t.Errorf("ResolveApiUrl() with c.ApiUrl = %q, want %q", got, "https://api.ohiofiles.cloud")
	}

	// Case 2: OHIO_API_URL env var with ohfs.app
	_ = os.Setenv("OHIO_API_URL", "https://api.ohfs.app")
	defer os.Unsetenv("OHIO_API_URL")

	if got := cfg.ResolveApiUrl(); got != "https://api.ohiofiles.cloud" {
		t.Errorf("ResolveApiUrl() with OHIO_API_URL = %q, want %q", got, "https://api.ohiofiles.cloud")
	}

	_ = os.Unsetenv("OHIO_API_URL")

	// Case 3: OHFS_API_URL env var with ohfs.app
	_ = os.Setenv("OHFS_API_URL", "https://api.ohfs.app")
	defer os.Unsetenv("OHFS_API_URL")

	if got := cfg.ResolveApiUrl(); got != "https://api.ohiofiles.cloud" {
		t.Errorf("ResolveApiUrl() with OHFS_API_URL = %q, want %q", got, "https://api.ohiofiles.cloud")
	}
}
