package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/wellknittech/hayanix/internal/collection"
	"github.com/wellknittech/hayanix/internal/config"
	"github.com/wellknittech/hayanix/internal/engine"
	"github.com/wellknittech/hayanix/internal/rules"
	"github.com/wellknittech/hayanix/internal/wizard"
)

type Args struct {
	Version kong.VersionFlag `help:"Show version information."`

	// Analysis commands
	Analyze AnalyzeCmd `cmd:"" help:"Analyze logs with sigma rules."`

	// Collection analysis
	Collection CollectionCmd `cmd:"" help:"Analyze a collection of log files in a directory."`

	// Rule management commands
	Rules RulesCmd `cmd:"" help:"Manage sigma rules from external sources."`

	// Setup wizard
	Wizard WizardCmd `cmd:"" help:"Interactive setup wizard for configuring Hayanix."`

	Verbose bool `help:"Enable verbose output." short:"v"`
}

type AnalyzeCmd struct {
	Target    string `help:"Target log type (syslog, journald, auditd)." default:"syslog"`
	Rules     string `help:"Path to sigma rules directory." default:"./rules"`
	File      string `help:"Specific log file to analyze (optional)."`
	Output    string `help:"Output format (table, csv, json)." default:"table" enum:"table,csv,json"`
	UseConfig bool   `help:"Use saved configuration from wizard."`
}

type CollectionCmd struct {
	Path     string `help:"Path to directory containing log files."`
	RulesDir string `help:"Path to sigma rules directory." default:"./rules"`
	Format   string `help:"Output format (table, csv, json)." default:"table" enum:"table,csv,json"`
	Type     string `help:"Filter by log type (syslog, journald, auditd). Leave empty to analyze all types."`
	Detailed bool   `help:"Show detailed results for each file separately."`
	Summary  bool   `help:"Show collection summary only."`
	Verbose  bool   `help:"Enable verbose output." short:"v"`
}

type RulesCmd struct {
	List     RulesListCmd     `cmd:"" help:"List available rule sources."`
	Download RulesDownloadCmd `cmd:"" help:"Download rules from external sources."`
	Update   RulesUpdateCmd   `cmd:"" help:"Update existing rule sources."`
	Add      RulesAddCmd      `cmd:"" help:"Add a new rule source."`
	Remove   RulesRemoveCmd   `cmd:"" help:"Remove a rule source."`
	Enable   RulesEnableCmd   `cmd:"" help:"Enable a rule source."`
	Disable  RulesDisableCmd  `cmd:"" help:"Disable a rule source."`
}

type RulesListCmd struct {
	RulesDir string `help:"Path to rules directory." default:"./rules"`
}

type RulesDownloadCmd struct {
	Source   string `help:"Source name to download from."`
	RulesDir string `help:"Path to rules directory." default:"./rules"`
	All      bool   `help:"Download from all enabled sources."`
}

type RulesUpdateCmd struct {
	Source   string `help:"Source name to update."`
	RulesDir string `help:"Path to rules directory." default:"./rules"`
	All      bool   `help:"Update all enabled sources."`
}

type RulesAddCmd struct {
	Name        string `help:"Source name."`
	URL         string `help:"Repository URL."`
	Branch      string `help:"Branch name." default:"master"`
	Description string `help:"Source description."`
	RulesDir    string `help:"Path to rules directory." default:"./rules"`
}

type RulesRemoveCmd struct {
	Source   string `help:"Source name to remove."`
	RulesDir string `help:"Path to rules directory." default:"./rules"`
}

type RulesEnableCmd struct {
	Source   string `help:"Source name to enable."`
	RulesDir string `help:"Path to rules directory." default:"./rules"`
}

type RulesDisableCmd struct {
	Source   string `help:"Source name to disable."`
	RulesDir string `help:"Path to rules directory." default:"./rules"`
}

type WizardCmd struct {
	// No additional parameters needed for wizard
}

