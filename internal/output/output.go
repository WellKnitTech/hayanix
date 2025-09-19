package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/wellknittech/hayanix/internal/parser"
)

type Outputter struct {
	format string
}

func NewOutputter(format string) *Outputter {
	// Default to table format for invalid formats
	if format != "table" && format != "csv" && format != "json" {
		format = "table"
	}
	return &Outputter{format: format}
}

func (o *Outputter) Write(entries []parser.LogEntry) error {
	switch o.format {
	case "table":
		return o.writeTable(entries)
	case "csv":
		return o.writeCSV(entries)
	case "json":
		return o.writeJSON(entries)
	default:
		return fmt.Errorf("unsupported output format: %s", o.format)
	}
}

func (o *Outputter) writeTable(entries []parser.LogEntry) error {
	if len(entries) == 0 {
		fmt.Println("No matching entries found.")
		return nil
	}

	fmt.Printf("Found %d matching entries:\n\n", len(entries))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Timestamp", "Hostname", "Program", "Message", "Tags"})
	table.SetBorder(true)
	table.SetCenterSeparator("|")
	table.SetColumnSeparator("|")
	table.SetRowSeparator("-")

	for _, entry := range entries {
		tags := strings.Join(entry.MatchedRules, ", ")
		message := entry.Message
		if len(message) > 50 {
			message = message[:47] + "..."
		}
		
		table.Append([]string{
			entry.Timestamp,
			entry.Hostname,
			entry.Program,
			message,
			tags,
		})
	}

	table.Render()
	return nil
}

func (o *Outputter) writeCSV(entries []parser.LogEntry) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	header := []string{"timestamp", "hostname", "program", "pid", "message", "matched_rules"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, entry := range entries {
		record := []string{
			entry.Timestamp,
			entry.Hostname,
			entry.Program,
			entry.PID,
			entry.Message,
			strings.Join(entry.MatchedRules, ";"),
		}
		
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

func (o *Outputter) writeJSON(entries []parser.LogEntry) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	
	return encoder.Encode(entries)
}
