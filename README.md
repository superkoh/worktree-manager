# wt - Git Worktree Manager

A cross-platform CLI tool for managing Git worktrees with ease. Features interactive TUI selection, automatic file copying/linking, and shell integration for seamless workflow.

**Supported Platforms:** Linux, macOS, Windows

## Installation

### Linux / macOS (curl)

```bash
curl -sSL https://raw.githubusercontent.com/superkoh/worktree-manager/main/scripts/install.sh | bash
```

### Windows (PowerShell)

```powershell
iwr -useb https://raw.githubusercontent.com/superkoh/worktree-manager/main/scripts/install.ps1 | iex
```

Or with options:
```powershell
# Install specific version
.\install.ps1 -Version v1.0.0

# Custom install directory
.\install.ps1 -InstallDir C:\tools\wt

# Skip PATH modification
.\install.ps1 -NoPath
```

### From Source

```bash
git clone https://github.com/superkoh/worktree-manager.git
cd wt
make install
```

### Go Install

```bash
go install github.com/superkoh/worktree-manager/cmd/wt@latest
```

## Quick Start

```bash
# Initialize configuration in your repository
cd your-repo
wt init

# Create a worktree with a new branch
wt add -b feature/my-feature

# Create a worktree from existing branch
wt add main

# Interactive branch selection
wt add

# List all worktrees
wt list

# Interactive worktree selection
wt select

# Remove a worktree
wt remove ../your-repo-feature-my-feature

# Clean up stale worktree references
wt prune
```

## Configuration (.wt.json)

Create a `.wt.json` file in your repository root:

```json
{
  "version": "1.0",
  "worktree": {
    "basedir": "../",
    "naming": "{repo}-{branch}",
    "sanitize": { "/": "-", ":": "-" }
  },
  "setup": {
    "copy": [".env", ".env.local"],
    "link": ["node_modules", "vendor", ".idea"]
  }
}
```

### Configuration Options

| Field | Description | Default |
|-------|-------------|---------|
| `worktree.basedir` | Directory for new worktrees | `../` (sibling to repo) |
| `worktree.naming` | Naming template | `{repo}-{branch}` |
| `worktree.sanitize` | Character replacements | `{"/": "-"}` |
| `setup.copy` | Files to copy to new worktrees | `[]` |
| `setup.link` | Paths to symlink to new worktrees | `[]` |

> **Note (Windows):** If symlinks fail due to permission issues, `wt` automatically falls back to copying files instead.

## Shell Integration

Shell integration enables automatic `cd` to new worktrees after `wt add` or `wt select`.

### Bash / Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
wt() {
    if [ "$1" = "add" ] || [ "$1" = "select" ]; then
        local output
        output=$(command wt "$@" --print-path 2>&1)
        local exit_code=$?
        if [ $exit_code -eq 0 ] && [ -n "$output" ] && [ -d "$output" ]; then
            cd "$output" && echo "Switched to: $output"
        else
            echo "$output"
            return $exit_code
        fi
    else
        command wt "$@"
    fi
}
```

### PowerShell

Add to your PowerShell profile (`$PROFILE`):

```powershell
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
```

> **Tip:** The install scripts automatically set up shell integration for you.

## Commands

| Command | Description |
|---------|-------------|
| `wt add [branch]` | Create a new worktree |
| `wt add -b <branch>` | Create worktree with new branch |
| `wt remove <path>` | Remove a worktree |
| `wt remove -D <path>` | Remove worktree and delete branch |
| `wt list` | List all worktrees |
| `wt list --json` | List in JSON format |
| `wt select` | Interactive worktree selector |
| `wt prune` | Remove stale worktree references |
| `wt prune --dry-run` | Preview what would be pruned |
| `wt init` | Create .wt.json configuration |
| `wt version` | Show version information |

## Platform Notes

### Windows

- **Symlinks:** Windows requires Developer Mode or Administrator privileges for symlinks. If unavailable, `wt` automatically copies files instead.
- **Git Bash:** The bash install script works in Git Bash, MSYS2, or WSL.
- **PowerShell:** Use the PowerShell install script for native Windows experience.

### macOS / Linux

- Full symlink support
- Bash and Zsh shell integration

## Why wt?

- **Cross-platform** - Works on Windows, macOS, and Linux
- **Interactive TUI** - Fuzzy search and select branches/worktrees
- **Auto Setup** - Automatically copy `.env` files and symlink `node_modules`
- **Shell Integration** - Auto-cd to new worktrees
- **Simple Config** - One `.wt.json` file per repository
- **Fast** - Single binary, no runtime dependencies

## License

MIT
