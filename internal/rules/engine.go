package rules

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/wellknittech/hayanix/internal/parser"
)

type Engine struct {
	rules []Rule
}

type Rule struct {
	Title       string                 `yaml:"title"`
	ID          string                 `yaml:"id"`
	Status      string                 `yaml:"status"`
	Description string                 `yaml:"description"`
	Author      string                 `yaml:"author"`
	Date        string                 `yaml:"date"`
	Modified    string                 `yaml:"modified"`
	Tags        []string               `yaml:"tags"`
	Level       string                 `yaml:"level"`
	Logsource   LogSource              `yaml:"logsource"`
	Detection   map[string]interface{} `yaml:"detection"`
	Falsepositives []string            `yaml:"falsepositives"`
	Fields      []string               `yaml:"fields"`
}

type LogSource struct {
	Category string `yaml:"category"`
	Product  string `yaml:"product"`
	Service  string `yaml:"service"`
}

func NewEngine(rulesDir string) (*Engine, error) {
	engine := &Engine{
		rules: make([]Rule, 0),
	}

	if err := engine.loadRules(rulesDir); err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	return engine, nil
}

func (e *Engine) loadRules(rulesDir string) error {
	// Load rules from multiple directories
	dirs := []string{
		rulesDir,
		filepath.Join(rulesDir, "linux"),
		filepath.Join(rulesDir, "linux", "syslog"),
		filepath.Join(rulesDir, "linux", "journald"),
		filepath.Join(rulesDir, "linux", "auditd"),
		filepath.Join(rulesDir, "external"),
		filepath.Join(rulesDir, "external", "chopchopgo"),
		filepath.Join(rulesDir, "external", "sigmahq"),
		filepath.Join(rulesDir, "external", "custom"),
	}

	for _, dir := range dirs {
		if err := e.loadRulesFromDir(dir); err != nil {
			log.Printf("Warning: failed to load rules from %s: %v", dir, err)
		}
	}

	return nil
}

func (e *Engine) loadRulesFromDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".yml") && !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		rule, err := e.loadRule(path)
		if err != nil {
			log.Printf("Warning: failed to load rule %s: %v", path, err)
			return nil // Continue loading other rules
		}

		// Additional validation for rule structure
		if err := e.validateRule(rule); err != nil {
			log.Printf("Warning: rule %s failed validation: %v", path, err)
			return nil // Continue loading other rules
		}

		e.rules = append(e.rules, rule)
		return nil
	})
}

func (e *Engine) loadRule(filePath string) (Rule, error) {
	var rule Rule

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return rule, fmt.Errorf("failed to read rule file %s: %w", filePath, err)
	}

	if err := yaml.Unmarshal(data, &rule); err != nil {
		return rule, fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
	}

	// Validate required fields
	if rule.ID == "" {
		return rule, fmt.Errorf("rule in %s is missing required field 'id'", filePath)
	}
	if rule.Title == "" {
		return rule, fmt.Errorf("rule in %s is missing required field 'title'", filePath)
	}
	if rule.Detection == nil {
		return rule, fmt.Errorf("rule in %s is missing required field 'detection'", filePath)
	}

	return rule, nil
}

func (e *Engine) validateRule(rule Rule) error {
	// Check if detection section has valid structure
	if rule.Detection == nil {
		return fmt.Errorf("detection section is nil")
	}

	// Check for common detection patterns
	hasSelection := false
	hasCondition := false
	
	for key, value := range rule.Detection {
		if key == "selection" {
			hasSelection = true
			if value == nil {
				return fmt.Errorf("selection section is nil")
			}
		}
		if key == "condition" {
			hasCondition = true
			if value == nil {
				return fmt.Errorf("condition section is nil")
			}
		}
	}

	if !hasSelection {
		return fmt.Errorf("detection section missing 'selection'")
	}
	if !hasCondition {
		return fmt.Errorf("detection section missing 'condition'")
	}

	return nil
}

func (e *Engine) Evaluate(entry parser.LogEntry) []string {
	var matchedRules []string

	for _, rule := range e.rules {
		if e.matchesRule(entry, rule) {
			matchedRules = append(matchedRules, rule.ID)
		}
	}

	return matchedRules
}

func (e *Engine) matchesRule(entry parser.LogEntry, rule Rule) bool {
	// Check if rule applies to this log source
	if !e.matchesLogSource(entry, rule.Logsource) {
		return false
	}

	// Evaluate detection logic
	return e.evaluateDetection(entry, rule.Detection)
}

func (e *Engine) matchesLogSource(entry parser.LogEntry, logSource LogSource) bool {
	// Simple log source matching - can be enhanced
	if logSource.Category != "" && logSource.Category != entry.Category {
		return false
	}
	if logSource.Product != "" && logSource.Product != entry.Product {
		return false
	}
	if logSource.Service != "" && logSource.Service != entry.Service {
		return false
	}
	return true
}

