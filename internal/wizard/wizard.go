package wizard

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wellknittech/hayanix/internal/config"
	"github.com/wellknittech/hayanix/internal/rules"
)

type Wizard struct {
	reader *bufio.Reader
}

type WizardConfig struct {
	LogFile      string
	LogType      string
	RulesDir     string
	OutputFormat string
	DownloadRules bool
	RuleSources  []string
	SaveConfig   bool
}

func NewWizard() *Wizard {
	return &Wizard{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (w *Wizard) Run() (*WizardConfig, error) {
	fmt.Println("🔍 Welcome to Hayanix Setup Wizard!")
	fmt.Println("=====================================")
	fmt.Println("This wizard will help you configure Hayanix for log analysis.")
	fmt.Println()

	// Load existing config if available
	existingConfig, err := config.LoadConfig()
	if err == nil && existingConfig.LogFile != "" {
		fmt.Println("📋 Found existing configuration:")
		fmt.Printf("   Log Type: %s\n", existingConfig.LogType)
		fmt.Printf("   Log File: %s\n", existingConfig.LogFile)
		fmt.Printf("   Rules Dir: %s\n", existingConfig.RulesDir)
		fmt.Printf("   Output: %s\n", existingConfig.OutputFormat)
		fmt.Println()
		
		fmt.Print("Do you want to use existing configuration? (Y/n): ")
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		
		if input == "" || input == "y" || input == "yes" {
			return &WizardConfig{
				LogFile:      existingConfig.LogFile,
				LogType:      existingConfig.LogType,
				RulesDir:     existingConfig.RulesDir,
				OutputFormat: existingConfig.OutputFormat,
				RuleSources:  existingConfig.RuleSources,
				SaveConfig:   false, // Already saved
			}, nil
		}
	}

	wizardConfig := &WizardConfig{}

	// Step 1: Choose log type
	if err := w.selectLogType(wizardConfig); err != nil {
		return nil, err
	}

	// Step 2: Choose log file
	if err := w.selectLogFile(wizardConfig); err != nil {
		return nil, err
	}

	// Step 3: Choose rules directory
	if err := w.selectRulesDirectory(wizardConfig); err != nil {
		return nil, err
	}

	// Step 4: Choose output format
	if err := w.selectOutputFormat(wizardConfig); err != nil {
		return nil, err
	}

	// Step 5: Setup rule sources
	if err := w.setupRuleSources(wizardConfig); err != nil {
		return nil, err
	}

	// Step 6: Ask about saving config
	if err := w.askSaveConfig(wizardConfig); err != nil {
		return nil, err
	}

	// Step 7: Summary and confirmation
	if err := w.showSummary(wizardConfig); err != nil {
		return nil, err
	}

	return wizardConfig, nil
}

func (w *Wizard) selectLogType(config *WizardConfig) error {
	fmt.Println("📋 Step 1: Select Log Type")
	fmt.Println("-------------------------")
	fmt.Println("Choose the type of logs you want to analyze:")
	fmt.Println("1. Syslog (default: /var/log/messages)")
	fmt.Println("2. Journald (systemd logs)")
	fmt.Println("3. Auditd (audit logs)")
	fmt.Println()

	for {
		fmt.Print("Enter your choice (1-3) [1]: ")
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			input = "1"
		}

		switch input {
		case "1":
			config.LogType = "syslog"
			fmt.Println("✅ Selected: Syslog")
			break
		case "2":
			config.LogType = "journald"
			fmt.Println("✅ Selected: Journald")
			break
		case "3":
			config.LogType = "auditd"
			fmt.Println("✅ Selected: Auditd")
			break
		default:
			fmt.Println("❌ Invalid choice. Please enter 1, 2, or 3.")
			continue
		}
		break
	}

	fmt.Println()
	return nil
}

func (w *Wizard) selectLogFile(config *WizardConfig) error {
	fmt.Println("📁 Step 2: Select Log File")
	fmt.Println("-------------------------")

	// Default paths based on log type
	defaultPaths := map[string]string{
		"syslog":   "/var/log/messages",
		"journald": "/var/log/journal",
		"auditd":   "/var/log/audit/audit.log",
	}

	defaultPath := defaultPaths[config.LogType]
	fmt.Printf("Default path for %s: %s\n", config.LogType, defaultPath)
	fmt.Println()

	for {
		fmt.Printf("Enter log file path [%s]: ", defaultPath)
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			input = defaultPath
		}

		// Check if file exists
		if _, err := os.Stat(input); os.IsNotExist(err) {
			fmt.Printf("⚠️  File does not exist: %s\n", input)
			fmt.Print("Do you want to continue anyway? (y/N): ")
			confirm, _ := w.reader.ReadString('\n')
			confirm = strings.TrimSpace(strings.ToLower(confirm))
			if confirm != "y" && confirm != "yes" {
				continue
			}
		}

		config.LogFile = input
		fmt.Printf("✅ Selected: %s\n", input)
		break
	}

	fmt.Println()
	return nil
}

