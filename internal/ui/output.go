package ui

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// OutputHighlighter handles real-time output streaming with intelligent highlighting
type OutputHighlighter struct {
	showTimestamps bool
	indentLevel    int
}

// NewOutputHighlighter creates a new output highlighter
func NewOutputHighlighter(showTimestamps bool, indentLevel int) *OutputHighlighter {
	return &OutputHighlighter{
		showTimestamps: showTimestamps,
		indentLevel:    indentLevel,
	}
}

// Pattern matchers for different types of output
var (
	errorPatterns = regexp.MustCompile(`(?i)(error|failed|fatal|panic|exception|denied|cannot|unable to|not found|invalid|illegal)`)

	warningPatterns = regexp.MustCompile(`(?i)(warning|warn|deprecated|caution|note|notice)`)

	successPatterns = regexp.MustCompile(`(?i)(success|successful|completed|installed|configured|done|finished|ok|ready|active|enabled|started)`)

	statusPatterns = regexp.MustCompile(`(?i)(downloading|installing|configuring|building|compiling|updating|connecting|loading|processing|running|starting|stopping)`)

	progressPatterns = regexp.MustCompile(`(\d+%|\d+/\d+|\[\d+/\d+\]|\d+\.\d+\s*(MB|GB|KB))`)
)

// StreamOutput processes output line by line with highlighting
func (oh *OutputHighlighter) StreamOutput(reader io.Reader, prefix string) error {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()

		// Build the formatted line
		var formattedLine strings.Builder

		// Add indent
		for i := 0; i < oh.indentLevel; i++ {
			formattedLine.WriteString("  ")
		}

		// Add timestamp if enabled
		if oh.showTimestamps {
			timestamp := time.Now().Format("15:04:05")
			formattedLine.WriteString(TimestampText(fmt.Sprintf("[%s] ", timestamp)))
		}

		// Add prefix if provided
		if prefix != "" {
			formattedLine.WriteString(prefix)
		}

		// Apply highlighting based on content
		highlightedLine := oh.highlightLine(line)
		formattedLine.WriteString(highlightedLine)

		// Print the formatted line
		fmt.Println(formattedLine.String())
	}

	return scanner.Err()
}

// highlightLine applies color highlighting based on line content
func (oh *OutputHighlighter) highlightLine(line string) string {
	// Check for different patterns in order of priority
	switch {
	case errorPatterns.MatchString(line):
		return ErrorMessage(line)
	case warningPatterns.MatchString(line):
		return WarningMessage(line)
	case successPatterns.MatchString(line):
		return SuccessMessage(line)
	case statusPatterns.MatchString(line):
		return InfoMessage(line)
	case progressPatterns.MatchString(line):
		// Highlight progress indicators within the line
		highlighted := progressPatterns.ReplaceAllStringFunc(line, func(match string) string {
			return HighlightText(match)
		})
		return Blue.Sprint(highlighted)
	default:
		return line
	}
}

// PrintKnightMessage prints a themed knight message
func PrintKnightMessage(message string) {
	fmt.Println(KnightMessage("🛡️  " + message))
}

// PrintSuccessMessage prints a themed success message
func PrintSuccessMessage(message string) {
	fmt.Println(SuccessMessage("🏆 " + message))
}

// PrintErrorMessage prints a themed error message
func PrintErrorMessage(message string) {
	fmt.Println(ErrorMessage("❌ " + message))
}

// PrintWarningMessage prints a themed warning message
func PrintWarningMessage(message string) {
	fmt.Println(WarningMessage("⚠️  " + message))
}

// PrintInfoMessage prints a themed info message
func PrintInfoMessage(message string) {
	fmt.Println(InfoMessage("🔍 " + message))
}

// PrintAIMessage prints a themed AI consultation message
func PrintAIMessage(message string) {
	fmt.Println(AIMessage("🧙 " + message))
}

// Default template instance
var defaultTemplate = DefaultTemplate()

// PrintSeparator prints a themed separator
func PrintSeparator() {
	defaultTemplate.PrintStandardSeparator()
}

// PrintExecutionHeader prints a header for command/script execution
func PrintExecutionHeader(title string) {
	defaultTemplate.PrintMainSection(title)
}

// PrintPhaseHeader prints a phase header
func PrintPhaseHeader(icon, phase string) {
	defaultTemplate.PrintPhase(icon, phase)
}

// PrintCommandBox prints a command in a structured box
func PrintCommandBox(command string) {
	defaultTemplate.PrintCommandBox(command)
}

// PrintScriptBox prints a script in a structured box
func PrintScriptBox(title string, scriptLines []string) {
	defaultTemplate.PrintScriptBox(title, scriptLines)
}

// PrintStatusBox prints a status message in a box
func PrintStatusBox(status, message, statusType string) {
	defaultTemplate.PrintStatusBox(status, message, statusType)
}

// PrintConfigBox prints configuration in a structured table
func PrintConfigBox(configs map[string]string) {
	defaultTemplate.PrintConfigTable(configs)
}
