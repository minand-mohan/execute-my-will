package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Template system for consistent UI structure

// ANSI escape sequence regex for stripping color codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape sequences to get visible text length
func stripANSI(text string) string {
	return ansiRegex.ReplaceAllString(text, "")
}

// visibleLen returns the visual character count using proper terminal width calculation
func visibleLen(text string) int {
	// First strip ANSI codes
	cleaned := stripANSI(text)
	// Use go-runewidth to get proper terminal display width (handles emojis correctly)
	return runewidth.StringWidth(cleaned)
}

// UITemplate represents different UI layout templates
type UITemplate struct {
	width int
}

// NewTemplate creates a new UI template with specified width
func NewTemplate(width int) *UITemplate {
	if width < 40 {
		width = 60 // Default minimum width
	}
	return &UITemplate{width: width}
}

// DefaultTemplate returns a template with standard width
func DefaultTemplate() *UITemplate {
	return NewTemplate(60)
}

// Section templates
func (t *UITemplate) PrintMainSection(title string) {
	fmt.Println()
	border := strings.Repeat("━", t.width)
	fmt.Println(Gold.Sprint(border))
	// Calculate proper padding using visible length
	contentWidth := t.width - 4 // Account for "┃ " and " ┃"
	titleVisibleLen := visibleLen(title)
	padding := contentWidth - titleVisibleLen

	fmt.Printf("%s %s%s %s\n",
		Gold.Sprint("┃"),
		title,
		strings.Repeat(" ", padding),
		Gold.Sprint("┃"))
	fmt.Println(Gold.Sprint(border))
	fmt.Println()
}

func (t *UITemplate) PrintSubSection(title string) {
	fmt.Println()
	border := strings.Repeat("─", t.width)
	fmt.Println(Gray.Sprint(border))
	fmt.Printf("%s %s\n", Gold.Sprint("▶"), Gold.Sprint(title))
	fmt.Println(Gray.Sprint(border))
}

func (t *UITemplate) PrintPhase(icon, phase string) {
	fmt.Println()
	remaining := t.width - len(phase) - len(icon) - 6
	if remaining < 0 {
		remaining = 0
	}
	fmt.Printf("%s %s %s %s\n",
		Gold.Sprint("┌─"),
		Gold.Sprint(fmt.Sprintf("%s %s", icon, phase)),
		Gold.Sprint(strings.Repeat("─", remaining)),
		Gold.Sprint("┐"))
	fmt.Println()
}

// Box templates
func (t *UITemplate) PrintBox(title string, content []string) {
	// Top border
	fmt.Printf("%s%s%s\n",
		Gold.Sprint("╭"),
		Gold.Sprint(strings.Repeat("─", t.width-2)),
		Gold.Sprint("╮"))

	// Title if provided
	if title != "" {
		// Calculate padding to center the title using visible length
		contentWidth := t.width - 4 // Account for "│ " and " │"
		titleVisibleLen := visibleLen(title)
		if titleVisibleLen > contentWidth {
			// Title too long, truncate
			title = title[:contentWidth-3] + "..."
			titleVisibleLen = contentWidth
		}
		leftPadding := (contentWidth - titleVisibleLen) / 2
		rightPadding := contentWidth - titleVisibleLen - leftPadding

		fmt.Printf("%s %s%s%s %s\n",
			Gold.Sprint("│"),
			strings.Repeat(" ", leftPadding),
			Gold.Sprint(title),
			strings.Repeat(" ", rightPadding),
			Gold.Sprint("│"))

		// Separator under title
		fmt.Printf("%s%s%s\n",
			Gold.Sprint("├"),
			Gold.Sprint(strings.Repeat("─", t.width-2)),
			Gold.Sprint("┤"))
	}

	// Content lines
	for _, line := range content {
		t.printBoxLine(line)
	}

	// Bottom border
	fmt.Printf("%s%s%s\n",
		Gold.Sprint("╰"),
		Gold.Sprint(strings.Repeat("─", t.width-2)),
		Gold.Sprint("╯"))
	fmt.Println()
}