func (w *Wizard) selectRulesDirectory(config *WizardConfig) error {
	fmt.Println("📚 Step 3: Select Rules Directory")
	fmt.Println("--------------------------------")
	fmt.Println("Choose where to store and load sigma rules:")
	fmt.Println("1. Default (./rules)")
	fmt.Println("2. Custom path")
	fmt.Println()

	for {
		fmt.Print("Enter your choice (1-2) [1]: ")
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			input = "1"
		}

		switch input {
		case "1":
			config.RulesDir = "./rules"
			fmt.Println("✅ Selected: ./rules")
			break
		case "2":
			fmt.Print("Enter custom rules directory path: ")
			customPath, _ := w.reader.ReadString('\n')
			customPath = strings.TrimSpace(customPath)
			if customPath == "" {
				fmt.Println("❌ Path cannot be empty.")
				continue
			}
			config.RulesDir = customPath
			fmt.Printf("✅ Selected: %s\n", customPath)
			break
		default:
			fmt.Println("❌ Invalid choice. Please enter 1 or 2.")
			continue
		}
		break
	}

	fmt.Println()
	return nil
}

func (w *Wizard) selectOutputFormat(config *WizardConfig) error {
	fmt.Println("📊 Step 4: Select Output Format")
	fmt.Println("------------------------------")
	fmt.Println("Choose the output format for analysis results:")
	fmt.Println("1. Table (human-readable)")
	fmt.Println("2. CSV (spreadsheet-compatible)")
	fmt.Println("3. JSON (machine-readable)")
	fmt.Println()

	for {
		fmt.Print("Enter your choice (1-3) [1]: ")
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			input = "1"
		}

		switch input {
		case "1":
			config.OutputFormat = "table"
			fmt.Println("✅ Selected: Table")
			break
		case "2":
			config.OutputFormat = "csv"
			fmt.Println("✅ Selected: CSV")
			break
		case "3":
			config.OutputFormat = "json"
			fmt.Println("✅ Selected: JSON")
			break
		default:
			fmt.Println("❌ Invalid choice. Please enter 1, 2, or 3.")
			continue
		}
		break
	}

	fmt.Println()
	return nil
}

func (w *Wizard) setupRuleSources(config *WizardConfig) error {
	fmt.Println("🔧 Step 5: Setup Rule Sources")
	fmt.Println("----------------------------")
	fmt.Println("Hayanix can download rules from external sources.")
	fmt.Println("Available sources:")
	fmt.Println("1. ChopChopGo (Linux forensics rules)")
	fmt.Println("2. SigmaHQ (Official Sigma rules)")
	fmt.Println("3. Skip rule setup")
	fmt.Println()

	for {
		fmt.Print("Do you want to download external rules? (y/N): ")
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" || input == "n" || input == "no" {
			config.DownloadRules = false
			fmt.Println("✅ Skipping rule download")
			break
		} else if input == "y" || input == "yes" {
			config.DownloadRules = true
			fmt.Println("✅ Will download external rules")
			break
		} else {
			fmt.Println("❌ Please enter 'y' or 'n'.")
			continue
		}
	}

	if config.DownloadRules {
		fmt.Println()
		fmt.Println("Select rule sources to download:")
		fmt.Println("1. ChopChopGo only")
		fmt.Println("2. SigmaHQ only")
		fmt.Println("3. Both ChopChopGo and SigmaHQ")
		fmt.Println()

		for {
			fmt.Print("Enter your choice (1-3) [3]: ")
			input, _ := w.reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "" {
				input = "3"
			}

			switch input {
			case "1":
				config.RuleSources = []string{"ChopChopGo"}
				fmt.Println("✅ Selected: ChopChopGo")
				break
			case "2":
				config.RuleSources = []string{"SigmaHQ"}
				fmt.Println("✅ Selected: SigmaHQ")
				break
			case "3":
				config.RuleSources = []string{"ChopChopGo", "SigmaHQ"}
				fmt.Println("✅ Selected: Both ChopChopGo and SigmaHQ")
				break
			default:
				fmt.Println("❌ Invalid choice. Please enter 1, 2, or 3.")
				continue
			}
			break
		}
	}

	fmt.Println()
	return nil
}

