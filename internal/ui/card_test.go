package ui

import (
	"strings"
	"testing"
	"time"
)

func TestRenderDownloadSuccess(t *testing.T) {
	out := RenderDownloadSuccess("test.txt", 1024*1024, "/tmp/test.txt", 500*time.Millisecond)

	if !strings.Contains(out, "Downloaded test.txt") {
		t.Errorf("expected output to contain 'Downloaded test.txt', got: %s", out)
	}
	if !strings.Contains(out, "Destination") {
		t.Errorf("expected output to contain 'Destination', got: %s", out)
	}
	if !strings.Contains(out, "/tmp/test.txt") {
		t.Errorf("expected output to contain '/tmp/test.txt', got: %s", out)
	}
	if !strings.Contains(out, "Size") {
		t.Errorf("expected output to contain 'Size', got: %s", out)
	}
	if !strings.Contains(out, "1.0 MB") {
		t.Errorf("expected output to contain '1.0 MB', got: %s", out)
	}
}

func TestRenderFolderZipSuccess(t *testing.T) {
	out := RenderFolderZipSuccess("/tmp/folder.zip", 2*1024*1024, 750*time.Millisecond)

	if !strings.Contains(out, "Downloaded archive") {
		t.Errorf("expected output to contain 'Downloaded archive', got: %s", out)
	}
	if !strings.Contains(out, "Archive") {
		t.Errorf("expected output to contain 'Archive', got: %s", out)
	}
	if !strings.Contains(out, "/tmp/folder.zip") {
		t.Errorf("expected output to contain '/tmp/folder.zip', got: %s", out)
	}
	if !strings.Contains(out, "Size") {
		t.Errorf("expected output to contain 'Size', got: %s", out)
	}
	if !strings.Contains(out, "2.0 MB") {
		t.Errorf("expected output to contain '2.0 MB', got: %s", out)
	}
}

func TestResolveWebRoot(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "https://ohiofiles.cloud"},
		{"https://api.ohiofiles.cloud", "https://ohiofiles.cloud"},
		{"https://api.ohfs.app", "https://ohiofiles.cloud"},
		{"https://ohfs.app", "https://ohiofiles.cloud"},
		{"http://localhost:3000/api", "http://localhost:3000"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ResolveWebRoot(tt.input)
			if got != tt.want {
				t.Errorf("ResolveWebRoot(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

