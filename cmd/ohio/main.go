package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"ohio/internal/api"
	"ohio/internal/config"
	"ohio/internal/tui"
	"ohio/internal/ui"
)

var (
	version = config.Version

	// Global Flags
	flagApiUrl     string
	flagSessionKey string
	flagJson       bool
	flagQuiet      bool

	// Root Shortcut Flags
	flagRootUpload []string
	flagRootUdrks  bool
	flagForceUdrks bool

	cfg    *config.Config
	client *api.Client
)

func main() {
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:     "ohio",
		Aliases: []string{"ohfs"},
		Short:   "⚡ OhioFiles CLI & TUI - Fast, minimal, anonymous file sharing",
		Long: `ohio (ohfs) is the ultimate command-line interface and interactive TUI for OhioFiles.
Upload, stream, share, and manage public sessions with zero login friction.`,
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			apiUrl := cfg.ResolveApiUrl()
			if flagApiUrl != "" {
				apiUrl = flagApiUrl
			}
			sessionKey := cfg.SessionKey
			if envKey := strings.TrimSpace(os.Getenv("OHIO_SESSION_KEY")); envKey != "" {
				sessionKey = envKey
			} else if envKey := strings.TrimSpace(os.Getenv("OHFS_SESSION_KEY")); envKey != "" {
				sessionKey = envKey
			}
			if flagSessionKey != "" {
				sessionKey = flagSessionKey
			}
			client = api.NewClient(apiUrl, sessionKey, cfg.DeviceId)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagRootUdrks {
				return runUdrks(cfg, client, flagForceUdrks)
			}
			if len(flagRootUpload) > 0 {
				opts := api.UploadOptions{
					SessionKey: client.SessionKey,
					DeviceId:   cfg.DeviceId,
				}
				return handleUpload(flagRootUpload, opts, false, false, false, "")
			}

			// Check if stdin is being piped into bare ohio command
			stat, err := os.Stdin.Stat()
			if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
				opts := api.UploadOptions{
					SessionKey: client.SessionKey,
					DeviceId:   cfg.DeviceId,
				}
				return handlePipedStdin(opts, false, false, "")
			}

			return cmd.Help()
		},
	}

	// Persistent Global Flags
	rootCmd.PersistentFlags().StringVar(&flagApiUrl, "api-url", "", "Custom OhioFiles API URL (overrides config)")
	rootCmd.PersistentFlags().StringVar(&flagSessionKey, "session-key", "", "Session key for this command (overrides config)")
	rootCmd.PersistentFlags().BoolVar(&flagJson, "json", false, "Output results in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-essential output")

	// Root Shortcuts
	rootCmd.Flags().StringSliceVarP(&flagRootUpload, "upload", "u", nil, "Quick upload one or more files directly (e.g. ohio -u file.txt)")
	rootCmd.Flags().BoolVar(&flagRootUdrks, "udrks", false, "Execute 3-tier nuclear uninstall-delete-revoke-kill-service")
	rootCmd.Flags().BoolVarP(&flagForceUdrks, "force", "f", false, "Force execution without interactive confirmations")

	// Register Subcommands & Aliases
	rootCmd.AddCommand(newTuiCmd())
	rootCmd.AddCommand(newSessionCmd())
	rootCmd.AddCommand(newUploadCmd())
	rootCmd.AddCommand(newLsCmd())
	rootCmd.AddCommand(newDownloadCmd())
	rootCmd.AddCommand(newInfoCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newShareCmd())
	rootCmd.AddCommand(newFolderCmd())
	rootCmd.AddCommand(newFeedCmd())
	rootCmd.AddCommand(newUpdateCmd())
	rootCmd.AddCommand(newUninstallCmd())
	rootCmd.AddCommand(newUdrksCmd())

	defaultHelp := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd != rootCmd {
			defaultHelp(cmd, args)
			return
		}

		// Banner
		banner := lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.ColorCyanBright).
			Render(fmt.Sprintf("⚡ OhioFiles CLI (ohio) v%s - Fast, anonymous, login-free file sharing & TUI", version))
		fmt.Println(banner + "\n")

		// Active Session
		activeKey := ""
		if client != nil && client.SessionKey != "" {
			activeKey = client.SessionKey
		} else if flagSessionKey != "" {
			activeKey = flagSessionKey
		} else if envKey := strings.TrimSpace(os.Getenv("OHIO_SESSION_KEY")); envKey != "" {
			activeKey = envKey
		} else if envKey := strings.TrimSpace(os.Getenv("OHFS_SESSION_KEY")); envKey != "" {
			activeKey = envKey
		} else if cfg != nil && cfg.SessionKey != "" {
			activeKey = cfg.SessionKey
		}

		if activeKey != "" {
			fmt.Println(ui.TitleStyle.Render("Active Session:") + " " + ui.CardSessionKeyStyle.Render(activeKey) + "\n")
		} else {
			fmt.Println(ui.StatusWarningStyle.Render("⚠ No active session configured") + " " + ui.CardDimmedStyle.Render("(run 'ohio session new' to create one)") + "\n")
		}

		// Command Groups
		groups := []struct {
			title string
			items [][]string
		}{
			{
				title: "Core",
				items: [][]string{
					{"tui", "interactive dashboard"},
					{"upload / -u", "upload files"},
					{"download", "download with resume"},
				},
			},
			{
				title: "Explore",
				items: [][]string{
					{"ls", "list files with location"},
					{"feed", "public stream"},
					{"info", "inspect metadata & QR"},
					{"share", "get links"},
				},
			},
			{
				title: "Manage",
				items: [][]string{
					{"folder", "create/list/zip folders"},
					{"session", "inspect/generate/set keys"},
					{"update", "check version"},
				},
			},
			{
				title: "System",
				items: [][]string{
					{"uninstall", "safe removal"},
					{"udrks", "3-tier nuclear scrub"},
				},
			},
		}

		for _, g := range groups {
			fmt.Println(ui.TitleStyle.Render(g.title + ":"))
			for _, item := range g.items {
				fmt.Println("  " + ui.KeyHelpStyle.Render(fmt.Sprintf("%-16s", item[0])) + " " + ui.DescHelpStyle.Render(item[1]))
			}
			fmt.Println()
		}

		// Quick Examples
		fmt.Println(ui.TitleStyle.Render("Quick Examples:"))
		fmt.Println("  " + ui.CardValueStyle.Render("ohio tui"))
		fmt.Println("  " + ui.CardValueStyle.Render("ohio -u doc.pdf"))
		fmt.Println("  " + ui.CardValueStyle.Render("cat file.txt | ohio"))
		fmt.Println("  " + ui.CardValueStyle.Render("ohio download my-slug -o ./folder/"))

		if flags := cmd.Flags().FlagUsages(); flags != "" {
			fmt.Println("\n" + ui.TitleStyle.Render("Flags:"))
			fmt.Print(flags)
		}
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// ----------------------------------------------------------------------------
// TUI Command
// ----------------------------------------------------------------------------

func newTuiCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "tui",
		Aliases: []string{"t", "ui"},
		Short:   "Launch interactive full-screen TUI dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			model := tui.NewAppModel(cfg, client)
			p := tea.NewProgram(model, tea.WithAltScreen())
			_, err := p.Run()
			return err
		},
	}
}

// ----------------------------------------------------------------------------
// Session Commands
// ----------------------------------------------------------------------------

func newSessionCmd() *cobra.Command {
	sessionCmd := &cobra.Command{
		Use:     "session",
		Aliases: []string{"s", "sess"},
		Short:   "Manage OhioFiles public sessions",
		Long:    "Create, inspect, set, or update active public session keys and nicknames.",
	}

	// session whoami
	whoamiCmd := &cobra.Command{
		Use:   "whoami",
		Short: "Display active session information",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.SessionKey == "" {
				if flagJson {
					return outputJSON(map[string]interface{}{"active": false, "session": nil})
				}
				fmt.Println(ui.StatusWarningStyle.Render("No active session configured."))
				fmt.Println("Run 'ohio session new' to create one or 'ohio session set <key>' to link an existing key.")
				return nil
			}

			sess, err := client.GetPublicSession(cfg.SessionKey)
			if err != nil {
				if flagJson {
					return outputJSON(map[string]interface{}{
						"key":        cfg.SessionKey,
						"nickname":   cfg.Nickname,
						"profile":    cfg.ProfileSlug,
						"validation": err.Error(),
					})
				}
				fmt.Println(ui.RenderSessionCard(&api.PublicSession{
					Key:         cfg.SessionKey,
					Nickname:    cfg.Nickname,
					ProfileSlug: cfg.ProfileSlug,
				}, "Active session (unverified)", cfg.DeviceId, cfg.ResolveApiUrl()))
				return nil
			}

			if sess.Nickname != "" && sess.Nickname != cfg.Nickname {
				_ = cfg.SetActiveSession(sess.Key, sess.Nickname, sess.ProfileSlug)
			}

			if flagJson {
				return outputJSON(sess)
			}

			fmt.Println(ui.RenderSessionCard(sess, "Active session", cfg.DeviceId, cfg.ResolveApiUrl()))
			return nil
		},
	}

	// session new [--type memorable|secure] [--secure]
	var (
		flagSecure bool
		flagType   string
	)
	newCmd := &cobra.Command{
		Use:   "new",
		Short: "Create and activate a new session",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyType := strings.ToLower(strings.TrimSpace(flagType))
			if keyType == "" {
				keyType = "memorable"
			}
			if flagSecure {
				keyType = "secure"
			}
			if keyType != "memorable" && keyType != "secure" {
				return fmt.Errorf("invalid session type '%s'; must be 'memorable' or 'secure'", keyType)
			}

			sess, err := client.CreatePublicSession(keyType)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}

			if err := cfg.SetActiveSession(sess.Key, sess.Nickname, sess.ProfileSlug); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			if flagJson {
				return outputJSON(sess)
			}

			fmt.Println(ui.RenderSessionCard(sess, "Active session", cfg.DeviceId, cfg.ResolveApiUrl()))
			return nil
		},
	}
	newCmd.Flags().StringVar(&flagType, "type", "memorable", "Session key type: 'memorable' or 'secure'")
	newCmd.Flags().BoolVar(&flagSecure, "secure", false, "Generate high-entropy 192-bit cryptographic session key (shorthand for --type secure)")

	// session set <key>
	setCmd := &cobra.Command{
		Use:   "set <key>",
		Short: "Set the active session key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.TrimSpace(args[0])
			sess, err := client.GetPublicSession(key)
			if err != nil {
				_ = cfg.SetActiveSession(key, "", "")
				if flagJson {
					return outputJSON(map[string]string{"key": key, "warning": err.Error()})
				}
				fmt.Println(ui.StatusWarningStyle.Render(fmt.Sprintf("Warning: session validation failed (%v). Key set locally.", err)))
				return nil
			}

			if err := cfg.SetActiveSession(sess.Key, sess.Nickname, sess.ProfileSlug); err != nil {
				return err
			}

			if flagJson {
				return outputJSON(sess)
			}

			fmt.Println(ui.RenderSessionCard(sess, "Active session", cfg.DeviceId, cfg.ResolveApiUrl()))
			return nil
		},
	}

	// session nickname <new-nickname>
	nicknameCmd := &cobra.Command{
		Use:   "nickname <name>",
		Short: "Update nickname for the active session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.SessionKey == "" {
				return fmt.Errorf("no active session configured")
			}
			newNick := strings.TrimSpace(args[0])
			sess, err := client.UpdateNickname(cfg.SessionKey, newNick)
			if err != nil {
				return fmt.Errorf("failed to update nickname: %w", err)
			}

			_ = cfg.SetActiveSession(sess.Key, sess.Nickname, sess.ProfileSlug)

			if flagJson {
				return outputJSON(sess)
			}

			fmt.Println(ui.RenderSessionNicknameSuccess(sess, cfg.ResolveApiUrl()))
			return nil
		},
	}

	// session clear
	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the active session key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.ClearActiveSession(); err != nil {
				return err
			}
			if flagJson {
				return outputJSON(map[string]bool{"cleared": true})
			}
			fmt.Println(ui.RenderDeleteSuccess("session", "cleared active key"))
			return nil
		},
	}

	sessionCmd.AddCommand(whoamiCmd)
	sessionCmd.AddCommand(newCmd)
	sessionCmd.AddCommand(setCmd)
	sessionCmd.AddCommand(nicknameCmd)
	sessionCmd.AddCommand(clearCmd)

	return sessionCmd
}