func (e *Engine) evaluateDetection(entry parser.LogEntry, detection map[string]interface{}) bool {
	// Evaluate selection criteria
	matches := make(map[string]bool)
	
	// Get selection criteria
	selection, ok := detection["selection"].(map[string]interface{})
	if !ok {
		return false
	}
	
	for field, criteria := range selection {
		matches[field] = e.evaluateField(entry, field, criteria)
	}

	// Evaluate condition
	condition, ok := detection["condition"].(string)
	if !ok {
		return false
	}
	
	return e.evaluateCondition(matches, condition)
}

func (e *Engine) evaluateField(entry parser.LogEntry, field string, criteria interface{}) bool {
	fieldValue := e.getFieldValue(entry, field)
	
	switch v := criteria.(type) {
	case string:
		return e.matchString(fieldValue, v)
	case []interface{}:
		for _, item := range v {
			var itemStr string
			switch val := item.(type) {
			case string:
				itemStr = val
			case int:
				itemStr = fmt.Sprintf("%d", val)
			case float64:
				itemStr = fmt.Sprintf("%.0f", val)
			default:
				itemStr = fmt.Sprintf("%v", val)
			}
			if e.matchString(fieldValue, itemStr) {
				return true
			}
		}
		return false
	case map[string]interface{}:
		return e.evaluateFieldModifiers(fieldValue, v)
	default:
		return false
	}
}

func (e *Engine) getFieldValue(entry parser.LogEntry, field string) string {
	switch field {
	case "message":
		return entry.Message
	case "hostname":
		return entry.Hostname
	case "program":
		return entry.Program
	case "pid":
		return entry.PID
	case "timestamp":
		return entry.Timestamp
	default:
		// Check custom fields
		if val, ok := entry.Fields[field]; ok {
			return val
		}
		return ""
	}
}

func (e *Engine) matchString(value, pattern string) bool {
	// Handle regex patterns
	if strings.HasPrefix(pattern, "|re|") {
		regex := strings.TrimPrefix(pattern, "|re|")
		matched, _ := regexp.MatchString(regex, value)
		return matched
	}
	
	// Handle contains patterns
	if strings.HasPrefix(pattern, "|contains|") {
		substring := strings.TrimPrefix(pattern, "|contains|")
		return strings.Contains(strings.ToLower(value), strings.ToLower(substring))
	}
	
	// Handle startswith patterns
	if strings.HasPrefix(pattern, "|startswith|") {
		prefix := strings.TrimPrefix(pattern, "|startswith|")
		return strings.HasPrefix(strings.ToLower(value), strings.ToLower(prefix))
	}
	
	// Handle endswith patterns
	if strings.HasPrefix(pattern, "|endswith|") {
		suffix := strings.TrimPrefix(pattern, "|endswith|")
		return strings.HasSuffix(strings.ToLower(value), strings.ToLower(suffix))
	}
	
	// Default: case-insensitive contains
	return strings.Contains(strings.ToLower(value), strings.ToLower(pattern))
}

func (e *Engine) evaluateFieldModifiers(value string, modifiers map[string]interface{}) bool {
	for modifier, criteria := range modifiers {
		switch modifier {
		case "contains":
			if criteriaStr, ok := criteria.(string); ok {
				return strings.Contains(strings.ToLower(value), strings.ToLower(criteriaStr))
			}
		case "startswith":
			if criteriaStr, ok := criteria.(string); ok {
				return strings.HasPrefix(strings.ToLower(value), strings.ToLower(criteriaStr))
			}
		case "endswith":
			if criteriaStr, ok := criteria.(string); ok {
				return strings.HasSuffix(strings.ToLower(value), strings.ToLower(criteriaStr))
			}
		case "re":
			if criteriaStr, ok := criteria.(string); ok {
				matched, _ := regexp.MatchString(criteriaStr, value)
				return matched
			}
		}
	}
	return false
}

func (e *Engine) evaluateCondition(matches map[string]bool, condition string) bool {
	if condition == "" {
		condition = "selection"
	}
	
	// Simple condition evaluation - can be enhanced for complex logic
	condition = strings.ToLower(condition)
	
	if condition == "selection" {
		// All selection criteria must match
		for _, match := range matches {
			if !match {
				return false
			}
		}
		return len(matches) > 0
	}
	
	if strings.Contains(condition, " and ") {
		parts := strings.Split(condition, " and ")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if !matches[part] {
				return false
			}
		}
		return true
	}
	
	if strings.Contains(condition, " or ") {
		parts := strings.Split(condition, " or ")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if matches[part] {
				return true
			}
		}
		return false
	}
	
	// Single field condition
	return matches[condition]
}
