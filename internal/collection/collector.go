package collection

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wellknittech/hayanix/internal/parser"
)

type Collector struct {
	basePath string
}

type LogFile struct {
	Path     string
	Type     string
	Size     int64
	Modified string
}

type Collection struct {
	BasePath string
	LogFiles []LogFile
	Summary  CollectionSummary
}

type CollectionSummary struct {
	TotalFiles      int
	TotalSize       int64
	FilesByType     map[string]int
	SizeByType      map[string]int64
	CompatibleTypes []string
}

func NewCollector(basePath string) *Collector {
	return &Collector{
		basePath: basePath,
	}
}

func (c *Collector) DiscoverLogFiles() (*Collection, error) {
	collection := &Collection{
		BasePath: c.basePath,
		LogFiles: make([]LogFile, 0),
		Summary: CollectionSummary{
			FilesByType:     make(map[string]int),
			SizeByType:      make(map[string]int64),
			CompatibleTypes: []string{"syslog", "journald", "auditd"},
		},
	}

	// Check if base path exists
	if _, err := os.Stat(c.basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", c.basePath)
	}

	// Walk through the directory tree
	err := filepath.Walk(c.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is a log file
		logType := c.detectLogType(path, info)
		if logType != "" {
			logFile := LogFile{
				Path:     path,
				Type:     logType,
				Size:     info.Size(),
				Modified: info.ModTime().Format("2006-01-02 15:04:05"),
			}

			collection.LogFiles = append(collection.LogFiles, logFile)
			collection.Summary.FilesByType[logType]++
			collection.Summary.SizeByType[logType] += info.Size()
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Calculate summary
	collection.Summary.TotalFiles = len(collection.LogFiles)
	for _, size := range collection.Summary.SizeByType {
		collection.Summary.TotalSize += size
	}

	return collection, nil
}

func (c *Collector) detectLogType(path string, info os.FileInfo) string {
	// Get file extension and basename
	ext := strings.ToLower(filepath.Ext(path))
	basename := strings.ToLower(filepath.Base(path))
	dirname := strings.ToLower(filepath.Dir(path))

	// Skip very small files (likely not real logs)
	if info.Size() < 100 {
		return ""
	}

	// Skip binary files and common non-log extensions
	skipExts := map[string]bool{
		".exe": true, ".bin": true, ".so": true, ".dll": true,
		".zip": true, ".tar": true, ".gz": true, ".bz2": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true,
		".jpg": true, ".png": true, ".gif": true, ".mp4": true,
	}

	if skipExts[ext] {
		return ""
	}

	// Detect log type based on filename patterns
	logPatterns := map[string][]string{
		"syslog": {
			"messages", "syslog", "auth", "auth.log", "secure",
			"mail", "mail.log", "daemon", "daemon.log",
			"kern", "kern.log", "user", "user.log",
			"local0", "local1", "local2", "local3",
			"local4", "local5", "local6", "local7",
		},
		"journald": {
			"journal", "system.journal", "user.journal",
		},
		"auditd": {
			"audit.log", "audit", "auditd",
		},
	}

	// Check filename patterns
	for logType, patterns := range logPatterns {
		for _, pattern := range patterns {
			if strings.Contains(basename, pattern) {
				return logType
			}
		}
	}

	// Check directory patterns
	dirPatterns := map[string][]string{
		"syslog":   {"log", "logs", "var/log"},
		"journald": {"journal", "systemd"},
		"auditd":   {"audit", "auditd"},
	}

	for logType, patterns := range dirPatterns {
		for _, pattern := range patterns {
			if strings.Contains(dirname, pattern) {
				return logType
			}
		}
	}

	// Check file extension
	if ext == ".log" || ext == "" {
		// Default to syslog for .log files and files without extensions
		return "syslog"
	}

	return ""
}

func (c *Collector) FilterByType(collection *Collection, logType string) *Collection {
	if logType == "" {
		return collection
	}

	filtered := &Collection{
		BasePath: collection.BasePath,
		LogFiles: make([]LogFile, 0),
		Summary: CollectionSummary{
			FilesByType:     make(map[string]int),
			SizeByType:      make(map[string]int64),
			CompatibleTypes: collection.Summary.CompatibleTypes,
		},
	}

	for _, logFile := range collection.LogFiles {
		if logFile.Type == logType {
			filtered.LogFiles = append(filtered.LogFiles, logFile)
			filtered.Summary.FilesByType[logFile.Type]++
			filtered.Summary.SizeByType[logFile.Type] += logFile.Size
		}
	}

	// Recalculate summary
	filtered.Summary.TotalFiles = len(filtered.LogFiles)
	for _, size := range filtered.Summary.SizeByType {
		filtered.Summary.TotalSize += size
	}

	return filtered
}

func (c *Collector) GetCompatibleFiles(collection *Collection) []LogFile {
	compatible := make([]LogFile, 0)

	for _, logFile := range collection.LogFiles {
		// Check if we have a parser for this log type
		if _, err := parser.NewParser(logFile.Type, logFile.Path); err == nil {
			compatible = append(compatible, logFile)
		}
	}

	return compatible
}

func (c *Collector) ValidateCollection(collection *Collection) error {
	if len(collection.LogFiles) == 0 {
		return fmt.Errorf("no compatible log files found in %s", c.basePath)
	}

	// Check if we can parse at least one file
	compatible := c.GetCompatibleFiles(collection)
	if len(compatible) == 0 {
		return fmt.Errorf("no parseable log files found in %s", c.basePath)
	}

	return nil
}
