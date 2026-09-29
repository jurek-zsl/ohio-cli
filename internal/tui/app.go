package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ohfs/internal/api"
	"ohfs/internal/config"
	"ohfs/internal/ui"
)

type tabIndex int

const (
	tabFiles tabIndex = iota
	tabUpload
	tabFeed
	tabSession
	tabHelp
)

// Messages
type errMsg error
type statusMsg string
type clearStatusMsg struct{ id int }
type filesLoadedMsg []api.FileItem
type feedLoadedMsg []api.FileItem
type sessionUpdatedMsg *api.PublicSession
type uploadDoneMsg *api.FileItem
type downloadDoneMsg string

// AppModel is the primary Bubble Tea model for the ohfs interactive dashboard.
type AppModel struct {
	cfg        *config.Config
	client     *api.Client
	activeTab  tabIndex
	width      int
	height     int
	statusText string
	statusErr  bool
	statusID   int
	spinner    spinner.Model
	loading    bool

	// Files Explorer Tab
	files            []api.FileItem
	filesCursor      int
	filesScroll      int
	filesFiltering   bool
	filesFilterInput textinput.Model

	// Public Feed Tab
	feed            []api.FileItem
	feedCursor      int
	feedScroll      int
	feedFiltering   bool
	feedFilterInput textinput.Model

	// Quick Upload Tab
	uploadInputs   []textinput.Model
	uploadFocusIdx int
	uploadPublic   bool
	uploadOneTime  bool
	uploading      bool
	uploadProgress string

	// Session Tab
	sessionNicknameInput textinput.Model
	sessionEditingNick   bool

	// QR Share Modal
	modalActive  bool
	modalTitle   string
	modalURL     string
	modalQRCode  string
	modalSubtext string
}

// NewAppModel initializes the TUI model.
func NewAppModel(cfg *config.Config, client *api.Client) AppModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(ui.ColorCyanBright)

	// Setup upload inputs
	filePathInput := textinput.New()
	filePathInput.Placeholder = "/path/to/file.ext"
	filePathInput.Focus()
	filePathInput.CharLimit = 256
	filePathInput.Width = 36

	slugInput := textinput.New()
	slugInput.Placeholder = "custom-slug (optional)"
	slugInput.CharLimit = 64
	slugInput.Width = 30

	pwdInput := textinput.New()
	pwdInput.Placeholder = "password (optional)"
	pwdInput.EchoMode = textinput.EchoPassword
	pwdInput.EchoCharacter = '•'
	pwdInput.CharLimit = 64
	pwdInput.Width = 30

	nickInput := textinput.New()
	nickInput.Placeholder = "NewNickname"
	nickInput.CharLimit = 24
	nickInput.Width = 24

	filesFilter := textinput.New()
	filesFilter.Placeholder = "search files..."
	filesFilter.CharLimit = 40
	filesFilter.Width = 24

	feedFilter := textinput.New()
	feedFilter.Placeholder = "search public feed..."
	feedFilter.CharLimit = 40
	feedFilter.Width = 24

	return AppModel{
		cfg:                  cfg,
		client:               client,
		activeTab:            tabFiles,
		spinner:              sp,
		uploadInputs:         []textinput.Model{filePathInput, slugInput, pwdInput},
		uploadPublic:         false,
		uploadOneTime:        false,
		sessionNicknameInput: nickInput,
		filesFilterInput:     filesFilter,
		feedFilterInput:      feedFilter,
	}
}

// Init triggers initial commands on startup.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadFilesCmd(),
		m.loadFeedCmd(),
	)
}

func (m *AppModel) setStatus(text string, isErr bool) tea.Cmd {
	m.statusText = text
	m.statusErr = isErr
	m.statusID++
	id := m.statusID
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return clearStatusMsg{id: id}
	})
}

func (m AppModel) loadFilesCmd() tea.Cmd {
	return func() tea.Msg {
		if m.cfg.SessionKey == "" {
			return filesLoadedMsg([]api.FileItem{})
		}
		res, err := m.client.ListFiles(m.cfg.SessionKey, "", false, 1, 100)
		if err != nil {
			return errMsg(err)
		}
		return filesLoadedMsg(res.Files)
	}
}

func (m AppModel) loadFeedCmd() tea.Cmd {
	return func() tea.Msg {
		res, err := m.client.ListFiles("", "", true, 1, 50)
		if err != nil {
			return errMsg(err)
		}
		return feedLoadedMsg(res.Files)
	}
}

func (m AppModel) updateSessionCmd(keyType string) tea.Cmd {
	return func() tea.Msg {
		sess, err := m.client.CreatePublicSession(keyType)
		if err != nil {
			return errMsg(err)
		}
		_ = m.cfg.SetActiveSession(sess.Key, sess.Nickname, sess.ProfileSlug)
		m.client.SessionKey = sess.Key
		return sessionUpdatedMsg(sess)
	}
}

func (m AppModel) getFilteredFiles() []api.FileItem {
	query := strings.ToLower(strings.TrimSpace(m.filesFilterInput.Value()))
	if query == "" {
		return m.files
	}
	var out []api.FileItem
	for _, f := range m.files {
		if strings.Contains(strings.ToLower(f.Filename), query) || strings.Contains(strings.ToLower(f.Slug), query) {
			out = append(out, f)
		}
	}
	return out
}

func (m AppModel) getFilteredFeed() []api.FileItem {
	query := strings.ToLower(strings.TrimSpace(m.feedFilterInput.Value()))
	if query == "" {
		return m.feed
	}
	var out []api.FileItem
	for _, f := range m.feed {
		if strings.Contains(strings.ToLower(f.Filename), query) || strings.Contains(strings.ToLower(f.Slug), query) {
			out = append(out, f)
		}
	}
	return out
}

