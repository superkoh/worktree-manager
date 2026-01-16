#!/bin/bash
set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="${REPO:-superkoh/worktree-manager}"
BINARY_NAME="wt"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

print_banner() {
    echo -e "${BLUE}"
    echo "╦ ╦╔╦╗  ╦╔╗╔╔═╗╔╦╗╔═╗╦  ╦  ╔═╗╦═╗"
    echo "║║║ ║   ║║║║╚═╗ ║ ╠═╣║  ║  ║╣ ╠╦╝"
    echo "╚╩╝ ╩   ╩╝╚╝╚═╝ ╩ ╩ ╩╩═╝╩═╝╚═╝╩╚═"
    echo -e "${NC}"
    echo "Git Worktree Manager"
    echo ""
}

# Detect OS
detect_os() {
    local os
    case "$(uname -s)" in
        Linux*)  os=linux;;
        Darwin*) os=darwin;;
        MINGW*|MSYS*|CYGWIN*) os=windows;;
        *)
            echo -e "${RED}Unsupported OS: $(uname -s)${NC}"
            exit 1
            ;;
    esac
    echo "$os"
}

# Detect architecture
detect_arch() {
    local arch
    case "$(uname -m)" in
        x86_64|amd64)  arch=amd64;;
        arm64|aarch64) arch=arm64;;
        *)
            echo -e "${RED}Unsupported architecture: $(uname -m)${NC}"
            exit 1
            ;;
    esac
    echo "$arch"
}

# Get latest version from GitHub
get_latest_version() {
    local version
    version=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        echo -e "${RED}Failed to get latest version. Using 'latest'.${NC}" >&2
        echo "latest"
    else
        echo "$version"
    fi
}

# Setup shell integration
setup_shell_integration() {
    local shell_rc=""

    # Detect shell
    case "$SHELL" in
        */bash) shell_rc="$HOME/.bashrc";;
        */zsh)  shell_rc="$HOME/.zshrc";;
        *)      return;;
    esac

    local start_marker="# wt - Git Worktree Manager shell integration"
    local end_marker="# wt - end"

    echo ""

    # Remove existing integration if present (supports upgrade)
    if grep -q "$start_marker" "$shell_rc" 2>/dev/null; then
        if grep -q "$end_marker" "$shell_rc" 2>/dev/null; then
            # Both markers present - safe to delete the range
            sed -i.bak "/$start_marker/,/$end_marker/d" "$shell_rc"
            rm -f "${shell_rc}.bak"
            echo -e "${BLUE}Updating shell integration...${NC}"
        else
            # Start marker without end marker - legacy block, warn user
            echo -e "${YELLOW}Warning: Found shell integration without end marker (legacy install).${NC}"
            echo -e "${YELLOW}Please manually remove the wt() block from $shell_rc and re-run this script.${NC}"
            return
        fi
    elif grep -q "wt()" "$shell_rc" 2>/dev/null; then
        # Old format without markers - warn user
        echo -e "${YELLOW}Warning: Found old wt() function without markers.${NC}"
        echo -e "${YELLOW}Please manually remove it from $shell_rc and re-run this script.${NC}"
        return
    else
        echo -e "${BLUE}Setting up shell integration...${NC}"
    fi

    # Add new integration with markers
    cat >> "$shell_rc" << 'EOF'

# wt - Git Worktree Manager shell integration
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
# wt - end
EOF

    echo -e "${GREEN}Shell integration added to $shell_rc${NC}"
    echo -e "${YELLOW}Please run: source $shell_rc${NC}"
}

# Main installation
main() {
    print_banner

    local os=$(detect_os)
    local arch=$(detect_arch)
    local version="${VERSION:-$(get_latest_version)}"
    local current_version=""
    local is_upgrade=false

    # Check existing installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        current_version=$("$BINARY_NAME" version 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
        is_upgrade=true
    fi

    echo "System:  $os/$arch"
    if [ "$is_upgrade" = true ]; then
        echo -e "Current: ${YELLOW}${current_version}${NC}"
        echo -e "Target:  ${GREEN}${version}${NC}"
    else
        echo "Version: $version"
    fi
    echo "Path:    $INSTALL_DIR/$BINARY_NAME"
    echo ""

    # Create install directory
    mkdir -p "$INSTALL_DIR"

    # Download URL
    local ext="tar.gz"
    if [ "$os" = "windows" ]; then
        ext="zip"
    fi

    local filename="${BINARY_NAME}_${version#v}_${os}_${arch}.${ext}"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"

    echo -e "${BLUE}Downloading...${NC}"
    echo "  $download_url"

    # Create temp directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    # Download
    if ! curl -sL "$download_url" -o "$tmp_dir/$filename"; then
        echo -e "${RED}Download failed${NC}"
        exit 1
    fi

    # Extract
    echo -e "${BLUE}Extracting...${NC}"
    if [ "$ext" = "tar.gz" ]; then
        tar -xzf "$tmp_dir/$filename" -C "$tmp_dir"
    else
        unzip -q "$tmp_dir/$filename" -d "$tmp_dir"
    fi

    # Install binary
    mv "$tmp_dir/$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"

    if [ "$is_upgrade" = true ]; then
        echo -e "${GREEN}Successfully upgraded $BINARY_NAME${NC}"
    else
        echo -e "${GREEN}Successfully installed $BINARY_NAME to $INSTALL_DIR${NC}"
    fi

    # Check PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo ""
        echo -e "${YELLOW}Note: $INSTALL_DIR is not in your PATH${NC}"
        echo ""
        echo "Add this to your shell profile:"
        echo ""
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi

    # Setup shell integration
    setup_shell_integration

    # Verify installation
    echo ""
    if command -v "$BINARY_NAME" &> /dev/null; then
        echo -e "${GREEN}Verification:${NC}"
        "$INSTALL_DIR/$BINARY_NAME" version
    else
        echo -e "${YELLOW}Installed. Restart your shell or run:${NC}"
        echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    fi

    echo ""
    if [ "$is_upgrade" = true ]; then
        echo -e "${GREEN}Upgrade complete!${NC}"
    else
        echo -e "${GREEN}Installation complete!${NC}"
        echo ""
        echo "Quick start:"
        echo "  cd your-git-repo"
        echo "  wt init          # Create .wt.json config"
        echo "  wt add -b feat   # Create worktree with new branch"
        echo "  wt list          # List all worktrees"
        echo "  wt select        # Interactive worktree selection"
    fi
}

main "$@"
