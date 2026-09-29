package tui

import (
	"strings"
	"testing"
	"time"

	"ohfs/internal/api"
	"ohfs/internal/config"
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
