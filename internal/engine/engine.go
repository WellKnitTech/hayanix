package engine

import (
	"fmt"
	"log"

	"github.com/wellknittech/hayanix/internal/parser"
	"github.com/wellknittech/hayanix/internal/rules"
	"github.com/wellknittech/hayanix/internal/output"
)

type Engine struct {
	target  string
	rules   string
	file    string
	output  string
	verbose bool
}

func New(target, rules, file, output string, verbose bool) *Engine {
	return &Engine{
		target:  target,
		rules:   rules,
		file:    file,
		output:  output,
		verbose: verbose,
	}
}

func (e *Engine) Run() error {
	if e.verbose {
		log.Printf("Starting analysis with target: %s, rules: %s, output: %s", e.target, e.rules, e.output)
	}

	// Load sigma rules from all directories
	ruleEngine, err := rules.NewEngine(e.rules)
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	// Determine log file path
	logFile := e.getLogFilePath()
	if e.verbose {
		log.Printf("Analyzing log file: %s", logFile)
	}

	// Parse log file
	logParser, err := parser.NewParser(e.target, logFile)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	// Process logs
	results, err := e.processLogs(logParser, ruleEngine)
	if err != nil {
		return fmt.Errorf("failed to process logs: %w", err)
	}

	// Output results
	outputter := output.NewOutputter(e.output)
	return outputter.Write(results)
}

func (e *Engine) getLogFilePath() string {
	if e.file != "" {
		return e.file
	}

	// Default log file paths based on target
	switch e.target {
	case "syslog":
		return "/var/log/messages"
	case "journald":
		return "/var/log/journal"
	case "auditd":
		return "/var/log/audit/audit.log"
	default:
		return "/var/log/messages"
	}
}

func (e *Engine) processLogs(logParser parser.Parser, ruleEngine *rules.Engine) ([]parser.LogEntry, error) {
	var results []parser.LogEntry
	
	entries, err := logParser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse log file: %w", err)
	}

	if e.verbose {
		log.Printf("Processing %d log entries", len(entries))
	}

	for _, entry := range entries {
		matches := ruleEngine.Evaluate(entry)
		if len(matches) > 0 {
			entry.MatchedRules = matches
			results = append(results, entry)
		}
	}

	if e.verbose {
		log.Printf("Found %d matching entries", len(results))
	}

	return results, nil
}
