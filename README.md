# OhioCLI (`ohio` / `ohfs`)

```text
 ██████╗ ██╗  ██╗██╗ ██████╗      ██████╗██╗     ██╗
██╔═══██╗██║  ██║██║██╔═══██╗    ██╔════╝██║     ██║
██║   ██║███████║██║██║   ██║    ██║     ██║     ██║
██║   ██║██╔══██║██║██║   ██║    ██║     ██║     ██║
╚██████╔╝██║  ██║██║╚██████╔╝    ╚██████╗███████╗██║
 ╚═════╝ ╚═╝  ╚═╝╚═╝ ╚═════╝      ╚═════╝╚══════╝╚═╝
```

<div align="center">

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue?style=for-the-badge&logo=apple&logoColor=white)](https://github.com/jurek-zsl/ohio-cli)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Release](https://img.shields.io/badge/Version-v2.2.0-orange?style=for-the-badge)](https://github.com/jurek-zsl/ohio-cli/releases)
[![Security](https://img.shields.io/badge/Security-UDRKS%203--Tier-red?style=for-the-badge)](https://github.com/jurek-zsl/ohio-cli)

**The Ultimate OhioFiles Command-Line Interface & Interactive Terminal Dashboard**  
*Fast, anonymous, login-free file sharing, folders, dev-watch synchronization, and real-time public feeds.*

[Install](#-one-line-installers) • [Core Features](#-core-features--shortcuts) • [Interactive TUI](#-interactive-tui-suite-ohio-tui) • [CLI Reference](#-cli-command-reference) • [Nuclear Protection](#-server-lifecycle--3-tier-nuclear-protection) • [Completions](#-shell-completions)

---

</div>

## ⚡ One-Line Installers

Install OhioCLI in seconds on any major operating system. Both installers configure the primary `ohio` binary, the backward-compatible `ohfs` symlink/alias, and register PATH variables.

### Linux & macOS (Bash / Zsh)
```bash
curl -fsSL https://api.ohiofiles.cloud/install-sh | bash
```
> *Or from a cloned repository:* `./install.sh`

### Windows (PowerShell)
```powershell
irm https://api.ohiofiles.cloud/install-ps1 | iex
```
> *Or locally via PowerShell:* `powershell -ExecutionPolicy Bypass -File .\install.ps1`

### Go Install
If you have Go 1.22+ installed, you can build and install directly to your `$GOPATH/bin`:
```bash
go install github.com/jurek-zsl/ohio-cli/cmd/ohio@latest
```

### Build From Source
```bash
git clone https://github.com/jurek-zsl/Next-OhioFiles.git
cd Next-OhioFiles/cli
go build -o bin/ohio ./cmd/ohio
# Optional: link alias
ln -sf $(pwd)/bin/ohio $(pwd)/bin/ohfs
```

---

## 🚀 Core Features & Shortcuts

OhioCLI gives power users and DevOps engineers high-velocity tools for sharing files straight from bash, zsh, tmux, or terminal windows:

* **Dual Command Support**: Use `ohio` (primary) or `ohfs` (legacy symlink) interchangeably.
* **Direct Root Flags**:
  * `ohio -u <files...>`: Instant multi-file upload without typing subcommands.
  * `ohio -udrks`: Immediate 3-tier nuclear revocation and scrub.
* **Ergonomic Subcommand Shorthands**:
  * `u` → `upload`
  * `d` → `download`
  * `l` → `ls` / `files`
  * `s` → `session`
  * `t` → `tui`
  * `sh` → `share`
  * `f` → `folder`
* **Unix Standard Stream Piping (`stdin`)**:
  Stream database dumps, logs, or encrypted text directly into a remote OhioFiles slug:
  ```bash
  cat report.txt | ohio
  cat notes.md | ohio u --name notes.md --qr
  mysqldump db_prod | gzip -9 | ohio u --name backup.sql.gz --one-time
  ```
* **Dev Auto-Upload Watch Mode (`-w` / `--watch`)**:
  Continuous file monitoring with delta uploads triggers automatically in under `<200ms` when saving changes:
  ```bash
  ohio u bundle.js -w
  ```
* **Instant Terminal QR Codes (`--qr`)**:
  Generates crisp, high-contrast Unicode/ASCII QR codes directly in your terminal output for instant phone transfers.

---

## 🖥️ Interactive TUI Suite (`ohio tui`)

Launch OhioCLI's terminal user interface by running `ohio tui` (or shorthand `ohio t`). Built with Charm's Bubble Tea framework, it provides an ultra-responsive, full-screen dashboard.

```bash
ohio tui
# or
ohio t
```

```text
┌─ ◆ ohio v2.2.0     👤 GopherPilot (ohio_FastMonkey252)                   ● ONLINE ─┐
│                                                                                    │
│   [ 1 Files ]   2 Folders   3 Upload   4 Feed   5 Session   6 Settings/Help        │
│                                                                                    │
│   FILENAME                  LOCATION    SLUG       SIZE       VISIBILITY  DLS      │
│  ▸architecture-spec.pdf     / (root)    arch-v2    2.4 MB     [Private]   14       │
│   database-backup.sql.gz    📁 backups  db-prod-82 84.1 MB    [1-Time]    0        │
│   presentation-slides.key   📁 assets   keynote-q3 18.9 MB    [Public]    88       │
│                                                                                    │
│  [Tab] Switch tabs  •  [1-6] Jump  •  [/] Filter  •  [s] Share/QR  •  [q] Quit     │
└────────────────────────────────────────────────────────────────────────────────────┘
```

### Detailed Tab Walkthrough

1. **📁 Tab 1: Files Explorer**
   * View all active uploads tied to your public session key.
   * Explicit **LOCATION** tracking: reveals whether files reside in `/ (root)` or a specific subfolder (`📁 <folder>`).
   * Visual indicators for **Public**, **Private**, **1-Time Burn**, and **Password Protected** files.
   * Real-time metrics: Download counts, human-readable file sizes, and relative upload timestamps.
   * Interactive search filter activated with `/` to match filenames or slugs on the fly.
   * Direct actions: download (`d`), share / QR modal (`s` or `Enter`), or delete (`x`).

2. **🗂️ Tab 2: Folders Explorer**
   * Full session folder tree management.
   * Browse folders with metadata: slug, visibility status, creation timestamps.
   * Jump into any folder by pressing `Enter` to filter the **Files Explorer** exclusively to that folder.
   * In-line quick folder creation (`n`).
   * One-touch streaming ZIP archive download (`z`) for entire folder contents.
   * Direct deletion of empty or obsolete folders (`x`).

3. **📤 Tab 3: Quick Upload**
   * Clean multi-field form supporting local file paths with auto-expansion.
   * Optional custom slug input and optional password lock.
   * Instant radio toggles for **Public Visibility** (`p`) and **One-Time Burn** (`o`).
   * Real-time, animated streaming progress bar displaying percentage, bytes transferred, and moving-average upload speed.

4. **🌐 Tab 4: Public Feed**
   * Live streaming discovery of files published publicly across OhioFiles.
   * Incremental pagination and real-time search filtering (`/`).
   * One-click download (`d`) or terminal QR view (`s` or `Enter`) without registering accounts.

5. **👤 Tab 5: Sessions & Identity**
   * Displays active session key, current display nickname, and public profile slug.
   * Live API endpoint connectivity indicator and device fingerprint ID.
   * Generate memorable human keys (`n`) or cryptographic 192-bit secure keys (`s`).
   * Edit display nickname inline (`e`) or copy session key straight to clipboard (`c`).

6. **⚙️ Tab 6: Settings & Help**
   * Two-column split reference showing all dashboard hotkeys alongside CLI commands.
   * Persistent status feedback for connected server endpoints and local config states.

### Complete Keyboard Shortcut Table

| Keybinding | Scope | Description |
| :--- | :--- | :--- |
| `Tab` / `Shift+Tab` | Global | Cycle forward / backward through tabs (1 through 6) |
| `1` – `6` | Global | Instantly switch to specific tab (Files, Folders, Upload, Feed, Session, Help) |
| `↑` / `↓` or `k` / `j` | List Views | Move cursor selection up / down |
| `g` / `G` | List Views | Jump cursor to top / bottom of current list |
| `/` | Files / Feed | Activate real-time fuzzy filter / search bar |
| `Enter` | Files / Feed | Open Share Modal with ASCII QR Code and direct URL |
| `Enter` | Folders Tab | Open selected folder in Files Explorer tab |
| `Enter` | Upload Tab | Toggle visibility, one-time burn, or trigger file upload |
| `d` | Files / Feed | Download currently highlighted file to working directory |
| `s` | Files / Feed | Show share card and ASCII QR modal |
| `z` | Folders Tab | Download entire folder as streaming compressed ZIP archive |
| `n` | Folders Tab | Create a new folder in session |
| `c` | Session Tab | Copy active session key to system clipboard |
| `e` | Session Tab | Edit and rename session nickname inline |
| `n` | Session Tab | Generate and activate a new **memorable** session key |
| `s` | Session Tab | Generate and activate a new **192-bit secure** cryptographic key |
| `x` | Files / Folders | Permanently delete the highlighted file or folder |
| `p` / `o` | Upload Tab | Toggle **Public** visibility (`p`) / **One-Time** burn (`o`) |
| `r` | All Tabs | Refresh remote file or folder listings |
| `Esc` | Modal / Search | Close share QR modal or cancel search filter input |
| `q` / `Ctrl+C` | Global | Quit the interactive TUI application |

---

## 📖 CLI Command Reference

OhioCLI supports a rich set of subcommands and flags for scripted environments and day-to-day command-line operations.

### 1. Uploading Files & Streams (`upload` / `u`)

```bash
# Basic upload
ohio upload document.pdf
ohio u document.pdf

# Root shortcut (upload without typing subcommand)
ohio -u image.png data.csv

# Upload with custom slug & in-terminal QR code
ohio u archive.tar.gz --slug my-backup --qr

# Public file (visible on global public feed)
ohio u release.zip --public

# One-time burn file (deleted immediately after first view or download)
ohio u credentials.env --one-time

# Password-protected upload
ohio u contract.pdf --password "OhioVault2026!"

# Upload into a specific folder
ohio u report.xlsx --folder fld_98a72b1c

# Unix standard stream piping (stdin)
cat dump.sql | ohio
tail -n 200 /var/log/nginx/access.log | ohio u --name nginx-tail.log --qr

# Dev Watch Mode (auto-uploads upon file save in <200ms)
ohio u main.wasm -w
```

#### Upload Flags Summary
| Flag | Shorthand | Type | Description |
| :--- | :---: | :---: | :--- |
| `--public` | `-p` | bool | Make file discoverable on the public feed |
| `--one-time` | `-o` | bool | Self-destruct / burn file after first download |
| `--slug` | `-s` | string | Assign custom vanity URL slug |
| `--password` | | string | Lock file with password protection |
| `--folder` | `-f` | string | Target folder ID to contain the file |
| `--watch` | `-w` | bool | Watch local file and auto-upload on save |
| `--qr` | | bool | Render ASCII QR code in terminal upon completion |
| `--name` | `-n` | string | Explicit filename when streaming via stdin |
| `--chunked` | | bool | Force chunked upload protocol (for large files) |

---

### 2. Exploring & Downloading (`ls`, `download`, `info`, `share`)

```bash
# List files in your active session
ohio ls
ohio l

# List public feed files
ohio ls --public --limit 30

# Download file with progress bar and resume support
ohio download my-backup -o ./restored.tar.gz
ohio d my-backup

# Download password-protected file
ohio d secret-doc --password "OhioVault2026!"

# Inspect file metadata (size, mime, downloads, dates, QR)
ohio info my-backup

# Display share links, markdown tags, and terminal QR
ohio share my-backup
ohio sh my-backup

# Delete a file
ohio delete my-backup
ohio rm my-backup
```

---

### 3. Session & Identity Management (`session` / `s`)

Sessions in OhioFiles give you seamless ownership of files without requiring emails or passwords.

```bash
# Display currently active session and device ID
ohio session whoami
ohio s whoami

# Generate a human-memorable key (e.g. ohio_SmoothDuck137)
ohio s new --type memorable

# Generate a high-entropy 192-bit cryptographic key
ohio s new --type secure
# or shorthand:
ohio s new --secure

# Update session display nickname
ohio s nickname "CloudArchitect"

# Switch to an existing session key
ohio s set "ohio_FastMonkey252"

# Clear active session from local config
ohio s clear
```

---

### 4. Folder Operations (`folder` / `f`)

Organize files into collections and stream entire directory hierarchies as a single compressed ZIP file.

```bash
# Create a new folder
ohio folder create "DevOpsRelease"
ohio f create "PublicAssets" --public --slug assets

# List all folders in your session
ohio f ls

# Stream and download an entire folder hierarchy as a ZIP archive
ohio f zip <folder-id-or-slug> -o devops-release.zip

# Delete a folder
ohio f delete <folder-id>
ohio f rm <folder-id>
```

---

## 🛡️ Server Lifecycle & 3-Tier Nuclear Protection

OhioCLI provides built-in update verification, graceful removal, and military-grade data sanitization.

### Update OhioCLI
```bash
ohio update
```
Checks for latest compiled binaries from `api.ohiofiles.cloud` and updates in-place.

### Soft Uninstall (Preserves User State)
```bash
ohio uninstall
```
Removes the `ohio` and `ohfs` binaries from `/usr/local/bin` (or `~/.local/bin`), while **safely preserving** your session keys and local cache at `~/.config/ohio/config.json`. When you reinstall later, your active session and files are instantly available.

### Nuclear Wipe (`ohio -udrks` / `ohio udrks`)
For air-gapped environments, shared workstations, or ephemeral cloud nodes, the **UDRKS** protocol (*Uninstall-Delete-Revoke-Kill-Service*) provides a complete 3-tier purge.

To prevent accidental wipe, UDRKS enforces a **3-step interactive confirmation sequence**:
1. `[Step 1/3]` Confirm initiation (`y` / `yes`)
2. `[Step 2/3]` Permanent and irreversible warning (type `CONFIRM` or `YES`)
3. `[Step 3/3]` Final safeguard to execute purge (type `PURGE`)

For non-interactive automation or headless scripts, pass `--force` / `-f` to bypass interactive prompts:
```bash
ohio -udrks --force
# or
ohio udrks -f
# or
ohio uninstall --udrks --force
```

```text
🛡️  OhioCLI Nuclear Scrub Protocol (UDRKS)
Executing 3 tiers of security protection...

  [1/3] ⚡ Revoking public session ohio_FastMonkey252 on server...
        ✔ Session revoked and all associated files destroyed.
  [2/3] ⚡ Terminating running OhioCLI background services...
        ✔ Background services and watchers terminated.
  [3/3] ⚡ Purging local caches, configurations, and binaries...
        ✔ Local cache, credentials, and binaries removed.

✔ System completely scrubbed of OhioCLI. Zero traces remaining.
```

1. **Tier 1 (Server Revocation)**: Issues an authenticated cryptographic wipe to OhioFiles servers, permanently invalidating the session key and unlinking all associated files and folders.
2. **Tier 2 (Process Execution Termination)**: Discovers and sends `SIGTERM` signals to all active OhioCLI background daemons, watch mode watchers, and spawned processes.
3. **Tier 3 (Local Artifact Scrub)**: Removes `~/.config/ohio`, `~/.config/ohfs`, `~/.cache/ohio`, temporary stdin spool files, and all executable binary links from the system PATH.

---

## 🐚 Shell Completions

Generate fast shell auto-completions for subcommands, options, and flags.

### Bash
```bash
# Current session:
source <(ohio completion bash)

# Persist for Linux:
ohio completion bash | sudo tee /etc/bash_completion.d/ohio > /dev/null

# Persist for macOS (Homebrew):
ohio completion bash > $(brew --prefix)/etc/bash_completion.d/ohio
```

### Zsh
```zsh
# Current session:
source <(ohio completion zsh)

# Persist:
mkdir -p ~/.zsh/completion
ohio completion zsh > ~/.zsh/completion/_ohio
# Ensure ~/.zshrc contains:
# fpath=(~/.zsh/completion $fpath)
# autoload -Uz compinit && compinit
```

### Fish
```fish
ohio completion fish | source
# Persist:
ohio completion fish > ~/.config/fish/completions/ohio.fish
```

### PowerShell (Windows)
```powershell
ohio completion powershell | Out-String | Invoke-Expression

# Persist to profile:
New-Item -Type File -Force $PROFILE
Add-Content $PROFILE "ohio completion powershell | Out-String | Invoke-Expression"
```

---

## ⚙️ Environment Variables & Configuration

OhioCLI persists its active profile in standard XDG configuration paths, with full override capabilities via environment variables and CLI flags.

### Configuration File: `~/.config/ohio/config.json`

```json
{
  "apiUrl": "https://api.ohiofiles.cloud",
  "sessionKey": "ohio_FastMonkey252",
  "nickname": "GopherPilot",
  "profileSlug": "gopher-pilot",
  "deviceId": "cli_e3a8910b2f76c54a89d1234567890abc"
}
```

*(Note: Legacy configurations located at `~/.config/ohfs/config.json` are automatically detected and honored).*

### Environment Variables

| Variable | Default Value | Description |
| :--- | :--- | :--- |
| `OHIO_API_URL` | `https://api.ohiofiles.cloud` | Custom OhioFiles backend API endpoint |
| `OHIO_SESSION_KEY` | *(Configured session)* | Temporary session key override for CI/CD runs |
| `OHFS_API_URL` | *(Fallback for legacy)* | Backward-compatible environment endpoint |
| `OHFS_SESSION_KEY` | *(Fallback for legacy)* | Backward-compatible session key override |

### Global Flags Priority
Command flags take highest precedence, followed by environment variables, followed by `config.json`:
```bash
# Run command against self-hosted instance without altering local config
ohio ls --api-url "https://files.internal.lan" --session-key "ohio_CustomKey99"

# Format all command output as JSON for scripting with jq
ohio ls --json | jq '.[].filename'

# Quiet mode for silent background automation
ohio -u release.tar.gz --quiet
```

---

## 📄 License & Community

OhioCLI is open-source software licensed under the **[MIT License](LICENSE)**.  
Built with ❤️ for privacy, simplicity, and blazing terminal speed.