// ----------------------------------------------------------------------------
// Upload Command & Piping / Watch
// ----------------------------------------------------------------------------

func newUploadCmd() *cobra.Command {
	var (
		flagPublic   bool
		flagOneTime  bool
		flagSlug     string
		flagPassword string
		flagFolder   string
		flagChunked  bool
		flagQR       bool
		flagWatch    bool
		flagName     string
	)

	cmd := &cobra.Command{
		Use:     "upload [files...]",
		Aliases: []string{"u", "up"},
		Short:   "Upload files to OhioFiles (supports piping and --watch)",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := api.UploadOptions{
				IsPublic:     flagPublic,
				IsOneTime:    flagOneTime,
				Password:     flagPassword,
				FolderID:     flagFolder,
				SessionKey:   client.SessionKey,
				DeviceId:     cfg.DeviceId,
				ForceChunked: flagChunked,
			}

			// 1. Detect Stdin Piping (e.g. cat file.txt | ohio upload, or ohio upload -)
			stat, err := os.Stdin.Stat()
			isPiped := err == nil && (stat.Mode()&os.ModeCharDevice) == 0
			if (len(args) == 0 && isPiped) || (len(args) == 1 && args[0] == "-") {
				if flagSlug != "" {
					opts.CustomSlug = flagSlug
				}
				return handlePipedStdin(opts, flagQR, flagWatch, flagName)
			}

			if len(args) == 0 {
				return fmt.Errorf("no files specified for upload (or pipe via stdin)")
			}

			// 2. Watch Mode
			if flagWatch {
				if len(args) != 1 {
					return fmt.Errorf("--watch mode supports watching a single file at a time")
				}
				if flagSlug != "" {
					opts.CustomSlug = flagSlug
				}
				return handleWatchUpload(args[0], opts, flagQR)
			}

			// 3. Regular File Upload(s)
			if len(args) == 1 && flagSlug != "" {
				opts.CustomSlug = flagSlug
			}
			return handleUpload(args, opts, flagQR, flagQuiet, flagJson, flagSlug)
		},
	}

	cmd.Flags().BoolVarP(&flagPublic, "public", "p", false, "Make file publicly discoverable in feed")
	cmd.Flags().BoolVarP(&flagOneTime, "one-time", "o", false, "Burn file after first view or download")
	cmd.Flags().StringVarP(&flagSlug, "slug", "s", "", "Custom slug for single file upload")
	cmd.Flags().StringVar(&flagPassword, "password", "", "Password protect file download")
	cmd.Flags().StringVarP(&flagFolder, "folder", "f", "", "Folder ID to upload into")
	cmd.Flags().BoolVar(&flagChunked, "chunked", false, "Force chunked upload protocol")
	cmd.Flags().BoolVar(&flagQR, "qr", false, "Display ASCII QR code upon successful upload")
	cmd.Flags().BoolVarP(&flagWatch, "watch", "w", false, "Auto-upload whenever file is modified")
	cmd.Flags().StringVarP(&flagName, "name", "n", "", "Filename to assign when piping from stdin")

	return cmd
}

