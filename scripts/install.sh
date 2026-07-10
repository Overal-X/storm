#!/bin/bash
set -euo pipefail

# Resolve the latest version from GitHub Releases API when no explicit version is provided
LATEST_RELEASE_API="https://api.github.com/repos/Overal-X/formatio.storm/releases/latest"

resolve_latest_version() {
    local latest_version
    latest_version=$(curl -fsSL "$LATEST_RELEASE_API" | sed -n 's/.*"tag_name":[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)

    if [ -z "$latest_version" ]; then
        echo "Failed to resolve latest release version from GitHub API" >&2
        exit 1
    fi

    echo "$latest_version"
}

# Get the version from the command line argument or resolve latest
if [ -n "${1:-}" ]; then
    VERSION="${1}"
else
    VERSION=$(resolve_latest_version)
fi

# Define the base URL for the release artifacts
BASE_URL="https://github.com/overal-x/storm/releases/download/${VERSION}"

# Define the file names (adjust these as needed)
LINUX_AMD64="storm_Linux_x86_64.tar.gz"
LINUX_ARM64="storm_Linux_arm64.tar.gz"
MACOS_AMD64="storm_Darwin_x86_64.tar.gz"
MACOS_ARM64="storm_Darwin_arm64.tar.gz"

# Determine the OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

# Set the file to download based on OS and architecture
case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64)
                FILE="$LINUX_AMD64"
                ;;
            aarch64)
                FILE="$LINUX_ARM64"
                ;;
            *)
                echo "Unsupported architecture: $ARCH"
                exit 1
                ;;
        esac
        ;;
    Darwin)
        case "$ARCH" in
            x86_64)
                FILE="$MACOS_AMD64"
                ;;
            arm64)
                FILE="$MACOS_ARM64"
                ;;
            *)
                echo "Unsupported architecture: $ARCH"
                exit 1
                ;;
        esac
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Define the destination directory
DEST_DIR="$HOME/.storm/bin"

# Create the destination directory if it does not exist
mkdir -p "$DEST_DIR"

# Download the file
echo "Using version: $VERSION"
echo "Downloading $FILE..."
if ! curl -fsSL "${BASE_URL}/${FILE}" -o "${FILE}"; then
    echo "Error: failed to download ${BASE_URL}/${FILE}" >&2
    exit 1
fi

# Extract the downloaded file to the .storm directory
echo "Extracting $FILE to $DEST_DIR..."
if ! tar -xzf "$FILE" -C "$DEST_DIR"; then
    echo "Error: failed to extract $FILE" >&2
    rm -f "$FILE"
    exit 1
fi

# Remove the downloaded file
rm "$FILE"

# Add .storm directory to PATH if not already present
PATH_ENTRY="export PATH=\"$DEST_DIR:\$PATH\""
PROFILE_FILE="$HOME/.bashrc"

# Check if we are running on Alpine Linux
if [ -f /etc/alpine-release ]; then
    PROFILE_FILE="$HOME/.profile"
fi

if ! grep -Fxq "$PATH_ENTRY" "$PROFILE_FILE"; then
    echo "$PATH_ENTRY" >> "$PROFILE_FILE"
    echo "Updated $PROFILE_FILE to include $DEST_DIR in PATH"
else
    echo "$DEST_DIR is already in PATH in $PROFILE_FILE"
fi
