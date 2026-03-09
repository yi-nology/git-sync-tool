#!/bin/bash

# Desktop Build Setup Script for Git Manage Service
# This script initializes the desktop application build environment

set -e

echo "========================================"
echo "Setting up Desktop Build Environment"
echo "========================================"

# Check if Wails is installed
if ! command -v wails &> /dev/null; then
    echo "Installing Wails..."
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
else
    echo "✓ Wails already installed"
fi

# Add Go bin to PATH if needed
if ! echo $PATH | grep -q "$(go env GOPATH)/bin"; then
    echo "Adding Go bin to PATH..."
    export PATH="$PATH:$(go env GOPATH)/bin"
    echo "Please add this to your shell profile:"
    echo '  export PATH="$PATH:$(go env GOPATH)/bin"'
fi

# Check Wails version
echo ""
echo "Wails version:"
wails version

# Check system dependencies
echo ""
echo "Checking system dependencies..."
wails doctor || true

# Check if wails.json exists
if [ ! -f "wails.json" ]; then
    echo ""
    echo "❌ wails.json not found!"
    echo "Please run this script from the project root directory"
    exit 1
fi

# Build frontend dependencies
echo ""
echo "Installing frontend dependencies..."
if [ -d "frontend" ]; then
    cd frontend
    if [ ! -d "node_modules" ]; then
        npm install
    fi
    cd ..
fi

echo ""
echo "========================================"
echo "✓ Setup Complete!"
echo "========================================"
echo ""
echo "You can now build the desktop application:"
echo ""
echo "  make desktop           # Build for current platform"
echo "  make desktop-darwin    # Build for macOS"
echo "  make desktop-windows   # Build for Windows"
echo "  make desktop-linux     # Build for Linux"
echo "  make desktop-all       # Build for all platforms"
echo ""