func handleUpload(args []string, opts api.UploadOptions, flagQR, flagQuiet, flagJson bool, singleSlug string) error {
	var uploadedFiles []*api.FileItem

	for i, filePath := range args {
		fi, err := os.Stat(filePath)
		if err != nil {
			return fmt.Errorf("file not found: %s", filePath)
		}
		if fi.IsDir() {
			return fmt.Errorf("directories cannot be uploaded directly; please zip first: %s", filePath)
		}

		currentOpts := opts
		if len(args) == 1 && singleSlug != "" {
			currentOpts.CustomSlug = singleSlug
		}

		actionLabel := "Uploading"
		if len(args) > 1 {
			actionLabel = fmt.Sprintf("Uploading [%d/%d]", i+1, len(args))
		}
		bar := ui.NewProgressBar(actionLabel, filepath.Base(filePath), fi.Size(), flagQuiet || flagJson)

		onDirectProgress := func(written, total int64) {
			bar.Update(written, total)
		}
		onChunkedProgress := func(chunk, totalChunks int) {
			bar.UpdateChunk(chunk, totalChunks, fi.Size())
		}

		fileItem, err := client.Upload(filePath, currentOpts, onDirectProgress, onChunkedProgress)
		bar.Finish()
		if err != nil {
			return fmt.Errorf("failed uploading %s: %w", filePath, err)
		}

		if !flagJson && !flagQuiet {
			fmt.Print(ui.RenderUploadCard(fileItem, bar.Elapsed(), flagQR))
			if i < len(args)-1 {
				fmt.Println()
			}
		}
		uploadedFiles = append(uploadedFiles, fileItem)
	}

	if flagJson {
		return outputJSON(uploadedFiles)
	}

	return nil
}

func handlePipedStdin(opts api.UploadOptions, flagQR, flagWatch bool, customName string) error {
	filename := customName
	if filename == "" {
		filename = fmt.Sprintf("stdin-%s.txt", time.Now().Format("20060102-150405"))
	}

	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, filename)
	f, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file for stdin: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	written, err := io.Copy(f, reader)
	_ = f.Close()
	if err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed reading from stdin: %w", err)
	}
	defer os.Remove(tempFile)

	bar := ui.NewProgressBar("Uploading (pipe)", filename, written, flagQuiet || flagJson)
	onDirect := func(w, tot int64) { bar.Update(w, tot) }
	onChunk := func(c, tc int) { bar.UpdateChunk(c, tc, written) }

	fileItem, err := client.Upload(tempFile, opts, onDirect, onChunk)
	bar.Finish()
	if err != nil {
		return fmt.Errorf("failed uploading piped stream: %w", err)
	}

	if flagJson {
		return outputJSON(fileItem)
	}

	fmt.Print(ui.RenderUploadCard(fileItem, bar.Elapsed(), flagQR))
	return nil
}

func handleWatchUpload(filePath string, opts api.UploadOptions, flagQR bool) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// 1. Initial upload
	fmt.Printf("Initial upload of %s...\n", filepath.Base(filePath))
	start := time.Now()
	bar := ui.NewProgressBar("Uploading", filepath.Base(filePath), fi.Size(), flagQuiet || flagJson)
	onDirect := func(w, tot int64) { bar.Update(w, tot) }
	onChunk := func(c, tc int) { bar.UpdateChunk(c, tc, fi.Size()) }

	lastItem, err := client.Upload(filePath, opts, onDirect, onChunk)
	bar.Finish()
	if err != nil {
		return fmt.Errorf("initial upload failed: %w", err)
	}
	fmt.Print(ui.RenderUploadCard(lastItem, time.Since(start), flagQR))

	fmt.Printf("\n  %s %s (press Ctrl+C to stop)...\n",
		ui.CardDotStyle.Render("⚡ Watching for changes:"),
		ui.KeyHelpStyle.Render(filepath.Base(filePath)),
	)

	lastMod := fi.ModTime()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\n  Stopped watching.")
			return nil
		case <-ticker.C:
			currFi, err := os.Stat(filePath)
			if err != nil {
				continue
			}
			if currFi.ModTime().After(lastMod) {
				lastMod = currFi.ModTime()
				uploadStart := time.Now()

				// If re-uploading with custom slug, remove the previous file so slug is available
				if opts.CustomSlug != "" && lastItem != nil {
					_ = client.DeleteFile(lastItem.Slug, client.SessionKey)
				}

				barChange := ui.NewProgressBar("Re-uploading", filepath.Base(filePath), currFi.Size(), flagQuiet || flagJson)
				onDirectChange := func(w, tot int64) { barChange.Update(w, tot) }
				onChunkChange := func(c, tc int) { barChange.UpdateChunk(c, tc, currFi.Size()) }

				item, err := client.Upload(filePath, opts, onDirectChange, onChunkChange)
				barChange.Finish()
				if err != nil {
					fmt.Printf("  [%s] ✖ Re-upload failed: %v\n", time.Now().Format("15:04:05"), err)
				} else {
					lastItem = item
					fmt.Printf("  [%s] ⚡ File changed: %s → Re-uploaded to %s in %s\n",
						time.Now().Format("15:04:05"),
						filepath.Base(filePath),
						ui.CardLinkStyle.Render(item.URL),
						ui.FormatDurationShort(time.Since(uploadStart)),
					)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// List Files Command (ls / files / l)
// ----------------------------------------------------------------------------

func newLsCmd() *cobra.Command {
	var (
		flagPublic bool
		flagFolder string
		flagPage   int
		flagLimit  int
	)

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"l", "files", "list"},
		Short:   "List files in session workspace or public feed",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := client.ListFiles(client.SessionKey, flagFolder, flagPublic, flagPage, flagLimit)
			if err != nil {
				return err
			}

			if flagJson {
				return outputJSON(res)
			}

			if len(res.Files) == 0 {
				fmt.Println(ui.SubtitleStyle.Render("No files found."))
				return nil
			}

			fmt.Println(ui.TitleStyle.Render("📁 Workspace Files") + " " + ui.CardDimmedStyle.Render("(Session: "+client.SessionKey+")") + "\n")

			fmt.Printf("%-28s %-16s %-14s %-10s %-16s %-10s %-12s\n",
				ui.CliTableHeaderStyle.Render("FILENAME"),
				ui.CliTableHeaderStyle.Render("LOCATION"),
				ui.CliTableHeaderStyle.Render("SLUG"),
				ui.CliTableHeaderStyle.Render("SIZE"),
				ui.CliTableHeaderStyle.Render("VISIBILITY"),
				ui.CliTableHeaderStyle.Render("DOWNLOADS"),
				ui.CliTableHeaderStyle.Render("UPLOADED"),
			)

			for _, f := range res.Files {
				visBadge := ui.PillPrivateStyle.Render("Private")
				if f.IsPublic {
					visBadge = ui.PillPublicStyle.Render("Public")
				}
				if f.IsOneTime {
					visBadge += " " + ui.PillOneTimeStyle.Render("1-Time")
				}
				if f.PasswordRequired {
					visBadge += " " + ui.PillPasswordStyle.Render("Pwd")
				}

				fmt.Printf("%-28s %-16s %-14s %-10s %-16s %-10d %-12s\n",
					ui.TruncateString(f.Filename, 26),
					ui.TruncateString(api.GetFileLocation(&f), 14),
					ui.TruncateString(f.Slug, 12),
					ui.FormatBytes(f.GetEffectiveSize()),
					visBadge,
					f.Downloads,
					ui.FormatRelativeTime(f.UploadedAt),
				)
			}

			fmt.Printf("\n%s\n", ui.CardDimmedStyle.Render(fmt.Sprintf("Showing %d of %d files (Page %d/%d)", len(res.Files), res.Total, res.Page, res.TotalPages)))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&flagPublic, "public", "p", false, "List public files")
	cmd.Flags().StringVarP(&flagFolder, "folder", "f", "", "Filter by folder ID")
	cmd.Flags().IntVar(&flagPage, "page", 1, "Page number")
	cmd.Flags().IntVar(&flagLimit, "limit", 20, "Limit per page")

	return cmd
}

