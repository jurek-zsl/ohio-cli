# ==============================================================================
# OhioFiles CLI (ohio) One-Line Installer for Windows PowerShell
# Usage: irm https://ohiofiles.cloud/install.ps1 | iex
# ==============================================================================

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "⚡ OhioFiles CLI (ohio) Windows Installer" -ForegroundColor Cyan
Write-Host "Fast, anonymous, login-free file sharing & TUI" -ForegroundColor DarkGray
Write-Host ""

# 1. Determine Target Directory
$InstallDir = "$env:LOCALAPPDATA\Programs\ohio"
if (!(Test-Path -Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$OhioExe = Join-Path $InstallDir "ohio.exe"
$OhfsExe = Join-Path $InstallDir "ohfs.exe"

# 2. Check for local build or download remote binary
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path 2>$null
$LocalBinary = $null

if ($ScriptDir) {
    if (Test-Path "$ScriptDir\cli\bin\ohio.exe") {
        $LocalBinary = "$ScriptDir\cli\bin\ohio.exe"
    } elseif (Test-Path "$ScriptDir\bin\ohio.exe") {
        $LocalBinary = "$ScriptDir\bin\ohio.exe"
    } elseif (Test-Path "$ScriptDir\ohio.exe") {
        $LocalBinary = "$ScriptDir\ohio.exe"
    }
}

if ($LocalBinary) {
    Write-Host "• Installing from local binary ($LocalBinary)..." -ForegroundColor Gray
    Copy-Item -Path $LocalBinary -Destination $OhioExe -Force
} else {
    $Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $DownloadUrl = "https://api.ohiofiles.cloud/releases/latest/ohio-windows-$Arch.exe"
    Write-Host "• Downloading OhioCLI ($Arch) from $DownloadUrl..." -ForegroundColor Gray

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $OhioExe -UseBasicParsing
    } catch {
        # Check if go compiler is available
        if (Get-Command go -ErrorAction SilentlyContinue) {
            Write-Host "• Remote binary unavailable, compiling using local Go installation..." -ForegroundColor Gray
            Push-Location "$ScriptDir\cli"
            go build -o $OhioExe .\cmd\ohfs
            Pop-Location
        } else {
            Write-Error "Failed to download OhioCLI binary: $_"
            exit 1
        }
    }
}

# Create ohfs.exe copy/link
Copy-Item -Path $OhioExe -Destination $OhfsExe -Force

# 3. Add to User PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "• Adding $InstallDir to User PATH..." -ForegroundColor Gray
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path = "$env:Path;$InstallDir"
}

Write-Host ""
Write-Host "✔ Successfully installed OhioCLI!" -ForegroundColor Green
Write-Host "  Location: $OhioExe (alias: ohfs.exe)"
Write-Host "  Version:  v2.1.0"
Write-Host ""
Write-Host "Quick Start:" -ForegroundColor White
Write-Host "  ohio tui              Launch interactive TUI dashboard" -ForegroundColor Cyan
Write-Host "  ohio -u file.txt      Quickly upload a file" -ForegroundColor Cyan
Write-Host "  ohio --help           View all available commands" -ForegroundColor Cyan
Write-Host ""