func (ac *AnalyzeCmd) Run() error {
	var target, rulesDir, file, output string

	// Use saved configuration if requested
	if ac.UseConfig {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load saved configuration: %w", err)
		}

		if cfg.LogFile == "" {
			return fmt.Errorf("no saved configuration found. Run 'hayanix wizard' first to create one")
		}

		target = cfg.LogType
		rulesDir = cfg.RulesDir
		file = cfg.LogFile
		output = cfg.OutputFormat

		fmt.Printf("Using saved configuration:\n")
		fmt.Printf("  Log Type: %s\n", target)
		fmt.Printf("  Log File: %s\n", file)
		fmt.Printf("  Rules Dir: %s\n", rulesDir)
		fmt.Printf("  Output: %s\n", output)
		fmt.Println()
	} else {
		target = ac.Target
		rulesDir = ac.Rules
		file = ac.File
		output = ac.Output
	}

	// Validate target
	validTargets := map[string]bool{
		"syslog":   true,
		"journald": true,
		"auditd":   true,
	}

	if !validTargets[target] {
		return fmt.Errorf("invalid target: %s. Valid targets are: syslog, journald, auditd", target)
	}

	// Validate rules directory
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return fmt.Errorf("rules directory does not exist: %s. Run 'hayanix wizard' to set up rules or create the directory manually", rulesDir)
	}

	// Check if rules directory is readable
	if _, err := os.Open(rulesDir); err != nil {
		return fmt.Errorf("cannot read rules directory %s: %w", rulesDir, err)
	}

	// Validate output format
	validOutputs := map[string]bool{
		"table": true,
		"csv":   true,
		"json":  true,
	}

	if !validOutputs[output] {
		return fmt.Errorf("invalid output format: %s. Valid formats are: table, csv, json", output)
	}

	// Create engine and run analysis
	eng := engine.New(target, rulesDir, file, output, false)
	return eng.Run()
}

func (rc *RulesListCmd) Run() error {
	rm := rules.NewRuleManager(rc.RulesDir)

	if err := rm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize rule manager: %w", err)
	}

	sources, err := rm.ListSources()
	if err != nil {
		return fmt.Errorf("failed to list sources: %w", err)
	}

	fmt.Println("Available rule sources:")
	fmt.Println("=====================")
	for _, source := range sources {
		status := "disabled"
		if source.Enabled {
			status = "enabled"
		}
		fmt.Printf("• %s (%s) - %s\n", source.Name, status, source.Description)
		fmt.Printf("  URL: %s\n", source.URL)
		fmt.Printf("  Branch: %s\n\n", source.Branch)
	}

	return nil
}

func (rc *RulesDownloadCmd) Run() error {
	rm := rules.NewRuleManager(rc.RulesDir)

	if err := rm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize rule manager: %w", err)
	}

	if rc.All {
		sources, err := rm.ListSources()
		if err != nil {
			return fmt.Errorf("failed to list sources: %w", err)
		}

		for _, source := range sources {
			if source.Enabled {
				fmt.Printf("Downloading rules from %s...\n", source.Name)
				if err := rm.DownloadRules(source.Name); err != nil {
					fmt.Printf("Warning: failed to download from %s: %v\n", source.Name, err)
				} else {
					fmt.Printf("Successfully downloaded rules from %s\n", source.Name)
				}
			}
		}
	} else if rc.Source != "" {
		fmt.Printf("Downloading rules from %s...\n", rc.Source)
		if err := rm.DownloadRules(rc.Source); err != nil {
			return fmt.Errorf("failed to download rules: %w", err)
		}
		fmt.Printf("Successfully downloaded rules from %s\n", rc.Source)
	} else {
		return fmt.Errorf("please specify a source name or use --all")
	}

	return nil
}

func (rc *RulesUpdateCmd) Run() error {
	// Update is the same as download for now
	downloadCmd := RulesDownloadCmd{
		Source:   rc.Source,
		RulesDir: rc.RulesDir,
		All:      rc.All,
	}
	return downloadCmd.Run()
}

func (rc *RulesAddCmd) Run() error {
	rm := rules.NewRuleManager(rc.RulesDir)

	if err := rm.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize rule manager: %w", err)
	}

	source := rules.RuleSource{
		Name:        rc.Name,
		URL:         rc.URL,
		Branch:      rc.Branch,
		Description: rc.Description,
		Enabled:     true,
	}

	if err := rm.AddSource(source); err != nil {
		return fmt.Errorf("failed to add source: %w", err)
	}

	fmt.Printf("Successfully added source: %s\n", rc.Name)
	return nil
}

func (rc *RulesRemoveCmd) Run() error {
	rm := rules.NewRuleManager(rc.RulesDir)

	if err := rm.RemoveSource(rc.Source); err != nil {
		return fmt.Errorf("failed to remove source: %w", err)
	}

	fmt.Printf("Successfully removed source: %s\n", rc.Source)
	return nil
}

