#Requires -Version 5.1
<#
.SYNOPSIS
    Install wt - Git Worktree Manager

.DESCRIPTION
    Downloads and installs the wt binary for Windows.
    Optionally sets up PowerShell integration for auto-cd functionality.

.PARAMETER Version
    Specific version to install (e.g., "v1.0.0"). Defaults to latest.

.PARAMETER InstallDir
    Installation directory. Defaults to $env:LOCALAPPDATA\Programs\wt

.PARAMETER NoPath
    Skip adding install directory to PATH.

.PARAMETER NoProfile
    Skip adding shell integration to PowerShell profile.

.EXAMPLE
    iwr -useb https://raw.githubusercontent.com/superkoh/worktree-manager/main/scripts/install.ps1 | iex

.EXAMPLE
    .\install.ps1 -Version v1.0.0 -InstallDir C:\tools\wt
#>

param(
    [string]$Version = "",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\wt",
    [switch]$NoPath,
    [switch]$NoProfile
)

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "superkoh/worktree-manager"
$BinaryName = "wt.exe"

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        return $release.tag_name
    }
    catch {
        Write-ColorOutput "Failed to get latest version: $_" "Yellow"
        return "latest"
    }
}

function Get-Architecture {
    if ([Environment]::Is64BitOperatingSystem) {
        if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
            return "arm64"
        }
        return "amd64"
    }
    throw "Unsupported architecture: 32-bit Windows is not supported"
}

function Install-Wt {
    Write-ColorOutput "`n=====================================" "Cyan"
    Write-ColorOutput "  wt - Git Worktree Manager Installer" "Cyan"
    Write-ColorOutput "=====================================`n" "Cyan"

    # Detect architecture
    $arch = Get-Architecture
    Write-ColorOutput "Architecture: windows/$arch" "Gray"

    # Get version
    if ([string]::IsNullOrEmpty($Version)) {
        Write-ColorOutput "Fetching latest version..." "Gray"
        $Version = Get-LatestVersion
    }
    Write-ColorOutput "Version: $Version" "Gray"
    Write-ColorOutput "Install directory: $InstallDir" "Gray"

    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Download URL
    $versionNum = $Version -replace "^v", ""
    $filename = "wt_${versionNum}_windows_${arch}.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$filename"

    Write-ColorOutput "`nDownloading from:" "White"
    Write-ColorOutput "  $downloadUrl" "Gray"

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "wt_install_$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        # Download
        $zipPath = Join-Path $tempDir $filename
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing

        Write-ColorOutput "Extracting..." "White"

        # Extract
        Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

        # Move binary
        $binaryPath = Join-Path $tempDir "wt.exe"
        if (-not (Test-Path $binaryPath)) {
            # Try looking in subdirectory
            $binaryPath = Get-ChildItem -Path $tempDir -Recurse -Filter "wt.exe" | Select-Object -First 1 -ExpandProperty FullName
        }

        if (-not $binaryPath -or -not (Test-Path $binaryPath)) {
            throw "wt.exe not found in archive"
        }

        Copy-Item -Path $binaryPath -Destination (Join-Path $InstallDir $BinaryName) -Force

        Write-ColorOutput "`nInstalled to: $InstallDir\$BinaryName" "Green"
    }
    finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }

    # Add to PATH
    if (-not $NoPath) {
        $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($currentPath -notlike "*$InstallDir*") {
            Write-ColorOutput "`nAdding to PATH..." "White"
            [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
            $env:Path = "$env:Path;$InstallDir"
            Write-ColorOutput "Added $InstallDir to user PATH" "Green"
        }
        else {
            Write-ColorOutput "Already in PATH" "Gray"
        }
    }

    # Setup PowerShell integration
    if (-not $NoProfile) {
        Setup-PowerShellIntegration
    }

    # Verify installation
    Write-ColorOutput "`nVerifying installation..." "White"
    try {
        $wtPath = Join-Path $InstallDir $BinaryName
        $versionOutput = & $wtPath version 2>&1
        Write-ColorOutput $versionOutput "Gray"
    }
    catch {
        Write-ColorOutput "Installed. Restart your terminal to use 'wt' command." "Yellow"
    }

    Write-ColorOutput "`n=====================================" "Green"
    Write-ColorOutput "  Installation complete!" "Green"
    Write-ColorOutput "=====================================`n" "Green"

    Write-ColorOutput "Quick start:" "White"
    Write-ColorOutput "  cd your-git-repo" "Gray"
    Write-ColorOutput "  wt init          # Create .wt.json config" "Gray"
    Write-ColorOutput "  wt add -b feat   # Create worktree with new branch" "Gray"
    Write-ColorOutput "  wt list          # List all worktrees" "Gray"
    Write-ColorOutput "  wt select        # Interactive worktree selection" "Gray"
}

function Setup-PowerShellIntegration {
    $profilePath = $PROFILE.CurrentUserAllHosts

    # Check if profile exists
    if (-not (Test-Path $profilePath)) {
        $profileDir = Split-Path $profilePath -Parent
        if (-not (Test-Path $profileDir)) {
            New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
        }
        New-Item -ItemType File -Path $profilePath -Force | Out-Null
    }

    # Check if already configured
    $profileContent = Get-Content $profilePath -Raw -ErrorAction SilentlyContinue
    if ($profileContent -and $profileContent -match "wt-shell-integration") {
        Write-ColorOutput "PowerShell integration already configured" "Gray"
        return
    }

    Write-ColorOutput "`nSetting up PowerShell integration..." "White"

    $integration = @'

# wt - Git Worktree Manager shell integration
# wt-shell-integration
function Invoke-Wt {
    param([Parameter(ValueFromRemainingArguments)]$Args)

    if ($Args.Count -gt 0 -and ($Args[0] -eq "add" -or $Args[0] -eq "select")) {
        $allArgs = $Args + @("--print-path")
        $output = & wt.exe @allArgs 2>&1
        $exitCode = $LASTEXITCODE

        if ($exitCode -eq 0 -and $output -and (Test-Path $output -PathType Container)) {
            Set-Location $output
            Write-Host "Switched to: $output" -ForegroundColor Green
        }
        else {
            Write-Output $output
        }
    }
    else {
        & wt.exe @Args
    }
}

Set-Alias -Name wt -Value Invoke-Wt -Scope Global -Force
'@

    Add-Content -Path $profilePath -Value $integration
    Write-ColorOutput "PowerShell integration added to: $profilePath" "Green"
    Write-ColorOutput "Restart PowerShell or run: . `$PROFILE" "Yellow"
}

# Run installation
Install-Wt
