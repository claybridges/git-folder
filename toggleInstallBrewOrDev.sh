#!/bin/bash
set -e

REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"

# Determine local bin directory
if [ -n "$XDG_BIN_HOME" ]; then
    LOCAL_BIN_DIR="$XDG_BIN_HOME"
elif [ -d "$HOME/.local/bin" ]; then
    LOCAL_BIN_DIR="$HOME/.local/bin"
else
    LOCAL_BIN_DIR="$HOME/bin"
fi
LOCAL_BIN="$LOCAL_BIN_DIR/git-folder"

# Check if brew version is linked
if brew list --formula git-folder &>/dev/null && \
   brew ls --verbose git-folder 2>/dev/null | grep -q "bin/git-folder"; then
    echo "Switching to local development version..."
    brew unlink git-folder

    # Build local version
    mkdir -p "$(dirname "$LOCAL_BIN")"
    cd "$REPO_ROOT"
    go build -o "$LOCAL_BIN" ./cmd/git-folder

    echo "✓ Using local version at $LOCAL_BIN"
    echo "  Version: $("$LOCAL_BIN" version)"
else
    echo "Switching to brew version..."

    # Remove local version if it exists
    if [ -f "$LOCAL_BIN" ]; then
        rm "$LOCAL_BIN"
        echo "✓ Removed local version"
    fi

    # Relink brew version
    brew link git-folder
    echo "✓ Using brew version"
    echo "  Version: $(git-folder version)"
fi
