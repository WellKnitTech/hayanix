package output

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/wellknittech/hayanix/internal/parser"
)

func TestNewOutputter(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "table format",
			format: "table",
			want:   "table",
		},
		{
			name:   "csv format",
			format: "csv",
			want:   "csv",
		},
		{
			name:   "json format",
			format: "json",
			want:   "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputter := NewOutputter(tt.format)
			if outputter.format != tt.want {
				t.Errorf("NewOutputter() format = %v, want %v", outputter.format, tt.want)
			}
		})
	}
}

func TestOutputter_Write(t *testing.T) {
	// Create test log entries
	entries := []parser.LogEntry{
		{
			Timestamp: "2025-01-01T10:30:15.000",
			Hostname:  "server1",
			Program:   "sshd[1234]",
			PID:       "1234",
			Message:   "Failed password for root from 192.168.1.100 port 22 ssh2",
			Category:  "process",
			Product:   "linux",
			Service:   "syslog",
			Fields:    make(map[string]string),
			MatchedRules: []string{"test-rule-001"},
		},
		{
			Timestamp: "2025-01-01T10:30:16.000",
			Hostname:  "server1",
			Program:   "sshd[1234]",
			PID:       "1234",
			Message:   "Failed password for root from 192.168.1.100 port 22 ssh2",
			Category:  "process",
			Product:   "linux",
			Service:   "syslog",
			Fields:    make(map[string]string),
			MatchedRules: []string{"test-rule-001"},
		},
	}

	t.Run("table format", func(t *testing.T) {
		outputter := NewOutputter("table")
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputter.Write(entries)
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Check that output contains expected elements
		if !strings.Contains(output, "TIMESTAMP") {
			t.Error("Expected output to contain 'TIMESTAMP' header")
		}
		if !strings.Contains(output, "HOSTNAME") {
			t.Error("Expected output to contain 'HOSTNAME' header")
		}
		if !strings.Contains(output, "PROGRAM") {
			t.Error("Expected output to contain 'PROGRAM' header")
		}
		if !strings.Contains(output, "MESSAGE") {
			t.Error("Expected output to contain 'MESSAGE' header")
		}
		if !strings.Contains(output, "TAGS") {
			t.Error("Expected output to contain 'TAGS' header")
		}
		if !strings.Contains(output, "2025-01-01T10:30:15.000") {
			t.Error("Expected output to contain timestamp")
		}
		if !strings.Contains(output, "server1") {
			t.Error("Expected output to contain hostname")
		}
	})

	t.Run("csv format", func(t *testing.T) {
		outputter := NewOutputter("csv")
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputter.Write(entries)
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Check that output contains expected CSV elements
		if !strings.Contains(output, "timestamp,hostname,program,pid,message,matched_rules") {
			t.Error("Expected output to contain CSV header")
		}
		if !strings.Contains(output, "2025-01-01T10:30:15.000,server1,sshd[1234],1234") {
			t.Error("Expected output to contain CSV data")
		}
	})

	t.Run("json format", func(t *testing.T) {
		outputter := NewOutputter("json")
		
		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputter.Write(entries)
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Check that output contains expected JSON elements
		if !strings.Contains(output, `"timestamp"`) {
			t.Error("Expected output to contain JSON timestamp field")
		}
		if !strings.Contains(output, `"hostname"`) {
			t.Error("Expected output to contain JSON hostname field")
		}
		if !strings.Contains(output, `"program"`) {
			t.Error("Expected output to contain JSON program field")
		}
		if !strings.Contains(output, `"message"`) {
			t.Error("Expected output to contain JSON message field")
		}
		if !strings.Contains(output, `"matched_rules"`) {
			t.Error("Expected output to contain JSON matched_rules field")
		}
		if !strings.Contains(output, `"2025-01-01T10:30:15.000"`) {
			t.Error("Expected output to contain timestamp value")
		}
		if !strings.Contains(output, `"server1"`) {
			t.Error("Expected output to contain hostname value")
		}
	})
}

func TestOutputter_WriteEmptyEntries(t *testing.T) {
	outputter := NewOutputter("table")
	
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputter.Write([]parser.LogEntry{})
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Should not crash and should show "No matching entries found"
	if !strings.Contains(output, "No matching entries found") {
		t.Error("Expected output to contain 'No matching entries found' for empty entries")
	}
}

func TestOutputter_InvalidFormat(t *testing.T) {
	// Test with invalid format - should default to table
	outputter := NewOutputter("invalid")
	if outputter.format != "table" {
		t.Errorf("Expected format to default to 'table', got '%s'", outputter.format)
	}
}
