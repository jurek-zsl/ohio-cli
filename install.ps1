# ==============================================================================
# OhioFiles CLI (ohio) One-Line Installer for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/jurek-zsl/ohio-cli/main/install.ps1 | iex
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
$ScriptDir = $null
if ($MyInvocation -and $MyInvocation.MyCommand -and $MyInvocation.MyCommand.Path) {
    try {
        $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path -ErrorAction SilentlyContinue
    } catch {
        $ScriptDir = $null
    }
}

$LocalBinary = $null

if ($ScriptDir) {
    if (Test-Path (Join-Path $ScriptDir "bin\ohio-windows-amd64.exe")) {
        $LocalBinary = Join-Path $ScriptDir "bin\ohio-windows-amd64.exe"
    } elseif (Test-Path (Join-Path $ScriptDir "bin\ohio.exe")) {
        $LocalBinary = Join-Path $ScriptDir "bin\ohio.exe"
    } elseif (Test-Path (Join-Path $ScriptDir "ohio.exe")) {
        $LocalBinary = Join-Path $ScriptDir "ohio.exe"
    }
}

if ($LocalBinary) {
    Write-Host "• Installing from local binary ($LocalBinary)..." -ForegroundColor Gray
    Copy-Item -Path $LocalBinary -Destination $OhioExe -Force
} else {
    $Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    
    # Candidate release URLs
    $CandidateUrls = @(
        "https://github.com/jurek-zsl/ohio-cli/releases/latest/download/ohio-windows-$Arch.exe",
        "https://github.com/jurek-zsl/ohio-cli/releases/download/v2.1.0/ohio-windows-$Arch.exe",
        "https://api.ohiofiles.cloud/releases/latest/ohio-windows-$Arch.exe"
    )
    
    $Downloaded = $false
    foreach ($Url in $CandidateUrls) {
        try {
            Write-Host "• Fetching binary from $Url..." -ForegroundColor Gray
            Invoke-WebRequest -Uri $Url -OutFile $OhioExe -UseBasicParsing -ErrorAction Stop
            if ((Test-Path -Path $OhioExe) -and ((Get-Item $OhioExe).Length -gt 500000)) {
                $Downloaded = $true
                break
            }
        } catch {
            # Continue to next candidate
        }
    }

    if (-not $Downloaded) {
        # Check if Go compiler is available on the machine
        if (Get-Command go -ErrorAction SilentlyContinue) {
            Write-Host "• Remote binary unavailable, compiling using local Go..." -ForegroundColor Gray
            $TempBuild = Join-Path $env:TEMP ("ohio-install-" + [Guid]::NewGuid().ToString("N"))
            try {
                git clone --depth 1 https://github.com/jurek-zsl/ohio-cli.git $TempBuild
                Push-Location $TempBuild
                go build -o $OhioExe .\cmd\ohfs
                Pop-Location
                $Downloaded = $true
            } catch {
                Write-Host "• Falling back to go install..." -ForegroundColor Gray
                go install github.com/jurek-zsl/ohio-cli/cmd/ohfs@latest
                $GoPath = Join-Path ([Environment]::GetFolderPath("UserProfile")) "go\bin\ohfs.exe"
                if (Test-Path $GoPath) {
                    Copy-Item -Path $GoPath -Destination $OhioExe -Force
                    $Downloaded = $true
                }
            } finally {
                if (Test-Path $TempBuild) {
                    Remove-Item -Recurse -Force $TempBuild -ErrorAction SilentlyContinue
                }
            }
        }
    }

    if (-not $Downloaded -or !(Test-Path -Path $OhioExe)) {
        Write-Host ""
        Write-Host "✖ Could not download or compile OhioCLI automatically." -ForegroundColor Red
        Write-Host "  Please ensure an internet connection is available or install via Go:" -ForegroundColor DarkGray
        Write-Host "  go install github.com/jurek-zsl/ohio-cli/cmd/ohfs@latest" -ForegroundColor Cyan
        exit 1
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
