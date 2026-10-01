package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"ohio/internal/api"
	"ohio/internal/config"
)

func TestViewportWindowing(t *testing.T) {
	cfg := &config.Config{
		SessionKey: "test-session-key",
		Nickname:   "Tester",
		DeviceId:   "device-123",
	}
	app := NewAppModel(cfg, nil)

	// Populate 50 dummy files
	for i := 1; i <= 50; i++ {
		app.files = append(app.files, api.FileItem{
			Slug:        "file-slug",
			Filename:    "test_file.txt",
			Size:        1024,
			IsPublic:    true,
			UploadedAt:  time.Now(),
			Downloads:   5,
		})
		app.feed = append(app.feed, api.FileItem{
			Slug:        "feed-slug",
			Filename:    "feed_file.txt",
			Size:        2048,
			IsPublic:    true,
			UploadedAt:  time.Now(),
			Downloads:   10,
		})
	}

	app.width = 100
	app.height = 24

	// Test Files View windowing
	bodyHeight := 18 // 24 - 6
	expectedVisibleRows := 11 // max(5, 18-7)

	renderedFiles := app.renderFilesView(bodyHeight)
	filesLines := strings.Split(renderedFiles, "\n")
	// Count data rows in rendered output (starts with "▸" or "  ")
	rowCount := 0
	for _, l := range filesLines {
		if strings.Contains(l, "test_file.txt") {
			rowCount++
		}
	}
	if rowCount > expectedVisibleRows {
		t.Errorf("Expected at most %d rows in files view, got %d", expectedVisibleRows, rowCount)
	}

	// Test Feed View windowing
	renderedFeed := app.renderFeedView(bodyHeight)
	feedLines := strings.Split(renderedFeed, "\n")
	feedRowCount := 0
	for _, l := range feedLines {
		if strings.Contains(l, "feed_file.txt") {
			feedRowCount++
		}
	}
	if feedRowCount > expectedVisibleRows {
		t.Errorf("Expected at most %d rows in feed view, got %d", expectedVisibleRows, feedRowCount)
	}
}

func TestTabHeightsStandardTerminal(t *testing.T) {
	cfg := &config.Config{
		SessionKey: "test-session-key",
		Nickname:   "Tester",
		DeviceId:   "device-123",
	}
	app := NewAppModel(cfg, nil)
	app.width = 100
	app.height = 24

	// Populate feed items
	for i := 1; i <= 20; i++ {
		app.feed = append(app.feed, api.FileItem{
			Slug:       "slug-item",
			Filename:   "file.png",
			Size:       5000,
			UploadedAt: time.Now(),
		})
	}

	// Test Help View Height
	app.activeTab = tabHelp
	helpView := app.View()
	helpLines := strings.Split(helpView, "\n")
	if len(helpLines) > 30 {
		t.Errorf("Help view too tall for standard terminal: %d lines", len(helpLines))
	}

	// Test Feed View Height
	app.activeTab = tabFeed
	feedView := app.View()
	feedLines := strings.Split(feedView, "\n")
	if len(feedLines) > 30 {
		t.Errorf("Feed view too tall for standard terminal: %d lines", len(feedLines))
	}

	// Test Session View Alignment
	app.activeTab = tabSession
	sessionView := app.renderSessionView()
	if !strings.Contains(sessionView, "[n]  New Memorable Key") {
		t.Errorf("Expected aligned '[n]  New Memorable Key' action button")
	}
	if !strings.Contains(sessionView, "[s]  New Secure Key") {
		t.Errorf("Expected aligned '[s]  New Secure Key' action button")
	}
}

func TestTabNavigation(t *testing.T) {
	cfg := &config.Config{SessionKey: "sess-123", Nickname: "Tester"}
	app := NewAppModel(cfg, nil)

	tabsRendered := app.renderTabs()
	for _, expected := range []string{"1 Files", "2 Folders", "3 Upload", "4 Feed", "5 Session", "6 Help"} {
		if !strings.Contains(tabsRendered, expected) {
			t.Errorf("Expected tabs to contain %q, got: %s", expected, tabsRendered)
		}
	}

	// Test numerical key jumps
	keys := []struct {
		key      string
		expected tabIndex
	}{
		{"1", tabFiles},
		{"2", tabFolders},
		{"3", tabUpload},
		{"4", tabFeed},
		{"5", tabSession},
		{"6", tabHelp},
	}

	for _, k := range keys {
		model, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k.key)})
		updated := model.(AppModel)
		if updated.activeTab != k.expected {
			t.Errorf("Key %q: expected tab %d, got %d", k.key, k.expected, updated.activeTab)
		}
	}

	// Test tab and shift+tab cycle modulo 6
	app.activeTab = tabHelp // tab index 5
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.(AppModel).activeTab != tabFiles {
		t.Errorf("Tab from help: expected tabFiles (0), got %d", m.(AppModel).activeTab)
	}

	app.activeTab = tabFiles // tab index 0
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.(AppModel).activeTab != tabHelp {
		t.Errorf("Shift+Tab from files: expected tabHelp (5), got %d", m.(AppModel).activeTab)
	}
}