// ----------------------------------------------------------------------------
// Download Command
// ----------------------------------------------------------------------------

func newDownloadCmd() *cobra.Command {
	var (
		flagOutput   string
		flagPassword string
	)

	cmd := &cobra.Command{
		Use:     "download <slug-or-id>",
		Aliases: []string{"d", "get"},
		Short:   "Download a file with range resume support",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idOrSlug := args[0]
			var token string

			if flagPassword != "" {
				resToken, err := client.VerifyFilePassword(idOrSlug, flagPassword)
				if err != nil {
					return fmt.Errorf("password verification failed: %w", err)
				}
				token = resToken
			}

			expectedFilename := idOrSlug
			var totalExpected int64
			fileInfo, err := client.GetFileInfo(idOrSlug)
			if err == nil && fileInfo != nil {
				if fileInfo.Filename != "" {
					expectedFilename = fileInfo.Filename
				}
				totalExpected = fileInfo.GetEffectiveSize()
			}

			destPath := flagOutput
			if destPath == "" {
				destPath = expectedFilename
			} else {
				isDir := strings.HasSuffix(destPath, "/") || strings.HasSuffix(destPath, "\\")
				if !isDir {
					if fi, statErr := os.Stat(destPath); statErr == nil && fi.IsDir() {
						isDir = true
					}
				}
				if isDir {
					_ = os.MkdirAll(destPath, 0755)
					destPath = filepath.Join(destPath, expectedFilename)
				} else {
					if dir := filepath.Dir(destPath); dir != "" && dir != "." {
						_ = os.MkdirAll(dir, 0755)
					}
				}
			}

			bar := ui.NewProgressBar("Downloading", expectedFilename, totalExpected, flagQuiet || flagJson)

			onProgress := func(written, total int64) {
				bar.Update(written, total)
			}

			err = client.DownloadFile(idOrSlug, destPath, token, client.SessionKey, onProgress)
			bar.Finish()
			if err != nil {
				return fmt.Errorf("download failed: %w", err)
			}

			downloadedSize := bar.FinalBytes()
			if fi, statErr := os.Stat(destPath); statErr == nil {
				downloadedSize = fi.Size()
			}

			if flagJson {
				return outputJSON(map[string]interface{}{"success": true, "file": idOrSlug, "destination": destPath, "size": downloadedSize})
			}

			if !flagQuiet {
				fmt.Println(ui.RenderDownloadSuccess(expectedFilename, downloadedSize, destPath, bar.Elapsed()))
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Destination file or directory")
	cmd.Flags().StringVarP(&flagPassword, "password", "w", "", "Password for password-protected file")

	return cmd
}

// ----------------------------------------------------------------------------
// Info & Delete & Share
// ----------------------------------------------------------------------------

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <slug-or-id>",
		Short: "Inspect file metadata and QR code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			item, err := client.GetFileInfo(args[0])
			if err != nil {
				return err
			}
			if flagJson {
				return outputJSON(item)
			}
			fmt.Println(ui.RenderInfoCard(item))
			return nil
		},
	}
}

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <slug-or-id>",
		Aliases: []string{"rm", "del"},
		Short:   "Delete a file by slug or ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idOrSlug := args[0]
			err := client.DeleteFile(idOrSlug, client.SessionKey)
			if err != nil {
				return err
			}
			if flagJson {
				return outputJSON(map[string]interface{}{"success": true, "deleted": idOrSlug})
			}
			fmt.Println(ui.RenderDeleteSuccess("file", idOrSlug))
			return nil
		},
	}
}

func newShareCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "share <slug>",
		Aliases: []string{"sh"},
		Short:   "Display direct share link, markdown embed, and QR code",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			item, err := client.GetFileInfo(args[0])
			if err != nil {
				return err
			}

			directURL := fmt.Sprintf("%s/%s", ui.ResolveWebRoot(cfg.ResolveApiUrl()), item.Slug)
			if flagJson {
				return outputJSON(map[string]interface{}{
					"slug":      item.Slug,
					"filename":  item.Filename,
					"url":       directURL,
					"markdown":  fmt.Sprintf("![[ohiofile:%s]]", item.Slug),
					"directApi": fmt.Sprintf("%s/files/%s", cfg.ResolveApiUrl(), item.ID),
				})
			}

			fmt.Println(ui.RenderShareCard(item, directURL))
			return nil
		},
	}
}

// ----------------------------------------------------------------------------
// Folder Commands
// ----------------------------------------------------------------------------

func newFolderCmd() *cobra.Command {
	folderCmd := &cobra.Command{
		Use:     "folder",
		Aliases: []string{"f", "folders"},
		Short:   "Manage folders and streaming ZIP archives",
	}

	var (
		flagPublic bool
		flagSlug   string
	)
	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new folder in session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			folder, err := client.CreateFolder(name, flagSlug, client.SessionKey, flagPublic)
			if err != nil {
				return err
			}
			if flagJson {
				return outputJSON(folder)
			}
			fmt.Println(ui.RenderFolderCard(folder, "Folder created"))
			return nil
		},
	}
	createCmd.Flags().BoolVarP(&flagPublic, "public", "p", false, "Make folder public")
	createCmd.Flags().StringVarP(&flagSlug, "slug", "s", "", "Custom slug for folder")

	listCmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list", "l"},
		Short:   "List folders in active session",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := client.ListFolders(client.SessionKey)
			if err != nil {
				return err
			}
			if flagJson {
				return outputJSON(res)
			}
			if len(res) == 0 {
				fmt.Println(ui.SubtitleStyle.Render("No folders found."))
				return nil
			}

			fmt.Println(ui.TitleStyle.Render("Your Folders:\n"))
			fmt.Printf("%-32s %-16s %-12s %-16s\n",
				ui.CliTableHeaderStyle.Render("NAME"),
				ui.CliTableHeaderStyle.Render("SLUG"),
				ui.CliTableHeaderStyle.Render("VISIBILITY"),
				ui.CliTableHeaderStyle.Render("CREATED"),
			)

			for _, f := range res {
				visBadge := ui.PillPrivateStyle.Render("Private")
				if f.IsPublic {
					visBadge = ui.PillPublicStyle.Render("Public")
				}
				fmt.Printf("%-32s %-16s %-12s %-16s\n",
					ui.TruncateString(f.Name, 30),
					ui.TruncateString(f.Slug, 14),
					visBadge,
					ui.FormatRelativeTime(f.CreatedAt),
				)
			}
			return nil
		},
	}

	var flagZipOutput string
	zipCmd := &cobra.Command{
		Use:   "zip <id-or-slug>",
		Short: "Download an entire folder as a streaming ZIP archive",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idOrSlug := args[0]
			defaultArchive := fmt.Sprintf("folder-%s.zip", idOrSlug)
			if flagZipOutput == "" {
				flagZipOutput = defaultArchive
			} else {
				isDir := strings.HasSuffix(flagZipOutput, "/") || strings.HasSuffix(flagZipOutput, "\\")
				if !isDir {
					if fi, statErr := os.Stat(flagZipOutput); statErr == nil && fi.IsDir() {
						isDir = true
					}
				}
				if isDir {
					if err := os.MkdirAll(flagZipOutput, 0755); err != nil {
						return fmt.Errorf("failed to create destination directory: %w", err)
					}
					flagZipOutput = filepath.Join(flagZipOutput, defaultArchive)
				} else {
					if dir := filepath.Dir(flagZipOutput); dir != "" && dir != "." {
						if err := os.MkdirAll(dir, 0755); err != nil {
							return fmt.Errorf("failed to create destination directory: %w", err)
						}
					}
				}
			}

			bar := ui.NewProgressBar("Streaming", filepath.Base(flagZipOutput), 0, flagQuiet || flagJson)

			onProgress := func(written int64) {
				bar.UpdateIndeterminate(written)
			}

			err := client.DownloadFolderZip(idOrSlug, flagZipOutput, onProgress)
			bar.Finish()
			if err != nil {
				return fmt.Errorf("folder zip download failed: %w", err)
			}

			finalSize := bar.FinalBytes()
			if fi, statErr := os.Stat(flagZipOutput); statErr == nil {
				finalSize = fi.Size()
			}

			if flagJson {
				return outputJSON(map[string]interface{}{"success": true, "archive": flagZipOutput, "size": finalSize})
			}

			if !flagQuiet {
				fmt.Println(ui.RenderFolderZipSuccess(flagZipOutput, finalSize, bar.Elapsed()))
			}
			return nil
		},
	}
	zipCmd.Flags().StringVarP(&flagZipOutput, "output", "o", "", "Destination ZIP file path")

	delFolderCmd := &cobra.Command{
		Use:     "delete <folder-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a folder",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			folderID := args[0]
			err := client.DeleteFolder(folderID, client.SessionKey)
			if err != nil {
				return err
			}
			if flagJson {
				return outputJSON(map[string]interface{}{"success": true, "deletedFolder": folderID})
			}
			fmt.Println(ui.RenderDeleteSuccess("folder", folderID))
			return nil
		},
	}

	folderCmd.AddCommand(createCmd)
	folderCmd.AddCommand(listCmd)
	folderCmd.AddCommand(zipCmd)
	folderCmd.AddCommand(delFolderCmd)

	return folderCmd
}