func (m *AppModel) syncFilesScroll(visibleRows int, totalItems int) {
	if totalItems == 0 {
		m.filesCursor = 0
		m.filesScroll = 0
		return
	}
	if m.filesCursor < 0 {
		m.filesCursor = 0
	}
	if m.filesCursor >= totalItems {
		m.filesCursor = totalItems - 1
	}
	if m.filesCursor < m.filesScroll {
		m.filesScroll = m.filesCursor
	}
	if m.filesCursor >= m.filesScroll+visibleRows {
		m.filesScroll = m.filesCursor - visibleRows + 1
	}
	if m.filesScroll+visibleRows > totalItems && totalItems > visibleRows {
		m.filesScroll = totalItems - visibleRows
	}
	if m.filesScroll < 0 {
		m.filesScroll = 0
	}
}

func (m *AppModel) syncFeedScroll(visibleRows int, totalItems int) {
	if totalItems == 0 {
		m.feedCursor = 0
		m.feedScroll = 0
		return
	}
	if m.feedCursor < 0 {
		m.feedCursor = 0
	}
	if m.feedCursor >= totalItems {
		m.feedCursor = totalItems - 1
	}
	if m.feedCursor < m.feedScroll {
		m.feedScroll = m.feedCursor
	}
	if m.feedCursor >= m.feedScroll+visibleRows {
		m.feedScroll = m.feedCursor - visibleRows + 1
	}
	if m.feedScroll+visibleRows > totalItems && totalItems > visibleRows {
		m.feedScroll = totalItems - visibleRows
	}
	if m.feedScroll < 0 {
		m.feedScroll = 0
	}
}

func padCol(content string, width int) string {
	return lipgloss.NewStyle().Width(width).MaxWidth(width).Render(content)
}