func TestFoldersTab(t *testing.T) {
	cfg := &config.Config{SessionKey: "sess-123", Nickname: "Tester"}
	app := NewAppModel(cfg, nil)
	app.width = 100
	app.height = 24

	app.folders = []api.FolderItem{
		{
			ID:        "fld-1",
			Name:      "Documents",
			Slug:      "docs-slug",
			IsPublic:  false,
			CreatedAt: time.Now(),
		},
		{
			ID:        "fld-2",
			Name:      "PublicPhotos",
			Slug:      "photos-slug",
			IsPublic:  true,
			CreatedAt: time.Now(),
		},
	}

	// Test renderFoldersView
	view := app.renderFoldersView(18)
	for _, col := range []string{"NAME", "SLUG", "VISIBILITY", "CREATED"} {
		if !strings.Contains(view, col) {
			t.Errorf("Expected folders view to contain column %q", col)
		}
	}
	if !strings.Contains(view, "Documents") || !strings.Contains(view, "docs-slug") {
		t.Errorf("Expected folders view to contain folder Documents and docs-slug")
	}

	// Test folder creation input display
	app.folderCreating = true
	app.folderNameInput.SetValue("NewProject")
	creatingView := app.renderFoldersView(18)
	if !strings.Contains(creatingView, "NewProject") {
		t.Errorf("Expected folder name input view to be rendered while folderCreating is true")
	}

	// Test updateFoldersTab keys
	app.activeTab = tabFolders
	app.folderCreating = false

	// Press 'n' to trigger create mode
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	updated := m.(AppModel)
	if !updated.folderCreating {
		t.Errorf("Expected 'n' to activate folderCreating")
	}

	// Press 'esc' to cancel create mode
	m, _ = updated.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated = m.(AppModel)
	if updated.folderCreating {
		t.Errorf("Expected 'esc' to cancel folderCreating")
	}

	// Press 'Enter' to filter files by selected folder
	app.foldersCursor = 0
	m, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated = m.(AppModel)
	if updated.activeTab != tabFiles {
		t.Errorf("Expected 'enter' on folder to switch to tabFiles, got %d", updated.activeTab)
	}
	if updated.folderFilterSlug != "docs-slug" {
		t.Errorf("Expected folderFilterSlug to be 'docs-slug', got %q", updated.folderFilterSlug)
	}
	if updated.folderFilterName != "Documents" {
		t.Errorf("Expected folderFilterName to be 'Documents', got %q", updated.folderFilterName)
	}
}

func TestFilesLocationAndFolderFilter(t *testing.T) {
	cfg := &config.Config{SessionKey: "sess-123", Nickname: "Tester"}
	app := NewAppModel(cfg, nil)
	app.width = 100
	app.height = 24

	folderName := "Projects"
	app.files = []api.FileItem{
		{
			Slug:       "slug-root",
			Filename:   "root.txt",
			Size:       1024,
			IsPublic:   true,
			UploadedAt: time.Now(),
		},
		{
			Slug:       "slug-nested",
			Filename:   "nested.txt",
			Size:       2048,
			FolderName: &folderName,
			IsPublic:   false,
			UploadedAt: time.Now(),
		},
	}

	// Test LOCATION column in header and row
	filesView := app.renderFilesView(18)
	if !strings.Contains(filesView, "LOCATION") {
		t.Errorf("Expected LOCATION column in files view header")
	}
	if !strings.Contains(filesView, "📁 Projects") {
		t.Errorf("Expected '📁 Projects' from api.GetFileLocation in rendered files view")
	}

	// Test folder filter display in files view
	app.folderFilterSlug = "proj-slug"
	app.folderFilterName = "Projects"
	filteredView := app.renderFilesView(18)
	expectedFilterBadge := "Filter: 📁 Projects (Esc to clear)"
	if !strings.Contains(filteredView, expectedFilterBadge) {
		t.Errorf("Expected files view to show %q, got: %s", expectedFilterBadge, filteredView)
	}

	// Test 'esc' clears folder filter in updateFilesTab
	app.activeTab = tabFiles
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	cleared := m.(AppModel)
	if cleared.folderFilterSlug != "" || cleared.folderFilterName != "" {
		t.Errorf("Expected 'esc' in files tab to clear folder filter, got slug=%q name=%q", cleared.folderFilterSlug, cleared.folderFilterName)
	}
}

func TestUploadProgress(t *testing.T) {
	cfg := &config.Config{SessionKey: "sess-123", Nickname: "Tester"}
	app := NewAppModel(cfg, nil)
	app.width = 100
	app.height = 24

	// When not uploading
	normalView := app.renderUploadView()
	if !strings.Contains(normalView, "[ Upload File (Enter) ]") {
		t.Errorf("Expected normal upload view to have upload button")
	}

	// When uploading
	app.uploading = true
	app.uploadProgressPct = 0.45
	app.uploadProgressText = "4.5 MB / 10.0 MB (45.0%) • 1.5 MB/s"

	uploadView := app.renderUploadView()
	if !strings.Contains(uploadView, "45.0%") {
		t.Errorf("Expected upload view to render percentage 45.0%%")
	}
	if !strings.Contains(uploadView, "4.5 MB / 10.0 MB") {
		t.Errorf("Expected upload view to render byte progress")
	}
	if !strings.Contains(uploadView, "1.5 MB/s") {
		t.Errorf("Expected upload view to render speed")
	}

	// Test uploadProgressMsg in Update
	ch := make(chan tea.Msg, 5)
	app.uploadChan = ch
	m, cmd := app.Update(uploadProgressMsg{
		written: 5000,
		total:   10000,
		pct:     0.5,
		speed:   "2.0 MB/s",
	})
	updated := m.(AppModel)
	if updated.uploadProgressPct != 0.5 {
		t.Errorf("Expected uploadProgressPct = 0.5, got %f", updated.uploadProgressPct)
	}
	if !strings.Contains(updated.uploadProgressText, "2.0 MB/s") {
		t.Errorf("Expected uploadProgressText to contain speed, got %q", updated.uploadProgressText)
	}
	if cmd == nil {
		t.Errorf("Expected Update(uploadProgressMsg) to return waitForUploadMsg command")
	}
}

