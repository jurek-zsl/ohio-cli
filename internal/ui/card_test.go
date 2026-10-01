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