func (w *Wizard) askSaveConfig(config *WizardConfig) error {
	fmt.Println("💾 Step 6: Save Configuration")
	fmt.Println("---------------------------")
	fmt.Println("Do you want to save this configuration for future use?")
	fmt.Println("This will allow you to quickly reuse these settings.")
	fmt.Println()

	for {
		fmt.Print("Save configuration? (Y/n): ")
		input, _ := w.reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" || input == "y" || input == "yes" {
			config.SaveConfig = true
			fmt.Println("✅ Configuration will be saved")
			break
		} else if input == "n" || input == "no" {
			config.SaveConfig = false
			fmt.Println("✅ Configuration will not be saved")
			break
		} else {
			fmt.Println("❌ Please enter 'y' or 'n'.")
			continue
		}
	}

	fmt.Println()
	return nil
}

func (w *Wizard) showSummary(config *WizardConfig) error {
	fmt.Println("📋 Configuration Summary")
	fmt.Println("========================")
	fmt.Printf("Log Type: %s\n", config.LogType)
	fmt.Printf("Log File: %s\n", config.LogFile)
	fmt.Printf("Rules Directory: %s\n", config.RulesDir)
	fmt.Printf("Output Format: %s\n", config.OutputFormat)
	fmt.Printf("Download Rules: %t\n", config.DownloadRules)
	if config.DownloadRules {
		fmt.Printf("Rule Sources: %s\n", strings.Join(config.RuleSources, ", "))
	}
	fmt.Println()

	fmt.Print("Do you want to proceed with this configuration? (Y/n): ")
	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "n" || input == "no" {
		fmt.Println("❌ Configuration cancelled.")
		return fmt.Errorf("user cancelled configuration")
	}

	fmt.Println("✅ Configuration confirmed!")
	fmt.Println()
	return nil
}

func (w *Wizard) ExecuteConfiguration(wizardConfig *WizardConfig) error {
	fmt.Println("🚀 Executing Configuration")
	fmt.Println("==========================")

	// Save configuration if requested
	if wizardConfig.SaveConfig {
		fmt.Println("💾 Saving configuration...")
		cfg := &config.Config{
			LogFile:      wizardConfig.LogFile,
			LogType:      wizardConfig.LogType,
			RulesDir:     wizardConfig.RulesDir,
			OutputFormat: wizardConfig.OutputFormat,
			RuleSources:  wizardConfig.RuleSources,
			LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
		}
		
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("⚠️  Warning: failed to save configuration: %v\n", err)
		} else {
			fmt.Println("✅ Configuration saved successfully")
		}
		fmt.Println()
	}

	// Initialize rule manager
	if wizardConfig.DownloadRules {
		fmt.Println("📥 Setting up rule sources...")
		rm := rules.NewRuleManager(wizardConfig.RulesDir)
		
		if err := rm.Initialize(); err != nil {
			fmt.Printf("❌ Failed to initialize rule manager: %v\n", err)
			return err
		}

		// Download selected rule sources
		for _, source := range wizardConfig.RuleSources {
			fmt.Printf("📥 Downloading rules from %s...\n", source)
			if err := rm.DownloadRules(source); err != nil {
				fmt.Printf("⚠️  Warning: failed to download %s rules: %v\n", source, err)
			} else {
				fmt.Printf("✅ Successfully downloaded %s rules\n", source)
			}
		}
		fmt.Println()
	}

	fmt.Println("🎉 Setup Complete!")
	fmt.Println("==================")
	fmt.Println("You can now run Hayanix with the following command:")
	fmt.Println()
	fmt.Printf("./hayanix analyze --target %s --file %s --rules %s --output %s\n",
		wizardConfig.LogType, wizardConfig.LogFile, wizardConfig.RulesDir, wizardConfig.OutputFormat)
	fmt.Println()
	fmt.Println("Or run the wizard again anytime with: ./hayanix wizard")

	return nil
}