func (t *UITemplate) printBoxLine(content string) {
	maxWidth := t.width - 4 // Account for "│ " and " │"

	// Handle empty lines
	if strings.TrimSpace(content) == "" {
		fmt.Printf("%s %s %s\n",
			Gold.Sprint("│"),
			strings.Repeat(" ", maxWidth),
			Gold.Sprint("│"))
		return
	}

	// Handle lines that fit within width
	contentVisibleLen := visibleLen(content)
	if contentVisibleLen <= maxWidth {
		padding := maxWidth - contentVisibleLen
		// Debug: Check if padding is negative or zero
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("%s %s%s %s\n", Gold.Sprint("│"), content, strings.Repeat(" ", padding), Gold.Sprint("│"))
		return
	}

	// Handle long lines with word wrapping
	words := strings.Fields(content)
	var line strings.Builder

	for _, word := range words {
		// Calculate visible length including what's already in the line
		currentLineVisible := visibleLen(line.String())
		wordVisible := visibleLen(word)

		// If the word itself is longer than maxWidth, we need to break it
		if wordVisible > maxWidth {
			// First, print the current line if it has content
			if currentLineVisible > 0 {
				lineContent := line.String()
				linePadding := maxWidth - visibleLen(lineContent)
				if linePadding < 0 {
					linePadding = 0
				}
				fmt.Printf("%s %s%s %s\n",
					Gold.Sprint("│"),
					lineContent,
					strings.Repeat(" ", linePadding),
					Gold.Sprint("│"))
				line.Reset()
			}

			// Break the long word into chunks
			runes := []rune(word)
			for len(runes) > 0 {
				chunk := ""
				chunkRunes := 0

				// Build chunk that fits within maxWidth
				for i, r := range runes {
					testChunk := chunk + string(r)
					if visibleLen(testChunk) > maxWidth {
						break
					}
					chunk = testChunk
					chunkRunes = i + 1
				}

				// If we couldn't fit even one character, take it anyway to avoid infinite loop
				if chunkRunes == 0 {
					chunkRunes = 1
					chunk = string(runes[0])
				}

				// Print the chunk
				chunkPadding := maxWidth - visibleLen(chunk)
				if chunkPadding < 0 {
					chunkPadding = 0
				}
				fmt.Printf("%s %s%s %s\n",
					Gold.Sprint("│"),
					chunk,
					strings.Repeat(" ", chunkPadding),
					Gold.Sprint("│"))

				// Remove processed runes
				runes = runes[chunkRunes:]
			}
			continue
		}

		// Check if adding this word would exceed the width
		if currentLineVisible > 0 && currentLineVisible+wordVisible+1 > maxWidth {
			// Print current line with proper padding
			lineContent := line.String()
			linePadding := maxWidth - visibleLen(lineContent)
			if linePadding < 0 {
				linePadding = 0
			}
			fmt.Printf("%s %s%s %s\n",
				Gold.Sprint("│"),
				lineContent,
				strings.Repeat(" ", linePadding),
				Gold.Sprint("│"))
			line.Reset()
		}

		if line.Len() > 0 {
			line.WriteString(" ")
		}
		line.WriteString(word)
	}

	// Print remaining content with proper padding
	if line.Len() > 0 {
		lineContent := line.String()
		linePadding := maxWidth - visibleLen(lineContent)
		if linePadding < 0 {
			linePadding = 0
		}
		fmt.Printf("%s %s%s %s\n",
			Gold.Sprint("│"),
			lineContent,
			strings.Repeat(" ", linePadding),
			Gold.Sprint("│"))
	}
}

// Command/Script display templates
func (t *UITemplate) PrintCommandBox(command string) {
	t.PrintBox("⚔️  PROPOSED COMMAND", []string{
		"",
		CommandText(command),
		"",
	})
}

func (t *UITemplate) PrintScriptBox(title string, scriptLines []string) {
	content := []string{""}
	for _, line := range scriptLines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// Comment line
			comment := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "#"))
			content = append(content, CommentText("• "+comment))
		} else if strings.TrimSpace(line) != "" {
			// Command line
			content = append(content, CommandText("→ "+strings.TrimSpace(line)))
		}
	}
	content = append(content, "")

	t.PrintBox("📜 "+title, content)
}

// Status and message templates
func (t *UITemplate) PrintStatusBox(status, message string, statusType string) {
	var icon string
	var colorFunc func(string) string

	switch statusType {
	case "success":
		icon = "🏆"
		colorFunc = SuccessMessage
	case "error":
		icon = "❌"
		colorFunc = ErrorMessage
	case "warning":
		icon = "⚠️"
		colorFunc = WarningMessage
	case "info":
		icon = "ℹ️"
		colorFunc = InfoMessage
	default:
		icon = "📋"
		colorFunc = func(s string) string { return s }
	}

	t.PrintBox(fmt.Sprintf("%s %s", icon, status), []string{
		"",
		colorFunc(message),
		"",
	})
}

// Configuration display template
func (t *UITemplate) PrintConfigTable(configs map[string]string) {
	// Find the longest key for alignment
	maxKeyLen := 0
	for key := range configs {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	content := []string{""}
	for key, value := range configs {
		line := fmt.Sprintf("%-*s : %s", maxKeyLen, key, value)
		content = append(content, line)
	}
	content = append(content, "")

	t.PrintBox("📋 CONFIGURATION", content)
}

// Separators
func (t *UITemplate) PrintSeparator(char string, colorFunc func(...interface{}) string) {
	separator := strings.Repeat(char, t.width)
	fmt.Println(colorFunc(separator))
}

func (t *UITemplate) PrintThickSeparator() {
	t.PrintSeparator("━", Gold.Sprint)
}

func (t *UITemplate) PrintThinSeparator() {
	t.PrintSeparator("─", Gray.Sprint)
}

func (t *UITemplate) PrintStandardSeparator() {
	t.PrintSeparator("═", Gold.Sprint)
}
