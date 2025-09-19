package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Parser interface {
	Parse() ([]LogEntry, error)
}

type LogEntry struct {
	Timestamp   string
	Hostname    string
	Program     string
	PID         string
	Message     string
	Category    string
	Product     string
	Service     string
	Fields      map[string]string
	MatchedRules []string
}

func NewParser(target, filePath string) (Parser, error) {
	switch target {
	case "syslog":
		return NewSyslogParser(filePath), nil
	case "journald":
		return NewJournaldParser(filePath), nil
	case "auditd":
		return NewAuditdParser(filePath), nil
	default:
		return nil, fmt.Errorf("unsupported parser target: %s", target)
	}
}

type SyslogParser struct {
	filePath string
}

func NewSyslogParser(filePath string) *SyslogParser {
	return &SyslogParser{filePath: filePath}
}

func (p *SyslogParser) Parse() ([]LogEntry, error) {
	// Check if file exists
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("log file does not exist: %s", p.filePath)
	}

	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", p.filePath, err)
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	
	// Syslog format: Jan 2 15:04:05 hostname program[pid]: message
	syslogRegex := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(\S+)(?:\[(\d+)\])?:\s*(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		matches := syslogRegex.FindStringSubmatch(line)
		if len(matches) < 6 {
			// Try to parse as a continuation line or malformed entry
			if len(entries) > 0 {
				entries[len(entries)-1].Message += " " + line
			}
			continue
		}

		// Parse syslog timestamp and convert to ISO 8601 format
		currentYear := time.Now().Year()
		timestampStr := fmt.Sprintf("%d %s", currentYear, matches[1])
		t, err := time.Parse("2006 Jan  2 15:04:05", timestampStr)
		if err != nil {
			// Fallback to original string if parsing fails
			t, _ = time.Parse("Jan  2 15:04:05", matches[1])
		}

		entry := LogEntry{
			Timestamp:   t.Format("2006-01-02T15:04:05.000"),
			Hostname:    matches[2],
			Program:     matches[3],
			PID:         matches[4],
			Message:     matches[5],
			Category:    "process",
			Product:     "linux",
			Service:     "syslog",
			Fields:      make(map[string]string),
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

type JournaldParser struct {
	filePath string
}

func NewJournaldParser(filePath string) *JournaldParser {
	return &JournaldParser{filePath: filePath}
}

func (p *JournaldParser) Parse() ([]LogEntry, error) {
	// Check if file exists
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("log file does not exist: %s", p.filePath)
	}

	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", p.filePath, err)
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	
	// Journald format: timestamp hostname program[pid]: message
	journaldRegex := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})?)\s+(\S+)\s+(\S+)(?:\[(\d+)\])?:\s*(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		matches := journaldRegex.FindStringSubmatch(line)
		if len(matches) < 6 {
			continue
		}

		// Parse journald timestamp and ensure ISO 8601 format
		t, err := time.Parse(time.RFC3339, matches[1])
		if err != nil {
			// Try parsing without timezone
			t, err = time.Parse("2006-01-02T15:04:05", matches[1])
			if err != nil {
				// Try parsing with microseconds
				t, err = time.Parse("2006-01-02T15:04:05.000000", matches[1])
				if err != nil {
					// Fallback to original string
					t = time.Now()
				}
			}
		}

		entry := LogEntry{
			Timestamp:   t.Format("2006-01-02T15:04:05.000"),
			Hostname:    matches[2],
			Program:     matches[3],
			PID:         matches[4],
			Message:     matches[5],
			Category:    "process",
			Product:     "linux",
			Service:     "journald",
			Fields:      make(map[string]string),
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

type AuditdParser struct {
	filePath string
}

func NewAuditdParser(filePath string) *AuditdParser {
	return &AuditdParser{filePath: filePath}
}

func (p *AuditdParser) Parse() ([]LogEntry, error) {
	// Check if file exists
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("log file does not exist: %s", p.filePath)
	}

	file, err := os.Open(p.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", p.filePath, err)
	}
	defer file.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	
	// Auditd format: type=... msg=audit(timestamp:pid): ...
	auditdRegex := regexp.MustCompile(`^type=(\S+)\s+msg=audit\((\d+\.\d+):(\d+)\):\s*(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		matches := auditdRegex.FindStringSubmatch(line)
		if len(matches) < 5 {
			continue
		}

		// Parse timestamp and convert to ISO 8601 format
		timestamp := time.Unix(int64(mustParseFloat(matches[2])), 0).Format("2006-01-02T15:04:05.000")

		entry := LogEntry{
			Timestamp:   timestamp,
			Hostname:    "localhost", // Auditd doesn't include hostname
			Program:     "auditd",
			PID:         matches[3],
			Message:     matches[4],
			Category:    "audit",
			Product:     "linux",
			Service:     "auditd",
			Fields:      make(map[string]string),
		}

		// Parse audit fields
		fieldRegex := regexp.MustCompile(`(\w+)=([^\s]+)`)
		fieldMatches := fieldRegex.FindAllStringSubmatch(matches[4], -1)
		for _, fieldMatch := range fieldMatches {
			if len(fieldMatch) >= 3 {
				entry.Fields[fieldMatch[1]] = fieldMatch[2]
			}
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

func mustParseFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return val
}