func (rc *RulesEnableCmd) Run() error {
	rm := rules.NewRuleManager(rc.RulesDir)

	if err := rm.EnableSource(rc.Source); err != nil {
		return fmt.Errorf("failed to enable source: %w", err)
	}

	fmt.Printf("Successfully enabled source: %s\n", rc.Source)
	return nil
}

func (rc *RulesDisableCmd) Run() error {
	rm := rules.NewRuleManager(rc.RulesDir)

	if err := rm.DisableSource(rc.Source); err != nil {
		return fmt.Errorf("failed to disable source: %w", err)
	}

	fmt.Printf("Successfully disabled source: %s\n", rc.Source)
	return nil
}

func (wc *WizardCmd) Run() error {
	w := wizard.NewWizard()

	config, err := w.Run()
	if err != nil {
		return err
	}

	return w.ExecuteConfiguration(config)
}

func (cc *CollectionCmd) Run() error {
	// Validate path
	if cc.Path == "" {
		return fmt.Errorf("collection path is required. Use --path to specify a directory containing log files")
	}

	// Check if path exists
	if _, err := os.Stat(cc.Path); os.IsNotExist(err) {
		return fmt.Errorf("collection path does not exist: %s", cc.Path)
	}

	// Check if path is a directory
	fileInfo, err := os.Stat(cc.Path)
	if err != nil {
		return fmt.Errorf("cannot access collection path %s: %w", cc.Path, err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("collection path must be a directory: %s", cc.Path)
	}

	// Validate output format
	validOutputs := map[string]bool{
		"table": true,
		"csv":   true,
		"json":  true,
	}

	if !validOutputs[cc.Format] {
		return fmt.Errorf("invalid output format: %s. Valid formats are: table, csv, json", cc.Format)
	}

	// Validate log type if specified
	if cc.Type != "" {
		validTypes := map[string]bool{
			"syslog":   true,
			"journald": true,
			"auditd":   true,
		}

		if !validTypes[cc.Type] {
			return fmt.Errorf("invalid log type: %s. Valid types are: syslog, journald, auditd", cc.Type)
		}
	}

	// Discover log files
	collector := collection.NewCollector(cc.Path)
	logCollection, err := collector.DiscoverLogFiles()
	if err != nil {
		return fmt.Errorf("failed to discover log files: %w", err)
	}

	// Filter by type if specified
	if cc.Type != "" {
		logCollection = collector.FilterByType(logCollection, cc.Type)
	}

	// Validate collection
	if err := collector.ValidateCollection(logCollection); err != nil {
		return fmt.Errorf("collection validation failed: %w", err)
	}

	// Show collection summary
	fmt.Println("📁 Log Collection Discovered")
	fmt.Println("============================")
	fmt.Printf("Base Path: %s\n", logCollection.BasePath)
	fmt.Printf("Total Files: %d\n", logCollection.Summary.TotalFiles)
	fmt.Printf("Total Size: %.2f MB\n", float64(logCollection.Summary.TotalSize)/(1024*1024))
	fmt.Println()

	fmt.Println("Files by Type:")
	for logType, count := range logCollection.Summary.FilesByType {
		size := logCollection.Summary.SizeByType[logType]
		fmt.Printf("  %s: %d files (%.2f MB)\n", logType, count, float64(size)/(1024*1024))
	}
	fmt.Println()

	// If summary only, return here
	if cc.Summary {
		return nil
	}

	// Create analyzer
	analyzer, err := collection.NewCollectionAnalyzer(logCollection, cc.RulesDir, cc.Format, cc.Verbose)
	if err != nil {
		return fmt.Errorf("failed to create analyzer: %w", err)
	}

	// Analyze collection
	result, err := analyzer.AnalyzeCollection()
	if err != nil {
		return fmt.Errorf("failed to analyze collection: %w", err)
	}

	// Write results
	if cc.Detailed {
		if err := analyzer.WriteDetailedResults(result); err != nil {
			return fmt.Errorf("failed to write detailed results: %w", err)
		}
	} else {
		if err := analyzer.WriteResults(result); err != nil {
			return fmt.Errorf("failed to write results: %w", err)
		}
	}

	// Show summary
	analyzer.WriteSummary(result)

	return nil
}
