package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wellknittech/hayanix/internal/parser"
)

func TestNewEngine(t *testing.T) {
	// Create a temporary rules directory with a test rule
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")
	err := os.MkdirAll(rulesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create rules directory: %v", err)
	}

	// Create a test rule file
	testRule := `title: Test Rule
id: test-rule-001
status: experimental
description: A test rule for unit testing
author: Test Author
date: 2025/01/01
modified: 2025/01/01
tags:
    - attack.test
level: low
logsource:
    category: process
    product: linux
    service: syslog
detection:
    selection:
        message:
            - 'test message'
    condition: selection
falsepositives:
    - Test false positive
fields:
    - message
    - hostname`

	ruleFile := filepath.Join(rulesDir, "test_rule.yml")
	err = os.WriteFile(ruleFile, []byte(testRule), 0644)
	if err != nil {
		t.Fatalf("Failed to create test rule file: %v", err)
	}

	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if engine == nil {
		t.Error("Expected engine to be created, got nil")
	}
}

func TestEngine_Match(t *testing.T) {
	// Create a temporary rules directory with a test rule
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")
	err := os.MkdirAll(rulesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create rules directory: %v", err)
	}

	// Create a test rule file
	testRule := `title: Test Rule
id: test-rule-001
status: experimental
description: A test rule for unit testing
author: Test Author
date: 2025/01/01
modified: 2025/01/01
tags:
    - attack.test
level: low
logsource:
    category: process
    product: linux
    service: syslog
detection:
    selection:
        message:
            - 'test message'
    condition: selection
falsepositives:
    - Test false positive
fields:
    - message
    - hostname`

	ruleFile := filepath.Join(rulesDir, "test_rule.yml")
	err = os.WriteFile(ruleFile, []byte(testRule), 0644)
	if err != nil {
		t.Fatalf("Failed to create test rule file: %v", err)
	}

	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	// Create test log entries
	entries := []parser.LogEntry{
		{
			Timestamp: "2025-01-01T10:30:15.000",
			Hostname:  "server1",
			Program:   "test",
			Message:   "This is a test message",
			Category:  "process",
			Product:   "linux",
			Service:   "syslog",
			Fields:    make(map[string]string),
		},
		{
			Timestamp: "2025-01-01T10:30:16.000",
			Hostname:  "server1",
			Program:   "test",
			Message:   "This is not a test message",
			Category:  "process",
			Product:   "linux",
			Service:   "syslog",
			Fields:    make(map[string]string),
		},
	}

	matchedEntries := engine.Match(entries)

	if len(matchedEntries) != 1 {
		t.Errorf("Expected 1 matched entry, got %d", len(matchedEntries))
	}

	if len(matchedEntries[0].MatchedRules) == 0 {
		t.Error("Expected matched entry to have rules, got none")
	}

	if matchedEntries[0].MatchedRules[0] != "test-rule-001" {
		t.Errorf("Expected matched rule 'test-rule-001', got '%s'", matchedEntries[0].MatchedRules[0])
	}
}

func TestEngine_EmptyRulesDirectory(t *testing.T) {
	// Create an empty rules directory
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")
	err := os.MkdirAll(rulesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create rules directory: %v", err)
	}

	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if engine == nil {
		t.Error("Expected engine to be created, got nil")
	}
}

func TestEngine_InvalidRuleFile(t *testing.T) {
	// Create a temporary rules directory with an invalid rule
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")
	err := os.MkdirAll(rulesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create rules directory: %v", err)
	}

	// Create an invalid rule file (malformed YAML)
	invalidRule := `title: Test Rule
id: test-rule-001
status: experimental
description: A test rule for unit testing
author: Test Author
date: 2025/01/01
modified: 2025/01/01
tags:
    - attack.test
level: low
logsource:
    category: process
    product: linux
    service: syslog
detection:
    selection:
        message:
            - 'test message'
    condition: selection
falsepositives:
    - Test false positive
fields:
    - message
    - hostname
invalid: [unclosed array`

	ruleFile := filepath.Join(rulesDir, "invalid_rule.yml")
	err = os.WriteFile(ruleFile, []byte(invalidRule), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid rule file: %v", err)
	}

	// This should not fail, but should log warnings
	engine, err := NewEngine(rulesDir)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	if engine == nil {
		t.Error("Expected engine to be created, got nil")
	}
}

func TestEngine_NonExistentRulesDirectory(t *testing.T) {
	// Try to create engine with non-existent directory
	_, err := NewEngine("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent rules directory, got nil")
	}
}