// Update handles incoming messages and keybindings.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case clearStatusMsg:
		if msg.id == m.statusID {
			m.statusText = ""
			m.statusErr = false
		}
		return m, nil

	case filesLoadedMsg:
		m.files = msg
		m.loading = false
		bodyHeight := max(10, m.height-6)
		visibleRows := max(5, bodyHeight-7)
		m.syncFilesScroll(visibleRows, len(m.getFilteredFiles()))

	case feedLoadedMsg:
		m.feed = msg
		bodyHeight := max(10, m.height-6)
		visibleRows := max(5, bodyHeight-7)
		m.syncFeedScroll(visibleRows, len(m.getFilteredFeed()))

	case sessionUpdatedMsg:
		cmd := m.setStatus(fmt.Sprintf("Active session: %s (%s)", msg.Key, msg.Nickname), false)
		cmds = append(cmds, cmd, m.loadFilesCmd())

	case uploadDoneMsg:
		m.uploading = false
		statusCmd := m.setStatus(fmt.Sprintf("Uploaded successfully: %s", msg.URL), false)
		m.uploadInputs[0].SetValue("")
		m.uploadInputs[1].SetValue("")
		m.uploadInputs[2].SetValue("")
		cmds = append(cmds, statusCmd, m.loadFilesCmd())

	case downloadDoneMsg:
		cmds = append(cmds, m.setStatus(string(msg), false))

	case statusMsg:
		cmds = append(cmds, m.setStatus(string(msg), false))

	case errMsg:
		m.loading = false
		m.uploading = false
		cmds = append(cmds, m.setStatus(msg.Error(), true))

	case tea.KeyMsg:
		// Modal overlay keys
		if m.modalActive {
			switch msg.String() {
			case "esc", "q", "enter", "space":
				m.modalActive = false
				return m, nil
			}
			return m, nil
		}

		// When search filter is active on files/feed tabs, don't trigger global number switching
		isFiltering := (m.activeTab == tabFiles && m.filesFiltering) || (m.activeTab == tabFeed && m.feedFiltering)
		if !isFiltering {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.activeTab = (m.activeTab + 1) % 5
				return m, nil
			case "shift+tab":
				m.activeTab = (m.activeTab + 4) % 5
				return m, nil
			case "1":
				m.activeTab = tabFiles
				return m, nil
			case "2":
				m.activeTab = tabUpload
				return m, nil
			case "3":
				m.activeTab = tabFeed
				return m, nil
			case "4":
				m.activeTab = tabSession
				return m, nil
			case "5":
				m.activeTab = tabHelp
				return m, nil
			}
		}

		// Tab-specific key handling
		switch m.activeTab {
		case tabFiles:
			return m.updateFilesTab(msg)
		case tabUpload:
			return m.updateUploadTab(msg)
		case tabFeed:
			return m.updateFeedTab(msg)
		case tabSession:
			return m.updateSessionTab(msg)
		case tabHelp:
			if msg.String() == "q" || msg.String() == "esc" {
				return m, tea.Quit
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) updateFilesTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.getFilteredFiles()
	bodyHeight := max(10, m.height-6)
	visibleRows := max(5, bodyHeight-7)

	if m.filesFiltering {
		switch msg.String() {
		case "esc":
			m.filesFiltering = false
			m.filesFilterInput.SetValue("")
			m.filesFilterInput.Blur()
			m.syncFilesScroll(visibleRows, len(m.getFilteredFiles()))
			return m, nil
		case "enter":
			m.filesFiltering = false
			m.filesFilterInput.Blur()
			m.syncFilesScroll(visibleRows, len(filtered))
			return m, nil
		case "up":
			if m.filesCursor > 0 {
				m.filesCursor--
				m.syncFilesScroll(visibleRows, len(filtered))
			}
			return m, nil
		case "down":
			if m.filesCursor < len(filtered)-1 {
				m.filesCursor++
				m.syncFilesScroll(visibleRows, len(filtered))
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.filesFilterInput, cmd = m.filesFilterInput.Update(msg)
			filteredAfter := m.getFilteredFiles()
			m.syncFilesScroll(visibleRows, len(filteredAfter))
			return m, cmd
		}
	}

	switch msg.String() {
	case "/":
		m.filesFiltering = true
		m.filesFilterInput.Focus()
		return m, nil
	case "up", "k":
		if m.filesCursor > 0 {
			m.filesCursor--
			m.syncFilesScroll(visibleRows, len(filtered))
		}
	case "down", "j":
		if m.filesCursor < len(filtered)-1 {
			m.filesCursor++
			m.syncFilesScroll(visibleRows, len(filtered))
		}
	case "g":
		m.filesCursor = 0
		m.syncFilesScroll(visibleRows, len(filtered))
	case "G":
		if len(filtered) > 0 {
			m.filesCursor = len(filtered) - 1
			m.syncFilesScroll(visibleRows, len(filtered))
		}
	case "r":
		m.loading = true
		cmd := m.setStatus("Refreshing files...", false)
		return m, tea.Batch(cmd, m.loadFilesCmd())
	case "u":
		m.activeTab = tabUpload
	case "s", "enter":
		if len(filtered) > 0 && m.filesCursor < len(filtered) {
			f := filtered[m.filesCursor]
			qr, err := ui.GenerateTerminalQRCode(f.URL)
			if err == nil {
				m.modalActive = true
				m.modalTitle = fmt.Sprintf("Share: %s", f.Filename)
				m.modalURL = f.URL
				m.modalQRCode = qr
				m.modalSubtext = fmt.Sprintf("Size: %s  •  Downloads: %d  •  Slug: %s", ui.FormatBytes(f.GetEffectiveSize()), f.Downloads, f.Slug)
			}
		}
	case "d":
		if len(filtered) > 0 && m.filesCursor < len(filtered) {
			f := filtered[m.filesCursor]
			statusCmd := m.setStatus(fmt.Sprintf("Downloading %s...", f.Filename), false)
			return m, tea.Batch(statusCmd, func() tea.Msg {
				err := m.client.DownloadFile(f.Slug, f.Filename, "", m.cfg.SessionKey, nil)
				if err != nil {
					return errMsg(err)
				}
				return downloadDoneMsg(fmt.Sprintf("Downloaded %s to current directory", f.Filename))
			})
		}
	case "x":
		if len(filtered) > 0 && m.filesCursor < len(filtered) {
			f := filtered[m.filesCursor]
			statusCmd := m.setStatus(fmt.Sprintf("Deleting %s...", f.Filename), false)
			return m, tea.Batch(statusCmd, func() tea.Msg {
				err := m.client.DeleteFile(f.Slug, m.cfg.SessionKey)
				if err != nil {
					return errMsg(err)
				}
				return statusMsg(fmt.Sprintf("Deleted %s", f.Filename))
			})
		}
	case "esc":
		if m.filesFilterInput.Value() != "" {
			m.filesFilterInput.SetValue("")
			m.syncFilesScroll(visibleRows, len(m.getFilteredFiles()))
			return m, nil
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m AppModel) updateFeedTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	filtered := m.getFilteredFeed()
	bodyHeight := max(10, m.height-6)
	visibleRows := max(5, bodyHeight-7)

	if m.feedFiltering {
		switch msg.String() {
		case "esc":
			m.feedFiltering = false
			m.feedFilterInput.SetValue("")
			m.feedFilterInput.Blur()
			m.syncFeedScroll(visibleRows, len(m.getFilteredFeed()))
			return m, nil
		case "enter":
			m.feedFiltering = false
			m.feedFilterInput.Blur()
			m.syncFeedScroll(visibleRows, len(filtered))
			return m, nil
		case "up":
			if m.feedCursor > 0 {
				m.feedCursor--
				m.syncFeedScroll(visibleRows, len(filtered))
			}
			return m, nil
		case "down":
			if m.feedCursor < len(filtered)-1 {
				m.feedCursor++
				m.syncFeedScroll(visibleRows, len(filtered))
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.feedFilterInput, cmd = m.feedFilterInput.Update(msg)
			filteredAfter := m.getFilteredFeed()
			m.syncFeedScroll(visibleRows, len(filteredAfter))
			return m, cmd
		}
	}

	switch msg.String() {
	case "/":
		m.feedFiltering = true
		m.feedFilterInput.Focus()
		return m, nil
	case "up", "k":
		if m.feedCursor > 0 {
			m.feedCursor--
			m.syncFeedScroll(visibleRows, len(filtered))
		}
	case "down", "j":
		if m.feedCursor < len(filtered)-1 {
			m.feedCursor++
			m.syncFeedScroll(visibleRows, len(filtered))
		}
	case "g":
		m.feedCursor = 0
		m.syncFeedScroll(visibleRows, len(filtered))
	case "G":
		if len(filtered) > 0 {
			m.feedCursor = len(filtered) - 1
			m.syncFeedScroll(visibleRows, len(filtered))
		}
	case "r":
		m.loading = true
		cmd := m.setStatus("Refreshing public feed...", false)
		return m, tea.Batch(cmd, m.loadFeedCmd())
	case "s", "enter":
		if len(filtered) > 0 && m.feedCursor < len(filtered) {
			f := filtered[m.feedCursor]
			qr, err := ui.GenerateTerminalQRCode(f.URL)
			if err == nil {
				m.modalActive = true
				m.modalTitle = fmt.Sprintf("Public File: %s", f.Filename)
				m.modalURL = f.URL
				m.modalQRCode = qr
				m.modalSubtext = fmt.Sprintf("Size: %s  •  Downloads: %d  •  Slug: %s", ui.FormatBytes(f.GetEffectiveSize()), f.Downloads, f.Slug)
			}
		}
	case "d":
		if len(filtered) > 0 && m.feedCursor < len(filtered) {
			f := filtered[m.feedCursor]
			statusCmd := m.setStatus(fmt.Sprintf("Downloading %s...", f.Filename), false)
			return m, tea.Batch(statusCmd, func() tea.Msg {
				err := m.client.DownloadFile(f.Slug, f.Filename, "", "", nil)
				if err != nil {
					return errMsg(err)
				}
				return downloadDoneMsg(fmt.Sprintf("Downloaded %s to current directory", f.Filename))
			})
		}
	case "esc":
		if m.feedFilterInput.Value() != "" {
			m.feedFilterInput.SetValue("")
			m.syncFeedScroll(visibleRows, len(m.getFilteredFeed()))
			return m, nil
		}
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m AppModel) updateUploadTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.String() {
	case "up":
		if m.uploadFocusIdx > 0 {
			m.uploadFocusIdx--
			m.updateUploadInputsFocus()
		}
	case "down", "tab":
		if m.uploadFocusIdx < 5 {
			m.uploadFocusIdx++
			m.updateUploadInputsFocus()
		}
	case "shift+tab":
		if m.uploadFocusIdx > 0 {
			m.uploadFocusIdx--
			m.updateUploadInputsFocus()
		}
	case "p":
		if m.uploadFocusIdx >= 3 {
			m.uploadPublic = !m.uploadPublic
			return m, nil
		}
	case "o":
		if m.uploadFocusIdx >= 3 {
			m.uploadOneTime = !m.uploadOneTime
			return m, nil
		}
	case "enter":
		if m.uploadFocusIdx == 3 {
			m.uploadPublic = !m.uploadPublic
			return m, nil
		}
		if m.uploadFocusIdx == 4 {
			m.uploadOneTime = !m.uploadOneTime
			return m, nil
		}
		if m.uploadFocusIdx == 5 || (m.uploadFocusIdx == 0 && strings.TrimSpace(m.uploadInputs[0].Value()) != "") {
			filePath := strings.TrimSpace(m.uploadInputs[0].Value())
			if filePath == "" {
				cmd := m.setStatus("Please enter a valid file path", true)
				return m, cmd
			}
			if _, err := os.Stat(filePath); err != nil {
				cmd := m.setStatus(fmt.Sprintf("File not found: %s", filePath), true)
				return m, cmd
			}

			m.uploading = true
			statusCmd := m.setStatus("Uploading file to OhioFiles...", false)

			opts := api.UploadOptions{
				IsPublic:   m.uploadPublic,
				IsOneTime:  m.uploadOneTime,
				CustomSlug: strings.TrimSpace(m.uploadInputs[1].Value()),
				Password:   strings.TrimSpace(m.uploadInputs[2].Value()),
				SessionKey: m.cfg.SessionKey,
				DeviceId:   m.cfg.DeviceId,
			}

			uploadCmd := func() tea.Msg {
				file, err := m.client.Upload(filePath, opts, nil, nil)
				if err != nil {
					return errMsg(err)
				}
				return uploadDoneMsg(file)
			}

			return m, tea.Batch(statusCmd, uploadCmd)
		}
	}

	// Update active text input
	if m.uploadFocusIdx >= 0 && m.uploadFocusIdx < len(m.uploadInputs) {
		var cmd tea.Cmd
		m.uploadInputs[m.uploadFocusIdx], cmd = m.uploadInputs[m.uploadFocusIdx].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) updateUploadInputsFocus() {
	for i := range m.uploadInputs {
		if i == m.uploadFocusIdx {
			m.uploadInputs[i].Focus()
		} else {
			m.uploadInputs[i].Blur()
		}
	}
}

func (m AppModel) updateSessionTab(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.sessionEditingNick {
		switch msg.String() {
		case "esc":
			m.sessionEditingNick = false
			return m, nil
		case "enter":
			newNick := strings.TrimSpace(m.sessionNicknameInput.Value())
			if newNick != "" {
				m.sessionEditingNick = false
				statusCmd := m.setStatus(fmt.Sprintf("Updating nickname to %s...", newNick), false)
				return m, tea.Batch(statusCmd, func() tea.Msg {
					sess, err := m.client.UpdateNickname(m.cfg.SessionKey, newNick)
					if err != nil {
						return errMsg(err)
					}
					_ = m.cfg.SetActiveSession(sess.Key, sess.Nickname, sess.ProfileSlug)
					return sessionUpdatedMsg(sess)
				})
			}
		default:
			var cmd tea.Cmd
			m.sessionNicknameInput, cmd = m.sessionNicknameInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	switch msg.String() {
	case "n":
		cmd := m.setStatus("Generating new memorable session...", false)
		return m, tea.Batch(cmd, m.updateSessionCmd("memorable"))
	case "s":
		cmd := m.setStatus("Generating new secure session...", false)
		return m, tea.Batch(cmd, m.updateSessionCmd("secure"))
	case "e":
		m.sessionEditingNick = true
		m.sessionNicknameInput.Focus()
		m.sessionNicknameInput.SetValue(m.cfg.Nickname)
		return m, nil
	case "c":
		if m.cfg.SessionKey != "" {
			_ = clipboard.WriteAll(m.cfg.SessionKey)
			cmd := m.setStatus(fmt.Sprintf("Session key '%s' copied to clipboard!", m.cfg.SessionKey), false)
			return m, cmd
		}
		cmd := m.setStatus("No active session key to copy", true)
		return m, cmd
	case "q":
		return m, tea.Quit
	}

	return m, nil
}

// View renders the terminal dashboard.
func (m AppModel) View() string {
	if m.width <= 0 {
		m.width = 100
	}
	if m.height <= 0 {
		m.height = 30
	}

	bodyHeight := max(10, m.height-6)

	var b strings.Builder

	// Top Bar: Header logo, active session pill, current API url
	topBar := m.renderTopBar()
	b.WriteString(topBar)
	b.WriteString("\n\n")

	// Tabs Bar
	tabsBar := m.renderTabs()
	b.WriteString(tabsBar)
	b.WriteString("\n\n")

	// Content Body
	var body string
	if m.modalActive {
		body = m.renderModal()
	} else {
		switch m.activeTab {
		case tabFiles:
			body = m.renderFilesView(bodyHeight)
		case tabUpload:
			body = m.renderUploadView()
		case tabFeed:
			body = m.renderFeedView(bodyHeight)
		case tabSession:
			body = m.renderSessionView()
		case tabHelp:
			body = m.renderHelpView()
		}
	}
	b.WriteString(body)

	// Bottom Status & Help Bar
	b.WriteString("\n\n")
	statusBar := m.renderStatusBar()
	b.WriteString(statusBar)

	return b.String()
}

func (m AppModel) renderTopBar() string {
	logo := ui.TopBarBulletStyle.Render("◆") + " " + ui.TopBarLogoStyle.Render("ohfs") + " " + ui.TopBarVersionStyle.Render("v1.0")

	sessionText := "Session: None"
	if m.cfg.SessionKey != "" {
		nick := m.cfg.Nickname
		if nick == "" {
			nick = "User"
		}
		sessionText = fmt.Sprintf("👤 %s (%s)", nick, m.cfg.SessionKey)
	}
	sessionPill := ui.PillSessionStyle.Render(sessionText)

	pulse := ui.LivePulseStyle.Render(m.spinner.View() + " ONLINE")

	leftSide := lipgloss.JoinHorizontal(lipgloss.Center, logo, "   ", sessionPill)
	gap := m.width - lipgloss.Width(leftSide) - lipgloss.Width(pulse) - 2
	if gap < 2 {
		gap = 2
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, leftSide, strings.Repeat(" ", gap), pulse)
}

func (m AppModel) renderTabs() string {
	tabs := []string{
		"1 Files",
		"2 Upload",
		"3 Feed",
		"4 Session",
		"5 Help",
	}

	var renderedTabs []string
	for i, t := range tabs {
		if tabIndex(i) == m.activeTab {
			renderedTabs = append(renderedTabs, ui.TabActiveStyle.Render(" "+t+" "))
		} else {
			renderedTabs = append(renderedTabs, ui.TabInactiveStyle.Render(" "+t+" "))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}

func (m AppModel) renderFilesView(bodyHeight int) string {
	if m.cfg.SessionKey == "" {
		return ui.BoxCardStyle.Render(
			ui.StatusWarningStyle.Render("No active session configured!\n\n") +
				ui.DescHelpStyle.Render("Press '4' to switch to Session and generate a memorable or secure key."),
		)
	}

	items := m.getFilteredFiles()
	totalAll := len(m.files)
	totalFiltered := len(items)

	if totalAll == 0 {
		return ui.BoxCardStyle.Render(
			ui.TitleStyle.Render("📁 Workspace Files\n\n") +
				ui.DescHelpStyle.Render("No files in this session yet.\nPress 'u' or go to [2] Quick Upload to share your first file!"),
		)
	}

	var b strings.Builder

	// Top Title + Counter Badge + Search Filter Status
	title := ui.TitleStyle.Render("📁 Workspace Files")
	counterBadge := ui.DescHelpStyle.Render(fmt.Sprintf("[%d/%d files]", m.filesCursor+1, totalFiltered))
	if m.filesFilterInput.Value() != "" {
		counterBadge = ui.PillSessionStyle.Render(fmt.Sprintf("Filter: %q (%d/%d)", m.filesFilterInput.Value(), totalFiltered, totalAll))
	}

	// Scroll arrows
	arrows := ""
	visibleRows := max(5, bodyHeight-7)
	if m.filesScroll > 0 {
		arrows += "▲ "
	}
	if m.filesScroll+visibleRows < totalFiltered {
		arrows += "▼"
	}
	if arrows != "" {
		counterBadge += " " + ui.KeyHelpStyle.Render(arrows)
	}

	headerLine := title + "  " + counterBadge
	if m.filesFiltering {
		headerLine = title + "  " + ui.KeyHelpStyle.Render("Search: ") + m.filesFilterInput.View() + " " + ui.DescHelpStyle.Render("(Esc to close)")
	}
	b.WriteString(headerLine + "\n\n")

	// Table Headers
	contentWidth := max(70, m.width-8)
	nameWidth := max(22, contentWidth-56)
	hdrName := padCol("NAME", nameWidth)
	hdrSlug := padCol("SLUG", 12)
	hdrSize := padCol("SIZE", 9)
	hdrVis := padCol("VISIBILITY", 14)
	hdrDls := padCol("DLS", 6)
	hdrTime := padCol("UPLOADED", 11)
	headerText := fmt.Sprintf("  %s %s %s %s %s %s", hdrName, hdrSlug, hdrSize, hdrVis, hdrDls, hdrTime)
	b.WriteString(ui.TableHeaderStyle.Render(headerText) + "\n")

	if totalFiltered == 0 {
		b.WriteString(ui.DescHelpStyle.Render("\n  No files matching current filter.\n"))
	} else {
		end := min(totalFiltered, m.filesScroll+visibleRows)
		start := m.filesScroll
		if start < 0 {
			start = 0
		}
		if start >= totalFiltered {
			start = max(0, totalFiltered-1)
		}
		if end < start {
			end = start
		}

		for idx := start; idx < end; idx++ {
			f := items[idx]

			visBadge := ui.PillPrivateStyle.Render("PRIVATE")
			if f.IsPublic {
				visBadge = ui.PillPublicStyle.Render("PUBLIC")
			}
			if f.IsOneTime {
				visBadge += " " + ui.PillOneTimeStyle.Render("1-TIME")
			}
			if f.PasswordRequired {
				visBadge += " " + ui.PillPasswordStyle.Render("PWD")
			}

			colName := padCol(ui.TruncateString(f.Filename, nameWidth), nameWidth)
			colSlug := padCol(ui.TruncateString(f.Slug, 12), 12)
			colSize := padCol(ui.FormatBytes(f.GetEffectiveSize()), 9)
			colVis := padCol(visBadge, 14)
			colDls := padCol(fmt.Sprintf("%d", f.Downloads), 6)
			colTime := padCol(ui.FormatRelativeTime(f.UploadedAt), 11)

			rowLine := fmt.Sprintf("%s %s %s %s %s %s", colName, colSlug, colSize, colVis, colDls, colTime)

			if idx == m.filesCursor {
				b.WriteString(ui.TableSelectedRowStyle.Render("▸ "+rowLine) + "\n")
			} else {
				b.WriteString(ui.TableRowStyle.Render("  "+rowLine) + "\n")
			}
		}
	}

	// Bottom Scroll Indicator
	if m.filesScroll+visibleRows < totalFiltered {
		remaining := totalFiltered - (m.filesScroll + visibleRows)
		b.WriteString(ui.DescHelpStyle.Render(fmt.Sprintf("  ▼ %d more files below\n", remaining)))
	} else {
		b.WriteString("\n")
	}

	// Navigation shortcuts hint
	hint := ui.KeyHelpStyle.Render("Keys: ") +
		ui.DescHelpStyle.Render("[↑/↓] Move  •  [/] Filter  •  [d] Download  •  [s/Enter] Share/QR  •  [x] Delete  •  [r] Refresh  •  [u] Upload")
	b.WriteString(hint)

	return ui.BoxCardStyle.Render(b.String())
}

func (m AppModel) renderFeedView(bodyHeight int) string {
	items := m.getFilteredFeed()
	totalAll := len(m.feed)
	totalFiltered := len(items)

	if totalAll == 0 {
		return ui.BoxCardStyle.Render(
			ui.TitleStyle.Render("🌐 Public Feed\n\n") +
				ui.DescHelpStyle.Render("No public files available or still fetching...\nPress 'r' to refresh."),
		)
	}

	var b strings.Builder

	// Title + Counter Badge + Search Filter Status
	title := ui.TitleStyle.Render("🌐 Public Feed")
	counterBadge := ui.DescHelpStyle.Render(fmt.Sprintf("[%d/%d files]", m.feedCursor+1, totalFiltered))
	if m.feedFilterInput.Value() != "" {
		counterBadge = ui.PillSessionStyle.Render(fmt.Sprintf("Filter: %q (%d/%d)", m.feedFilterInput.Value(), totalFiltered, totalAll))
	}

	// Scroll arrows
	arrows := ""
	visibleRows := max(5, bodyHeight-7)
	if m.feedScroll > 0 {
		arrows += "▲ "
	}
	if m.feedScroll+visibleRows < totalFiltered {
		arrows += "▼"
	}
	if arrows != "" {
		counterBadge += " " + ui.KeyHelpStyle.Render(arrows)
	}

	headerLine := title + "  " + counterBadge
	if m.feedFiltering {
		headerLine = title + "  " + ui.KeyHelpStyle.Render("Search: ") + m.feedFilterInput.View() + " " + ui.DescHelpStyle.Render("(Esc to close)")
	}
	b.WriteString(headerLine + "\n\n")

	// Table Headers
	contentWidth := max(70, m.width-8)
	nameWidth := max(24, contentWidth-42)
	hdrName := padCol("NAME", nameWidth)
	hdrSlug := padCol("SLUG", 12)
	hdrSize := padCol("SIZE", 9)
	hdrDls := padCol("DLS", 6)
	hdrTime := padCol("UPLOADED", 11)
	headerText := fmt.Sprintf("  %s %s %s %s %s", hdrName, hdrSlug, hdrSize, hdrDls, hdrTime)
	b.WriteString(ui.TableHeaderStyle.Render(headerText) + "\n")

	if totalFiltered == 0 {
		b.WriteString(ui.DescHelpStyle.Render("\n  No files matching current filter.\n"))
	} else {
		end := min(totalFiltered, m.feedScroll+visibleRows)
		start := m.feedScroll
		if start < 0 {
			start = 0
		}
		if start >= totalFiltered {
			start = max(0, totalFiltered-1)
		}
		if end < start {
			end = start
		}

		for idx := start; idx < end; idx++ {
			f := items[idx]

			colName := padCol(ui.TruncateString(f.Filename, nameWidth), nameWidth)
			colSlug := padCol(ui.TruncateString(f.Slug, 12), 12)
			colSize := padCol(ui.FormatBytes(f.GetEffectiveSize()), 9)
			colDls := padCol(fmt.Sprintf("%d", f.Downloads), 6)
			colTime := padCol(ui.FormatRelativeTime(f.UploadedAt), 11)

			rowLine := fmt.Sprintf("%s %s %s %s %s", colName, colSlug, colSize, colDls, colTime)

			if idx == m.feedCursor {
				b.WriteString(ui.TableSelectedRowStyle.Render("▸ "+rowLine) + "\n")
			} else {
				b.WriteString(ui.TableRowStyle.Render("  "+rowLine) + "\n")
			}
		}
	}

	// Bottom Scroll Indicator
	if m.feedScroll+visibleRows < totalFiltered {
		remaining := totalFiltered - (m.feedScroll + visibleRows)
		b.WriteString(ui.DescHelpStyle.Render(fmt.Sprintf("  ▼ %d more files below\n", remaining)))
	} else {
		b.WriteString("\n")
	}

	// Navigation shortcuts hint
	hint := ui.KeyHelpStyle.Render("Keys: ") +
		ui.DescHelpStyle.Render("[↑/↓] Move  •  [/] Filter  •  [d] Download  •  [s/Enter] Share/QR  •  [r] Refresh")
	b.WriteString(hint)

	return ui.BoxCardStyle.Render(b.String())
}

func (m AppModel) renderUploadView() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.Render("QUICK FILE UPLOAD") + "\n\n")

	// Field 0: File Path
	f0Prefix := "  "
	if m.uploadFocusIdx == 0 {
		f0Prefix = "▸ "
	}
	b.WriteString(fmt.Sprintf("%s%-14s %s\n", f0Prefix, "File Path:", m.uploadInputs[0].View()))

	// Field 1: Custom Slug
	f1Prefix := "  "
	if m.uploadFocusIdx == 1 {
		f1Prefix = "▸ "
	}
	b.WriteString(fmt.Sprintf("%s%-14s %s\n", f1Prefix, "Custom Slug:", m.uploadInputs[1].View()))

	// Field 2: Password
	f2Prefix := "  "
	if m.uploadFocusIdx == 2 {
		f2Prefix = "▸ "
	}
	b.WriteString(fmt.Sprintf("%s%-14s %s\n\n", f2Prefix, "Password:", m.uploadInputs[2].View()))

	// Field 3: Visibility toggle
	f3Prefix := "  "
	if m.uploadFocusIdx == 3 {
		f3Prefix = "▸ "
	}
	var visText string
	if m.uploadPublic {
		visText = ui.KeyHelpStyle.Render("[●] Public") + "     " + ui.DescHelpStyle.Render("[○] Private")
	} else {
		visText = ui.DescHelpStyle.Render("[○] Public") + "     " + ui.KeyHelpStyle.Render("[●] Private")
	}
	b.WriteString(fmt.Sprintf("%s%-14s %s   %s\n", f3Prefix, "Visibility:", visText, ui.DescHelpStyle.Render("(press 'p' or Enter to toggle)")))

	// Field 4: One-time toggle
	f4Prefix := "  "
	if m.uploadFocusIdx == 4 {
		f4Prefix = "▸ "
	}
	var otText string
	if m.uploadOneTime {
		otText = ui.KeyHelpStyle.Render("[●] Burn after 1 view") + "   " + ui.DescHelpStyle.Render("[○] Normal")
	} else {
		otText = ui.DescHelpStyle.Render("[○] Burn after 1 view") + "   " + ui.KeyHelpStyle.Render("[●] Normal")
	}
	b.WriteString(fmt.Sprintf("%s%-14s %s   %s\n\n", f4Prefix, "One-Time:", otText, ui.DescHelpStyle.Render("(press 'o' or Enter to toggle)")))

	// Field 5: Action Button
	btnStyle := ui.TabInactiveStyle
	if m.uploadFocusIdx == 5 {
		btnStyle = ui.TabActiveStyle
	}
	if m.uploading {
		b.WriteString("  " + ui.LivePulseStyle.Render(m.spinner.View()+" Uploading file to OhioFiles... [████████████░░░░]") + "\n\n")
	} else {
		b.WriteString("  " + btnStyle.Render(" [ Upload File (Enter) ] ") + "\n\n")
	}

	b.WriteString("  " + ui.DescHelpStyle.Render("Use [↑/↓] or [Tab] to navigate fields  •  [p/o] to toggle  •  [Enter] to submit"))

	return ui.BoxCardStyle.Render(b.String())
}

func (m AppModel) renderSessionView() string {
	var b strings.Builder

	key := m.cfg.SessionKey
	if key == "" {
		key = "<none active>"
	}
	nick := m.cfg.Nickname
	if nick == "" {
		nick = "User"
	}
	slug := m.cfg.ProfileSlug
	if slug == "" {
		slug = "<none>"
	}
	profileURL := fmt.Sprintf("https://ohfs.app/p/%s", slug)
	if slug == "<none>" {
		profileURL = "<no public profile>"
	}

	// Section 1: Active Identity
	b.WriteString(ui.TitleStyle.Render("ACTIVE IDENTITY") + "\n\n")
	b.WriteString(fmt.Sprintf("  %-16s %s   %s\n", "Session Key", ui.KeyHelpStyle.Render(key), ui.DescHelpStyle.Render("[c] Copy Key")))
	if m.sessionEditingNick {
		b.WriteString(fmt.Sprintf("  %-16s %s %s\n", "Nickname", m.sessionNicknameInput.View(), ui.DescHelpStyle.Render("(Enter to save, Esc to cancel)")))
	} else {
		b.WriteString(fmt.Sprintf("  %-16s %s   %s\n", "Nickname", ui.TitleStyle.Render(nick), ui.DescHelpStyle.Render("[e] Rename")))
	}
	b.WriteString(fmt.Sprintf("  %-16s %s\n", "Profile URL", ui.DescHelpStyle.Render(profileURL)))
	b.WriteString(fmt.Sprintf("  %-16s %s\n", "Device ID", ui.DescHelpStyle.Render(m.cfg.DeviceId)))
	b.WriteString(fmt.Sprintf("  %-16s %s   %s\n\n", "API Endpoint", ui.DescHelpStyle.Render(m.cfg.ResolveApiUrl()), ui.LivePulseStyle.Render("● ONLINE")))

	// Section 2: Quick Actions
	b.WriteString(ui.TitleStyle.Render("QUICK ACTIONS") + "\n\n")
	b.WriteString(fmt.Sprintf("  %s  %-22s %s\n", ui.KeyHelpStyle.Render("[n]"), "New Memorable Key", ui.DescHelpStyle.Render("Generate easy-to-remember session (e.g. ohio_FastMonkey252)")))
	b.WriteString(fmt.Sprintf("  %s  %-22s %s\n", ui.KeyHelpStyle.Render("[s]"), "New Secure Key", ui.DescHelpStyle.Render("Generate 192-bit cryptographic session key")))
	b.WriteString(fmt.Sprintf("  %s  %-22s %s\n", ui.KeyHelpStyle.Render("[e]"), "Change Nickname", ui.DescHelpStyle.Render("Update your public profile display name")))
	b.WriteString(fmt.Sprintf("  %s  %-22s %s\n", ui.KeyHelpStyle.Render("[c]"), "Copy Session Key", ui.DescHelpStyle.Render("Copy active key directly to system clipboard")))

	return ui.BoxCardStyle.Render(b.String())
}

func (m AppModel) renderHelpView() string {
	col1Header := ui.TitleStyle.Render("DASHBOARD SHORTCUTS")
	col1 := col1Header + "\n\n" +
		ui.KeyHelpStyle.Render("  [Tab] / [1-5]   ") + ui.DescHelpStyle.Render("Switch active tabs") + "\n" +
		ui.KeyHelpStyle.Render("  [↑/↓] or [j/k]  ") + ui.DescHelpStyle.Render("Navigate items & rows") + "\n" +
		ui.KeyHelpStyle.Render("  [/]             ") + ui.DescHelpStyle.Render("Search & filter files") + "\n" +
		ui.KeyHelpStyle.Render("  [s] / [Enter]   ") + ui.DescHelpStyle.Render("Share file & view QR code") + "\n" +
		ui.KeyHelpStyle.Render("  [d]             ") + ui.DescHelpStyle.Render("Download selected file") + "\n" +
		ui.KeyHelpStyle.Render("  [x]             ") + ui.DescHelpStyle.Render("Delete selected file") + "\n" +
		ui.KeyHelpStyle.Render("  [u]             ") + ui.DescHelpStyle.Render("Jump to Quick Upload") + "\n" +
		ui.KeyHelpStyle.Render("  [r]             ") + ui.DescHelpStyle.Render("Refresh current view") + "\n" +
		ui.KeyHelpStyle.Render("  [Esc]           ") + ui.DescHelpStyle.Render("Clear search / Close modal") + "\n" +
		ui.KeyHelpStyle.Render("  [q]             ") + ui.DescHelpStyle.Render("Quit dashboard")

	col2Header := ui.TitleStyle.Render("CLI COMMAND REFERENCE")
	col2 := col2Header + "\n\n" +
		ui.KeyHelpStyle.Render("  ohfs upload <file>       ") + ui.DescHelpStyle.Render("Upload with progress bar") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs upload -p -o <file> ") + ui.DescHelpStyle.Render("Public & one-time upload") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs ls [--public]       ") + ui.DescHelpStyle.Render("List session or public files") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs download <slug>     ") + ui.DescHelpStyle.Render("Download with range resume") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs share <slug>        ") + ui.DescHelpStyle.Render("Direct URL & ASCII QR code") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs session new         ") + ui.DescHelpStyle.Render("Create memorable session") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs session new --sec   ") + ui.DescHelpStyle.Render("Create cryptographic key") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs session whoami      ") + ui.DescHelpStyle.Render("Inspect active identity") + "\n" +
		ui.KeyHelpStyle.Render("  ohfs folder zip <id>     ") + ui.DescHelpStyle.Render("Stream folder as ZIP")

	col1Width := min(44, max(36, (m.width-12)/2))
	col2Width := min(52, max(40, (m.width-12)/2))

	col1Box := lipgloss.NewStyle().Width(col1Width).Render(col1)
	col2Box := lipgloss.NewStyle().Width(col2Width).Render(col2)

	divider := lipgloss.NewStyle().Foreground(ui.ColorBorder).Render("│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│\n│")
	content := lipgloss.JoinHorizontal(lipgloss.Top, col1Box, "  ", divider, "  ", col2Box)

	return ui.BoxCardStyle.Render(content)
}

func (m AppModel) renderModal() string {
	var b strings.Builder
	b.WriteString(ui.TitleStyle.Render("  " + m.modalTitle) + "\n\n")
	b.WriteString(m.modalQRCode + "\n")
	b.WriteString("  Link: " + ui.KeyHelpStyle.Render(m.modalURL) + "\n")
	if m.modalSubtext != "" {
		b.WriteString("  " + ui.SubtitleStyle.Render(m.modalSubtext) + "\n\n")
	}
	b.WriteString("  " + ui.DescHelpStyle.Render("Press [Esc], [Enter], or [q] to close."))
	return ui.ModalBoxStyle.Render(b.String())
}

func (m AppModel) renderStatusBar() string {
	if m.statusText == "" {
		return ui.DescHelpStyle.Render(" [Tab] Switch tabs  •  [1-5] Jump  •  [/] Filter  •  [q] Quit")
	}

	if m.statusErr {
		return ui.StatusErrorStyle.Render(" ✖ " + m.statusText)
	}
	return ui.StatusSuccessStyle.Render(" ✔ " + m.statusText)
}
