# Hayanix

[![Go Report Card](https://goreportcard.com/badge/github.com/wellknittech/hayanix)](https://goreportcard.com/report/github.com/wellknittech/hayanix)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Hayanix** is a sigma-based threat hunting and fast forensics timeline generator for *nix logs, inspired by [Hayabusa](https://github.com/Yamato-Security/hayabusa) and [ChopChopGo](https://github.com/M00NLIG7/ChopChopGo).

## Features

- 🎯 **Sigma Rule Engine**: Hunt for threats using Sigma detection rules
- ⚡ **Lightning Fast**: Written in Go for optimal performance
- 🪶 **Clean Output**: Multiple output formats (table, CSV, JSON)
- 💻 **Multi-Platform**: Runs on Linux, macOS, and Windows
- 📊 **Multiple Log Sources**: Supports syslog, journald, and auditd
- 🔍 **Comprehensive Rules**: Pre-built rules for common *nix threats
- 📁 **Collection Analysis**: Automatically discover and analyze all log files in a directory
- 🧙‍♂️ **Interactive Wizard**: Guided setup for easy configuration
- 🔧 **Rule Management**: Download and manage rules from external sources

## Quick Start

### Installation

#### From Source
```bash
git clone https://github.com/wellknittech/hayanix.git
cd hayanix
make build
make setup-rules  # Optional: download external rule sources
```

#### Pre-built Binaries
Download the latest release from the [releases page](https://github.com/wellknittech/hayanix/releases).

```bash
# Download and extract the appropriate binary for your platform
wget https://github.com/wellknittech/hayanix/releases/download/v0.1.0/hayanix-v0.1.0.tar.gz
tar -xzf hayanix-v0.1.0.tar.gz
cd hayanix-v0.1.0

# Make executable and run
chmod +x hayanix-linux-amd64  # or hayanix-darwin-amd64, hayanix-windows-amd64.exe
./hayanix-linux-amd64 wizard
```

### Basic Usage

#### Quick Start with Wizard
```bash
# Run the interactive setup wizard
./hayanix wizard

# Use saved configuration
./hayanix analyze --use-config
```

#### Manual Configuration
```bash
# Analyze syslog with default rules
./hayanix analyze

# Analyze specific log file
./hayanix analyze --target syslog --file /var/log/messages

# Use custom rules directory
./hayanix analyze --target syslog --rules ./custom-rules

# Output in CSV format
./hayanix analyze --target syslog --output csv

# Analyze journald logs
./hayanix analyze --target journald --rules ./rules/linux/journald/

# Analyze auditd logs
./hayanix analyze --target auditd --rules ./rules/linux/auditd/ --file /var/log/audit/audit.log
```

#### Collection Analysis
```bash
# Analyze all log files in a directory
./hayanix collection --path /var/log

# Analyze specific log types only
./hayanix collection --path /var/log --type syslog

# Show detailed results for each file
./hayanix collection --path /var/log --detailed

# Export results to CSV
./hayanix collection --path /var/log --format csv

# Show collection summary only
./hayanix collection --path /var/log --summary
```

### Interactive Setup Wizard

The easiest way to get started with Hayanix is using the interactive setup wizard:

```bash
# Run the setup wizard
./hayanix wizard
```

The wizard will guide you through:
1. **Log Type Selection** - Choose between syslog, journald, or auditd
2. **Log File Configuration** - Specify the log file to analyze
3. **Rules Directory** - Choose where to store sigma rules
4. **Output Format** - Select table, CSV, or JSON output
5. **Rule Sources** - Download rules from ChopChopGo, SigmaHQ, or custom sources
6. **Configuration Saving** - Save settings for future use

After running the wizard, you can use your saved configuration:
```bash
# Use saved configuration
./hayanix analyze --use-config
```

### Rule Management

Hayanix includes a comprehensive rule management system that allows you to import and use rules from external sources like ChopChopGo and SigmaHQ.

```bash
# List available rule sources
./hayanix rules list

# Download rules from ChopChopGo
./hayanix rules download --source ChopChopGo

# Download rules from SigmaHQ
./hayanix rules download --source SigmaHQ

# Download from all enabled sources
./hayanix rules download --all

# Update existing rules
./hayanix rules update --source ChopChopGo

# Add a custom rule source
./hayanix rules add --name "MyRules" --url "https://github.com/user/rules" --description "Custom rules"

# Enable/disable sources
./hayanix rules enable --source ChopChopGo
./hayanix rules disable --source SigmaHQ
```

## Command Line Options

### Analyze Command
| Option | Description | Default |
|--------|-------------|---------|
| `--target` | Target log type (syslog, journald, auditd) | syslog |
| `--rules` | Path to sigma rules directory | ./rules |
| `--file` | Specific log file to analyze | Auto-detected |
| `--output` | Output format (table, csv, json) | table |
| `--use-config` | Use saved configuration from wizard | false |

### Collection Command
| Option | Description | Default |
|--------|-------------|---------|
| `--path` | Path to directory containing log files | Required |
| `--rules-dir` | Path to sigma rules directory | ./rules |
| `--format` | Output format (table, csv, json) | table |
| `--type` | Filter by log type (syslog, journald, auditd) | All types |
| `--detailed` | Show detailed results for each file separately | false |
| `--summary` | Show collection summary only | false |

### Wizard Command
| Command | Description |
|---------|-------------|
| `wizard` | Interactive setup wizard for configuring Hayanix |

### Rules Management Commands
| Command | Description |
|---------|-------------|
| `rules list` | List available rule sources |
| `rules download` | Download rules from external sources |
| `rules update` | Update existing rule sources |
| `rules add` | Add a new rule source |
| `rules remove` | Remove a rule source |
| `rules enable` | Enable a rule source |
| `rules disable` | Disable a rule source |

### Global Options
| Option | Description | Default |
|--------|-------------|---------|
| `--verbose` | Enable verbose output | false |
| `--version` | Show version information | - |

## Supported Log Formats

### Syslog
- **Default Path**: `/var/log/messages`
- **Format**: `Jan 2 15:04:05 hostname program[pid]: message`
- **Use Case**: General system logs, authentication events

### Journald
- **Default Path**: `/var/log/journal`
- **Format**: `2025-01-01T15:04:05Z hostname program[pid]: message`
- **Use Case**: Modern Linux systems with systemd

### Auditd
- **Default Path**: `/var/log/audit/audit.log`
- **Format**: `type=... msg=audit(timestamp:pid): ...`
- **Use Case**: Detailed system call auditing

## Sigma Rules

Hayanix uses Sigma rules for threat detection. Rules are organized by log source and can be loaded from multiple sources:

```
rules/
├── linux/                    # Built-in rules
│   ├── syslog/
│   │   ├── suspicious_login_attempts.yml
│   │   ├── privilege_escalation.yml
│   │   └── network_scanning.yml
│   ├── journald/
│   │   └── systemd_service_manipulation.yml
│   └── auditd/
│       └── file_access.yml
├── external/                  # External rule sources
│   ├── chopchopgo/           # ChopChopGo rules
│   ├── sigmahq/              # Official SigmaHQ rules
│   └── custom/               # Custom rule sources
└── sources.yml               # Rule source configuration
```

### Pre-configured Rule Sources

Hayanix comes with two pre-configured rule sources:

1. **ChopChopGo** - Linux forensics rules from the ChopChopGo project
   - URL: https://github.com/M00NLIG7/ChopChopGo
   - Focus: Linux-specific threat detection
   - Rules: syslog, journald, auditd

2. **SigmaHQ** - Official Sigma rules repository
   - URL: https://github.com/SigmaHQ/sigma
   - Focus: Comprehensive threat detection across platforms
   - Rules: Linux auditd, systemd, rsyslog, and more

### Adding Custom Rule Sources

You can add your own rule sources or community repositories:

```bash
# Add a custom rule source
./hayanix rules add --name "MyOrgRules" \
  --url "https://github.com/myorg/sigma-rules" \
  --branch "main" \
  --description "My organization's custom sigma rules"

# Download rules from the new source
./hayanix rules download --source MyOrgRules
```

### Creating Custom Rules

Sigma rules follow the standard format. Here's an example:

```yaml
title: Suspicious Login Attempts
id: hayanix-linux-syslog-suspicious-login-attempts
status: experimental
description: Detects suspicious login attempts and authentication failures
author: Hayanix Team
date: 2025/01/01
modified: 2025/01/01
tags:
    - attack.credential_access
    - attack.t1110
level: medium
logsource:
    category: process
    product: linux
    service: syslog
detection:
    selection:
        message:
            - 'Failed password for'
            - 'Invalid user'
            - 'authentication failure'
    condition: selection
falsepositives:
    - Legitimate users forgetting passwords
fields:
    - message
    - hostname
    - program
```

## Output Formats

### Table Format (Default)
```
Found 3 matching entries:

+-----------------+----------+---------+----------------------------------+------------------+
|    TIMESTAMP    | HOSTNAME | PROGRAM |             MESSAGE               |       TAGS       |
+-----------------+----------+---------+----------------------------------+------------------+
| Jan  1 10:30:15 | server1  | sshd    | Failed password for root from... | suspicious_login |
| Jan  1 10:31:22 | server1  | sudo    | user : TTY=pts/0 ; PWD=/home...  | privilege_escal  |
+-----------------+----------+---------+----------------------------------+------------------+
```

### CSV Format
```csv
timestamp,hostname,program,pid,message,matched_rules
Jan  1 10:30:15,server1,sshd,1234,Failed password for root,suspicious_login
Jan  1 10:31:22,server1,sudo,5678,user : TTY=pts/0,privilege_escal
```

### JSON Format
```json
[
  {
    "timestamp": "Jan  1 10:30:15",
    "hostname": "server1",
    "program": "sshd",
    "pid": "1234",
    "message": "Failed password for root",
    "matched_rules": ["suspicious_login"]
  }
]
```

## Building from Source

### Prerequisites
- Go 1.21 or later
- Make (optional, for using Makefile)

### Build Commands
```bash
# Install dependencies
make deps

# Build binary
make build

# Build for all platforms
make build-all

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean
```

## Development

### Project Structure
```
hayanix/
├── internal/
│   ├── cli/          # Command-line interface
│   ├── engine/       # Core analysis engine
│   ├── parser/       # Log file parsers
│   ├── rules/        # Sigma rule engine
│   └── output/       # Output formatters
├── rules/            # Sigma rules
├── main.go          # Application entry point
├── go.mod           # Go module definition
└── Makefile         # Build system
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Hayabusa](https://github.com/Yamato-Security/hayabusa) - Windows event log analysis tool
- [ChopChopGo](https://github.com/M00NLIG7/ChopChopGo) - Linux forensics artifact recovery
- [Sigma](https://github.com/SigmaHQ/sigma) - Generic signature format for SIEM systems

## Troubleshooting

### Common Issues

#### "rules directory does not exist"
```bash
# Solution: Run the wizard to set up rules
./hayanix wizard

# Or create the directory manually
mkdir -p ./rules
```

#### "No matching entries found"
- Check that your log files contain the expected format
- Verify that rules are loaded correctly: `./hayanix rules list`
- Try with verbose output: `./hayanix analyze --verbose`

#### "Failed to parse YAML" warnings
- These warnings indicate some external rules have formatting issues
- They don't affect functionality - valid rules will still be loaded
- You can ignore these warnings or update the problematic rule files

#### "Collection path must be a directory"
```bash
# Make sure you're pointing to a directory, not a file
./hayanix collection --path /var/log  # ✅ Correct
./hayanix collection --path /var/log/messages  # ❌ Incorrect
```

#### Permission denied errors
```bash
# Make sure you have read access to log files
sudo ./hayanix analyze --target syslog --file /var/log/messages

# Or run as root for system logs
sudo ./hayanix collection --path /var/log
```

### Performance Tips

- For large log files, use the collection feature to process multiple files
- Use CSV output for large datasets: `--format csv`
- Filter by log type to reduce processing time: `--type syslog`

### Getting Help

1. Check the [Issues](https://github.com/wellknittech/hayanix/issues) page
2. Run with verbose output to see detailed information
3. Ensure you're using the latest version
4. Check that your log files are in the expected format

## Roadmap

- [ ] Support for more log formats (Apache, Nginx, etc.)
- [ ] Real-time log monitoring
- [ ] Web-based dashboard
- [ ] Integration with SIEM systems
- [ ] Machine learning-based anomaly detection
- [ ] Performance optimizations for large log files
