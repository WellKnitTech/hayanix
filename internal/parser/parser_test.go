package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewParser(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		filePath string
		wantErr  bool
	}{
		{
			name:     "valid syslog parser",
			target:   "syslog",
			filePath: "/tmp/test.log",
			wantErr:  false,
		},
		{
			name:     "valid journald parser",
			target:   "journald",
			filePath: "/tmp/test.log",
			wantErr:  false,
		},
		{
			name:     "valid auditd parser",
			target:   "auditd",
			filePath: "/tmp/test.log",
			wantErr:  false,
		},
		{
			name:     "invalid parser type",
			target:   "invalid",
			filePath: "/tmp/test.log",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewParser(tt.target, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewParser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyslogParser_Parse(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	
	testContent := `Jan  1 10:30:15 server1 sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2
Jan  1 10:30:16 server1 sshd[1234]: Failed password for root from 192.168.1.100 port 22 ssh2
Jan  1 10:31:22 server1 sudo: user : TTY=pts/0 ; PWD=/home/user ; USER=root ; COMMAND=/bin/bash`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewSyslogParser(testFile)
	entries, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Test first entry
	firstEntry := entries[0]
	if firstEntry.Hostname != "server1" {
		t.Errorf("Expected hostname 'server1', got '%s'", firstEntry.Hostname)
	}
	if firstEntry.Program != "sshd[1234]" {
		t.Errorf("Expected program 'sshd[1234]', got '%s'", firstEntry.Program)
	}
	if firstEntry.PID != "1234" {
		t.Errorf("Expected PID '1234', got '%s'", firstEntry.PID)
	}
	if !contains(firstEntry.Message, "Failed password") {
		t.Errorf("Expected message to contain 'Failed password', got '%s'", firstEntry.Message)
	}
	
	// Test timestamp format (should be ISO 8601)
	expectedYear := time.Now().Year()
	expectedTimestamp := fmt.Sprintf("%d-01-01T10:30:15.000", expectedYear)
	if firstEntry.Timestamp != expectedTimestamp {
		t.Errorf("Expected timestamp '%s', got '%s'", expectedTimestamp, firstEntry.Timestamp)
	}
}

func TestJournaldParser_Parse(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	
	testContent := `2025-01-01T10:30:15Z server1 systemd[1]: Started Network Manager
2025-01-01T10:30:16.123Z server1 sshd[1234]: Accepted publickey for root from 192.168.1.100 port 22 ssh2`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewJournaldParser(testFile)
	entries, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Test first entry
	firstEntry := entries[0]
	if firstEntry.Hostname != "server1" {
		t.Errorf("Expected hostname 'server1', got '%s'", firstEntry.Hostname)
	}
	if firstEntry.Program != "systemd[1]" {
		t.Errorf("Expected program 'systemd[1]', got '%s'", firstEntry.Program)
	}
	if firstEntry.Service != "journald" {
		t.Errorf("Expected service 'journald', got '%s'", firstEntry.Service)
	}
	
	// Test timestamp format (should be ISO 8601)
	if firstEntry.Timestamp != "2025-01-01T10:30:15.000" {
		t.Errorf("Expected timestamp '2025-01-01T10:30:15.000', got '%s'", firstEntry.Timestamp)
	}
}

func TestAuditdParser_Parse(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")
	
	testContent := `type=SYSCALL msg=audit(1640999999.123:456): arch=c000003e syscall=open success=yes exit=3 a0=7fff12345678 a1=0 a2=1b6 a3=0 items=1 ppid=1234 pid=5678 auid=1000 uid=1000 gid=1000 euid=1000 suid=1000 fsuid=1000 egid=1000 sgid=1000 fsgid=1000 tty=pts0 ses=1 comm="bash" exe="/bin/bash" key="test-key"`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewAuditdParser(testFile)
	entries, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}

	// Test first entry
	firstEntry := entries[0]
	if firstEntry.Hostname != "localhost" {
		t.Errorf("Expected hostname 'localhost', got '%s'", firstEntry.Hostname)
	}
	if firstEntry.Program != "auditd" {
		t.Errorf("Expected program 'auditd', got '%s'", firstEntry.Program)
	}
	if firstEntry.PID != "456" {
		t.Errorf("Expected PID '456', got '%s'", firstEntry.PID)
	}
	if firstEntry.Service != "auditd" {
		t.Errorf("Expected service 'auditd', got '%s'", firstEntry.Service)
	}
	
	// Test that fields were parsed
	if firstEntry.Fields["arch"] != "c000003e" {
		t.Errorf("Expected arch field 'c000003e', got '%s'", firstEntry.Fields["arch"])
	}
	if firstEntry.Fields["syscall"] != "open" {
		t.Errorf("Expected syscall field 'open', got '%s'", firstEntry.Fields["syscall"])
	}
}

func TestParser_EmptyFile(t *testing.T) {
	// Create an empty test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.log")
	
	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewSyslogParser(testFile)
	entries, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for empty file, got %d", len(entries))
	}
}

func TestParser_NonExistentFile(t *testing.T) {
	parser := NewSyslogParser("/non/existent/file.log")
	_, err := parser.Parse()
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// Helper function
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
