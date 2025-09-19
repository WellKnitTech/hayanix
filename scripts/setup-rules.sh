#!/bin/bash

# Hayanix Rule Setup Script
# This script helps users set up external rule sources

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
HAYANIX_BIN="$PROJECT_DIR/hayanix"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
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

check_hayanix() {
    if [ ! -f "$HAYANIX_BIN" ]; then
        print_error "Hayanix binary not found at $HAYANIX_BIN"
        print_info "Please build hayanix first: make build"
        exit 1
    fi
}

setup_rules() {
    print_info "Setting up Hayanix rule sources..."
    
    # Initialize rule manager
    print_info "Initializing rule manager..."
    "$HAYANIX_BIN" rules list > /dev/null
    
    # Download ChopChopGo rules
    print_info "Downloading ChopChopGo rules..."
    if "$HAYANIX_BIN" rules download --source ChopChopGo; then
        print_success "ChopChopGo rules downloaded successfully"
    else
        print_warning "Failed to download ChopChopGo rules"
    fi
    
    # Download SigmaHQ rules
    print_info "Downloading SigmaHQ rules..."
    if "$HAYANIX_BIN" rules download --source SigmaHQ; then
        print_success "SigmaHQ rules downloaded successfully"
    else
        print_warning "Failed to download SigmaHQ rules"
    fi
    
    print_success "Rule setup completed!"
    print_info "You can now use hayanix to analyze logs with the downloaded rules"
}

show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --help, -h     Show this help message"
    echo "  --list         List available rule sources"
    echo "  --download-all Download all enabled rule sources"
    echo "  --chopchopgo   Download only ChopChopGo rules"
    echo "  --sigmahq      Download only SigmaHQ rules"
    echo ""
    echo "Examples:"
    echo "  $0                    # Setup all default rules"
    echo "  $0 --list            # List available sources"
    echo "  $0 --download-all    # Download all enabled sources"
    echo "  $0 --chopchopgo      # Download only ChopChopGo rules"
}

main() {
    case "${1:-}" in
        --help|-h)
            show_usage
            exit 0
            ;;
        --list)
            check_hayanix
            "$HAYANIX_BIN" rules list
            ;;
        --download-all)
            check_hayanix
            print_info "Downloading all enabled rule sources..."
            "$HAYANIX_BIN" rules download --all
            ;;
        --chopchopgo)
            check_hayanix
            print_info "Downloading ChopChopGo rules..."
            "$HAYANIX_BIN" rules download --source ChopChopGo
            ;;
        --sigmahq)
            check_hayanix
            print_info "Downloading SigmaHQ rules..."
            "$HAYANIX_BIN" rules download --source SigmaHQ
            ;;
        "")
            check_hayanix
            setup_rules
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
