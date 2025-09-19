#!/bin/bash

# Release script for Hayanix
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Version from main.go
VERSION=$(grep 'var version =' main.go | cut -d'"' -f2)
echo -e "${GREEN}Building Hayanix v${VERSION}${NC}"

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
make clean

# Build for all platforms
echo -e "${YELLOW}Building for all platforms...${NC}"
make build-all

# Create release directory
RELEASE_DIR="releases/v${VERSION}"
mkdir -p "$RELEASE_DIR"

# Copy binaries to release directory
echo -e "${YELLOW}Organizing release files...${NC}"
cp hayanix-linux-amd64 "$RELEASE_DIR/"
cp hayanix-linux-arm64 "$RELEASE_DIR/"
cp hayanix-darwin-amd64 "$RELEASE_DIR/"
cp hayanix-darwin-arm64 "$RELEASE_DIR/"
cp hayanix-windows-amd64.exe "$RELEASE_DIR/"

# Create checksums
echo -e "${YELLOW}Generating checksums...${NC}"
cd "$RELEASE_DIR"
sha256sum * > checksums.txt
cd ../..

# Copy documentation
echo -e "${YELLOW}Copying documentation...${NC}"
cp README.md "$RELEASE_DIR/"
cp LICENSE "$RELEASE_DIR/"
cp CHANGELOG.md "$RELEASE_DIR/"

# Create release notes
echo -e "${YELLOW}Creating release notes...${NC}"
cat > "$RELEASE_DIR/RELEASE_NOTES.md" << EOF
# Hayanix v${VERSION} Release Notes

## What's New

This is the first release of Hayanix, a sigma-based threat hunting and fast forensics timeline generator for *nix logs.

### Features

- 🎯 **Sigma Rule Engine**: Hunt for threats using Sigma detection rules
- ⚡ **Lightning Fast**: Written in Go for optimal performance
- 🪶 **Clean Output**: Multiple output formats (table, CSV, JSON)
- 💻 **Multi-Platform**: Runs on Linux, macOS, and Windows
- 📊 **Multiple Log Sources**: Supports syslog, journald, and auditd
- 🔍 **Comprehensive Rules**: Pre-built rules for common *nix threats
- 📁 **Collection Analysis**: Automatically discover and analyze all log files in a directory
- 🧙‍♂️ **Interactive Wizard**: Guided setup for easy configuration
- 🔧 **Rule Management**: Download and manage rules from external sources

### Installation

1. Download the appropriate binary for your platform
2. Make it executable: \`chmod +x hayanix\`
3. Run the setup wizard: \`./hayanix wizard\`
4. Start analyzing: \`./hayanix analyze\`

### Supported Platforms

- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

### Quick Start

\`\`\`bash
# Run the interactive setup wizard
./hayanix wizard

# Analyze logs with saved configuration
./hayanix analyze --use-config

# Analyze a collection of log files
./hayanix collection --path /var/log
\`\`\`

### Documentation

See README.md for complete documentation and usage examples.

### Support

- GitHub Issues: https://github.com/wellknittech/hayanix/issues
- Documentation: https://github.com/wellknittech/hayanix#readme

EOF

# Create archive
echo -e "${YELLOW}Creating release archive...${NC}"
cd releases
tar -czf "hayanix-v${VERSION}.tar.gz" "v${VERSION}"
zip -r "hayanix-v${VERSION}.zip" "v${VERSION}"
cd ..

echo -e "${GREEN}Release v${VERSION} created successfully!${NC}"
echo -e "${GREEN}Release files are in: ${RELEASE_DIR}${NC}"
echo -e "${GREEN}Archives created:${NC}"
echo -e "  - releases/hayanix-v${VERSION}.tar.gz"
echo -e "  - releases/hayanix-v${VERSION}.zip"

echo -e "${YELLOW}Next steps:${NC}"
echo -e "1. Test the binaries on different platforms"
echo -e "2. Create a GitHub release with these files"
echo -e "3. Update the installation instructions in README.md"
