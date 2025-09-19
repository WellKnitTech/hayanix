package collection

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wellknittech/hayanix/internal/output"
	"github.com/wellknittech/hayanix/internal/parser"
	"github.com/wellknittech/hayanix/internal/rules"
)

type CollectionAnalyzer struct {
	collection *Collection
	ruleEngine *rules.Engine
	outputter  *output.Outputter
	verbose    bool
}

type AnalysisResult struct {
	LogFile     string
	LogType     string
	Entries     []parser.LogEntry
	MatchCount  int
	ProcessTime time.Duration
	Error       error
}

type CollectionResult struct {
	Collection     *Collection
	Results        []AnalysisResult
	TotalMatches   int
	TotalFiles     int
	ProcessedFiles int
	FailedFiles    int
	TotalTime      time.Duration
}

func NewCollectionAnalyzer(collection *Collection, rulesDir string, outputFormat string, verbose bool) (*CollectionAnalyzer, error) {
	// Load rules
	ruleEngine, err := rules.NewEngine(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}

	// Create outputter
	outputter := output.NewOutputter(outputFormat)

	return &CollectionAnalyzer{
		collection: collection,
		ruleEngine: ruleEngine,
		outputter:  outputter,
		verbose:    verbose,
	}, nil
}

func (ca *CollectionAnalyzer) AnalyzeCollection() (*CollectionResult, error) {
	startTime := time.Now()

	result := &CollectionResult{
		Collection: ca.collection,
		Results:    make([]AnalysisResult, 0),
		TotalFiles: len(ca.collection.LogFiles),
	}

	if ca.verbose {
		log.Printf("Starting collection analysis of %d files", result.TotalFiles)
	}

	// Process each log file
	for _, logFile := range ca.collection.LogFiles {
		analysisResult := ca.analyzeLogFile(logFile)
		result.Results = append(result.Results, analysisResult)

		if analysisResult.Error != nil {
			result.FailedFiles++
			if ca.verbose {
				log.Printf("Failed to analyze %s: %v", logFile.Path, analysisResult.Error)
			}
		} else {
			result.ProcessedFiles++
			result.TotalMatches += analysisResult.MatchCount
		}
	}

	result.TotalTime = time.Since(startTime)

	if ca.verbose {
		log.Printf("Collection analysis completed: %d processed, %d failed, %d total matches in %v",
			result.ProcessedFiles, result.FailedFiles, result.TotalMatches, result.TotalTime)
	}

	return result, nil
}

func (ca *CollectionAnalyzer) analyzeLogFile(logFile LogFile) AnalysisResult {
	startTime := time.Now()

	result := AnalysisResult{
		LogFile: logFile.Path,
		LogType: logFile.Type,
	}

	// Create parser
	logParser, err := parser.NewParser(logFile.Type, logFile.Path)
	if err != nil {
		result.Error = fmt.Errorf("failed to create parser: %w", err)
		return result
	}

	// Parse log file
	entries, err := logParser.Parse()
	if err != nil {
		result.Error = fmt.Errorf("failed to parse log file: %w", err)
		return result
	}

	// Evaluate rules against entries
	var matchingEntries []parser.LogEntry
	for _, entry := range entries {
		matches := ca.ruleEngine.Evaluate(entry)
		if len(matches) > 0 {
			entry.MatchedRules = matches
			matchingEntries = append(matchingEntries, entry)
		}
	}

	result.Entries = matchingEntries
	result.MatchCount = len(matchingEntries)
	result.ProcessTime = time.Since(startTime)

	if ca.verbose {
		log.Printf("Analyzed %s: %d entries, %d matches in %v",
			logFile.Path, len(entries), result.MatchCount, result.ProcessTime)
	}

	return result
}

func (ca *CollectionAnalyzer) WriteResults(result *CollectionResult) error {
	// Collect all matching entries from all files
	var allEntries []parser.LogEntry
	for _, analysisResult := range result.Results {
		if analysisResult.Error == nil {
			allEntries = append(allEntries, analysisResult.Entries...)
		}
	}

	// Sort entries by timestamp if possible
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Timestamp < allEntries[j].Timestamp
	})

	// Write results
	return ca.outputter.Write(allEntries)
}

func (ca *CollectionAnalyzer) WriteSummary(result *CollectionResult) {
	fmt.Println("📊 Collection Analysis Summary")
	fmt.Println("=============================")
	fmt.Printf("Base Path: %s\n", result.Collection.BasePath)
	fmt.Printf("Total Files: %d\n", result.TotalFiles)
	fmt.Printf("Processed Files: %d\n", result.ProcessedFiles)
	fmt.Printf("Failed Files: %d\n", result.FailedFiles)
	fmt.Printf("Total Matches: %d\n", result.TotalMatches)
	fmt.Printf("Processing Time: %v\n", result.TotalTime)
	fmt.Println()

	// Show file breakdown
	fmt.Println("📁 Files by Type:")
	for logType, count := range result.Collection.Summary.FilesByType {
		fmt.Printf("  %s: %d files\n", logType, count)
	}
	fmt.Println()

	// Show results by file
	if len(result.Results) > 0 {
		fmt.Println("📋 Results by File:")
		for _, analysisResult := range result.Results {
			status := "✅"
			if analysisResult.Error != nil {
				status = "❌"
			}

			relativePath, _ := filepath.Rel(result.Collection.BasePath, analysisResult.LogFile)
			fmt.Printf("  %s %s (%s): %d matches\n",
				status, relativePath, analysisResult.LogType, analysisResult.MatchCount)
		}
	}
}

func (ca *CollectionAnalyzer) WriteDetailedResults(result *CollectionResult) error {
	// Write results for each file separately
	for _, analysisResult := range result.Results {
		if analysisResult.Error != nil {
			continue
		}

		if len(analysisResult.Entries) == 0 {
			continue
		}

		// Create file-specific outputter
		relativePath, _ := filepath.Rel(result.Collection.BasePath, analysisResult.LogFile)
		fmt.Printf("\n📄 Results for %s (%s):\n", relativePath, analysisResult.LogType)
		fmt.Println(strings.Repeat("=", 50))

		fileOutputter := output.NewOutputter("table")
		if err := fileOutputter.Write(analysisResult.Entries); err != nil {
			return fmt.Errorf("failed to write results for %s: %w", analysisResult.LogFile, err)
		}
	}

	return nil
}
