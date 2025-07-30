#!/bin/bash

# ADL CLI Installation Script
# This script downloads and installs the latest release of adl CLI from GitHub.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REPO="inference-gateway/adl-cli"
BINARY_NAME="adl"

if [ -n "${INSTALL_DIR:-}" ]; then
    INSTALL_DIR="$INSTALL_DIR"
else
    INSTALL_DIR="$HOME/.local/bin"
fi

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    case "$(uname -s)" in
        Linux*)     os="Linux";;
        Darwin*)    os="Darwin";;
        *)          
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="x86_64";;
        arm64|aarch64)  arch="arm64";;
        *)              
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

# Function to get the latest release version
get_latest_version() {
    local latest_version=""
    
    if command -v curl >/dev/null 2>&1; then
        latest_version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        latest_version=$(wget -qO- "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi
    
    if [ -z "$latest_version" ]; then
        print_error "Failed to fetch latest version information"
        exit 1
    fi
    
    echo "$latest_version"
}

# Function to download and install the binary
install_binary() {
    local version="$1"
    local platform="$2"
    local archive_name="${BINARY_NAME}_${platform}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"
    
    if [[ "$platform" == *"Windows"* ]]; then
        local binary_path="${INSTALL_DIR}/${BINARY_NAME}.exe"
    else
        local binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    print_status "Downloading ${archive_name} ${version}..."
    
    local temp_dir=$(mktemp -d)
    local temp_archive="${temp_dir}/${archive_name}"
    
    if command -v curl >/dev/null 2>&1; then
        if ! curl -L -o "$temp_archive" "$download_url"; then
            print_error "Failed to download archive from ${download_url}"
            rm -rf "$temp_dir"
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -O "$temp_archive" "$download_url"; then
            print_error "Failed to download archive from ${download_url}"
            rm -rf "$temp_dir"
            exit 1
        fi
    fi
    
    if [ ! -s "$temp_archive" ]; then
        print_error "Downloaded file is empty or download failed"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    print_status "Extracting archive..."
    
    if ! tar -xzf "$temp_archive" -C "$temp_dir"; then
        print_error "Failed to extract archive"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    local extracted_binary=""
    if [[ "$platform" == *"Windows"* ]]; then
        extracted_binary=$(find "$temp_dir" -name "${BINARY_NAME}.exe" -type f | head -n1)
    else
        extracted_binary=$(find "$temp_dir" -name "${BINARY_NAME}" -type f | head -n1)
    fi
    
    if [ -z "$extracted_binary" ] || [ ! -f "$extracted_binary" ]; then
        print_error "Binary not found in extracted archive"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    if [ ! -d "$INSTALL_DIR" ]; then
        print_status "Creating install directory: $INSTALL_DIR"
        if ! mkdir -p "$INSTALL_DIR"; then
            print_error "Failed to create install directory: $INSTALL_DIR"
            print_error "Please ensure you have write permissions to this location or set a different INSTALL_DIR"
            rm -rf "$temp_dir"
            exit 1
        fi
    fi
    
    print_status "Installing binary to ${binary_path}..."
    if ! cp "$extracted_binary" "$binary_path"; then
        print_error "Failed to install binary to ${binary_path}"
        print_error "Please ensure you have write permissions to $INSTALL_DIR"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    if ! chmod +x "$binary_path"; then
        print_error "Failed to make binary executable"
        rm -rf "$temp_dir"
        exit 1
    fi
    
    rm -rf "$temp_dir"
    
    print_success "${BINARY_NAME} ${version} installed successfully to ${binary_path}"
}

# Function to verify installation
verify_installation() {
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        local installed_version
        installed_version=$("${BINARY_NAME}" --version 2>/dev/null | head -n1 || echo "unknown")
        print_success "Installation verified! ${BINARY_NAME} is available in PATH"
        print_status "Installed version: ${installed_version}"
        echo
        print_success "🎉 Installation complete! You can now use '${BINARY_NAME}' from anywhere."
    else
        print_warning "Binary installed but not found in PATH."
        echo
        print_status "📍 Binary location: ${INSTALL_DIR}/${BINARY_NAME}"
        print_status "🔧 To use '${BINARY_NAME}' from anywhere, add ${INSTALL_DIR} to your PATH:"
        echo
        echo "   For Bash users, add this to ~/.bashrc:"
        echo "   export PATH=\"\$PATH:${INSTALL_DIR}\""
        echo
        echo "   For Zsh users, add this to ~/.zshrc:"
        echo "   export PATH=\"\$PATH:${INSTALL_DIR}\""
        echo
        echo "   For Fish users, run:"
        echo "   fish_add_path ${INSTALL_DIR}"
        echo
        print_status "💡 Alternatively, you can run the binary directly:"
        echo "   ${INSTALL_DIR}/${BINARY_NAME} --help"
        echo
        print_status "🔄 After updating your shell profile, restart your terminal or run:"
        echo "   source ~/.bashrc  # or ~/.zshrc"
    fi
}

# Main installation process
main() {
    echo "==================================="
    echo "   ADL CLI Installation"
    echo "==================================="
    echo
    
    if [ "$EUID" -eq 0 ]; then
        print_warning "Running as root. This is not recommended for security reasons."
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        print_error "tar is required but not installed. Please install tar and try again."
        exit 1
    fi
    
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        print_error "Either curl or wget is required but neither is installed. Please install one of them."
        exit 1
    fi
    
    local platform
    platform=$(detect_platform)
    print_status "Detected platform: ${platform}"
    
    local version
    if [ -n "$VERSION" ]; then
        version="$VERSION"
        print_status "Using specified version: ${version}"
    else
        print_status "Fetching latest release information..."
        version=$(get_latest_version)
        print_status "Latest version: ${version}"
    fi
    
    install_binary "$version" "$platform"
    
    verify_installation
}

# Handle command line arguments
case "${1:-}" in
    -h|--help)
        echo "ADL CLI Installation Script"
        echo
        echo "Usage: $0 [options]"
        echo
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo "  --version      Install a specific version (e.g., --version v1.0.0)"
        echo
        echo "Environment Variables:"
        echo "  INSTALL_DIR    Installation directory (default: ~/.local/bin)"
        echo
        echo "Examples:"
        echo "  $0                    # Install latest version"
        echo "  $0 --version v1.0.0   # Install specific version"
        echo "  INSTALL_DIR=~/bin $0  # Install to custom directory"
        exit 0
        ;;
    --version)
        if [ -z "$2" ]; then
            print_error "Version not specified. Use: $0 --version v1.0.0"
            exit 1
        fi
        VERSION="$2"
        shift 2
        ;;
    "")
        ;;
    *)
        print_error "Unknown option: $1"
        echo "Use '$0 --help' for usage information."
        exit 1
        ;;
esac

main "$@"