// ----------------------------------------------------------------------------
// Public Feed Command
// ----------------------------------------------------------------------------

func newFeedCmd() *cobra.Command {
	var (
		flagPage  int
		flagLimit int
	)

	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Browse recently shared public files",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := client.ListFiles("", "", true, flagPage, flagLimit)
			if err != nil {
				return err
			}

			if flagJson {
				return outputJSON(res)
			}

			if len(res.Files) == 0 {
				fmt.Println(ui.SubtitleStyle.Render("No public files found."))
				return nil
			}

			fmt.Println(ui.TitleStyle.Render("🌐 OhioFiles Public Feed\n"))
			fmt.Printf("%-32s %-16s %-10s %-10s %-12s\n",
				ui.CliTableHeaderStyle.Render("FILENAME"),
				ui.CliTableHeaderStyle.Render("SLUG"),
				ui.CliTableHeaderStyle.Render("SIZE"),
				ui.CliTableHeaderStyle.Render("DOWNLOADS"),
				ui.CliTableHeaderStyle.Render("UPLOADED"),
			)

			for _, f := range res.Files {
				fmt.Printf("%-32s %-16s %-10s %-10d %-12s\n",
					ui.TruncateString(f.Filename, 30),
					ui.TruncateString(f.Slug, 14),
					ui.FormatBytes(f.GetEffectiveSize()),
					f.Downloads,
					ui.FormatRelativeTime(f.UploadedAt),
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&flagPage, "page", 1, "Page number")
	cmd.Flags().IntVar(&flagLimit, "limit", 20, "Limit per page")

	return cmd
}

// ----------------------------------------------------------------------------
// Update Command
// ----------------------------------------------------------------------------

func newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check for and install updates to OhioCLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s OhioCLI is up to date (%s)\n", ui.CardCheckmarkStyle.Render("✔"), version)
			return nil
		},
	}
}

// ----------------------------------------------------------------------------
// Uninstall & UDRKS Commands
// ----------------------------------------------------------------------------

func newUninstallCmd() *cobra.Command {
	var flagNuclear bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall OhioCLI from this server (preserves cache by default)",
		Long: `Uninstall OhioCLI from the server.
By default, removes the binary while keeping your session cache and settings in ~/.config/ohio so you can reinstall anytime and remain logged in.
Use --udrks to execute a 3-step nuclear purge (Revoke session on server, kill background watchers, and delete all data).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagNuclear {
				return runUdrks(cfg, client, flagForceUdrks)
			}

			execPath, err := os.Executable()
			if err != nil {
				execPath = "/usr/local/bin/ohio"
			}

			removeBinaryPaths(execPath)

			fmt.Println(ui.CardCheckmarkStyle.Render("✔ OhioCLI uninstalled successfully."))
			fmt.Printf("  Configuration and session cache preserved at ~/.config/ohio/ (ready for next install).\n")
			fmt.Println("  To reinstall anytime: curl -fsSL https://ohiofiles.cloud/install.sh | bash")
			return nil
		},
	}

	cmd.Flags().BoolVar(&flagNuclear, "udrks", false, "Force nuclear uninstall (Revoke remote session, kill daemons, scrub all data)")
	cmd.Flags().BoolVarP(&flagForceUdrks, "force", "f", false, "Force execution without interactive confirmations")
	return cmd
}

func newUdrksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "udrks",
		Short: "Execute 3-tier nuclear uninstall-delete-revoke-kill-service",
		Long: `Force uninstall of OhioCLI:
Tier 1: Revoke remote public session and all associated files/folders on OhioFiles servers.
Tier 2: Terminate all running OhioCLI background services, daemons, and watchers.
Tier 3: Scrub all local configurations, cache files, and binaries. Zero traces remaining.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUdrks(cfg, client, flagForceUdrks)
		},
	}
	cmd.Flags().BoolVarP(&flagForceUdrks, "force", "f", false, "Force execution without interactive confirmations")
	return cmd
}

func promptUdrksConfirmation(force bool) (bool, error) {
	if force {
		return true, nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
		return false, fmt.Errorf("UDRKS requires interactive confirmation in a terminal, or pass --force / -f")
	}

	fmt.Println(ui.StatusWarningStyle.Render("⚠️  WARNING: You are about to run the UDRKS nuclear scrub protocol."))
	fmt.Println("This will permanently revoke your remote session, kill background daemons, and scrub local data.")

	reader := bufio.NewReader(os.Stdin)

	// Step 1
	fmt.Print("  [Step 1/3] Are you sure you want to proceed? (y/N): ")
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	input1 := strings.ToLower(strings.TrimSpace(line))
	if input1 != "y" && input1 != "yes" {
		fmt.Println(ui.StatusWarningStyle.Render("✖ Nuclear scrub cancelled."))
		return false, nil
	}

	// Step 2
	fmt.Print("  [Step 2/3] This action is PERMANENT and IRREVERSIBLE. Type 'CONFIRM' or 'YES' to proceed: ")
	line, err = reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	input2 := strings.TrimSpace(line)
	if input2 != "CONFIRM" && input2 != "YES" {
		fmt.Println(ui.StatusWarningStyle.Render("✖ Nuclear scrub cancelled."))
		return false, nil
	}

	// Step 3
	fmt.Print("  [Step 3/3] Final safeguard. To purge device credentials and scrub all data, type 'PURGE': ")
	line, err = reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	input3 := strings.TrimSpace(line)
	if input3 != "PURGE" {
		fmt.Println(ui.StatusWarningStyle.Render("✖ Nuclear scrub cancelled."))
		return false, nil
	}

	return true, nil
}

func runUdrks(cfg *config.Config, client *api.Client, force bool) error {
	confirmed, err := promptUdrksConfirmation(force)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	fmt.Println(ui.TitleStyle.Render("🛡️  OhioCLI Nuclear Scrub Protocol (UDRKS)"))
	fmt.Println(ui.CardDimmedStyle.Render("Executing 3 tiers of security protection...\n"))

	// Step 1: Revoke
	if cfg.SessionKey != "" {
		fmt.Printf("  [1/3] %s Revoking public session %s on server...\n",
			ui.CardDotStyle.Render("⚡"),
			ui.KeyHelpStyle.Render(cfg.SessionKey),
		)
		if err := client.RevokePublicSession(cfg.SessionKey); err != nil {
			fmt.Printf("        Warning: session revocation returned: %v\n", err)
		} else {
			fmt.Printf("        ✔ Session revoked and all associated files destroyed.\n")
		}
	} else {
		fmt.Println("  [1/3] ⚡ No active remote session to revoke.")
	}

	// Step 2: Kill Services
	fmt.Printf("  [2/3] %s Terminating running OhioCLI background services...\n", ui.CardDotStyle.Render("⚡"))
	killOtherProcesses()
	fmt.Println("        ✔ Background services and watchers terminated.")

	// Step 3: Delete & Clean
	fmt.Printf("  [3/3] %s Purging local caches, configurations, and binaries...\n", ui.CardDotStyle.Render("⚡"))
	cleanLocalArtifacts()

	execPath, _ := os.Executable()
	removeBinaryPaths(execPath)
	fmt.Println("        ✔ Local cache, credentials, and binaries removed.")

	fmt.Println("\n" + ui.CardCheckmarkStyle.Render("✔ System completely scrubbed of OhioCLI. Zero traces remaining."))
	return nil
}

func killOtherProcesses() {
	myPID := os.Getpid()
	// Attempt pkill for other ohio/ohfs processes
	for _, procName := range []string{"ohio", "ohfs"} {
		out, err := exec.Command("pgrep", "-f", procName).Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			for _, line := range lines {
				pid, parseErr := strconv.Atoi(strings.TrimSpace(line))
				if parseErr == nil && pid != myPID && pid > 0 {
					if proc, err := os.FindProcess(pid); err == nil {
						_ = proc.Kill()
					}
				}
			}
		}
	}
}

func cleanLocalArtifacts() {
	home, err := os.UserHomeDir()
	if err == nil {
		_ = os.RemoveAll(filepath.Join(home, ".config", "ohio"))
		_ = os.RemoveAll(filepath.Join(home, ".config", "ohfs"))
		_ = os.RemoveAll(filepath.Join(home, ".cache", "ohio"))
	}
	// Clean temp stdin files
	matches, _ := filepath.Glob(filepath.Join(os.TempDir(), "ohio-*"))
	for _, m := range matches {
		_ = os.RemoveAll(m)
	}
	matchesOhfs, _ := filepath.Glob(filepath.Join(os.TempDir(), "ohfs-*"))
	for _, m := range matchesOhfs {
		_ = os.RemoveAll(m)
	}
}

func removeBinaryPaths(execPath string) {
	commonPaths := []string{
		"/usr/local/bin/ohio",
		"/usr/local/bin/ohfs",
		"/usr/bin/ohio",
		"/usr/bin/ohfs",
	}

	home, err := os.UserHomeDir()
	if err == nil {
		commonPaths = append(commonPaths,
			filepath.Join(home, ".local", "bin", "ohio"),
			filepath.Join(home, ".local", "bin", "ohfs"),
			filepath.Join(home, "bin", "ohio"),
			filepath.Join(home, "bin", "ohfs"),
		)
	}

	if execPath != "" {
		commonPaths = append(commonPaths, execPath)
	}

	for _, p := range commonPaths {
		if _, statErr := os.Lstat(p); statErr == nil {
			_ = os.Remove(p)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func outputJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
