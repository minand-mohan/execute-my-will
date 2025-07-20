package ui

import (
	"github.com/fatih/color"
)

// Color definitions for the medieval knight theme
var (
	// Primary colors
	Gold   = color.New(color.FgYellow, color.Bold)     // Titles, success, knight messages
	Blue   = color.New(color.FgBlue)                   // Information, status updates
	Red    = color.New(color.FgRed, color.Bold)       // Errors, failures
	Green  = color.New(color.FgGreen)                  // Success messages, confirmations
	Purple = color.New(color.FgMagenta)               // AI consultation, mystical elements
	White  = color.New(color.FgWhite)                 // Default text
	Cyan   = color.New(color.FgCyan)                  // Commands, code blocks
	Yellow = color.New(color.FgYellow)                // Warnings

	// Background colors for emphasis
	RedBg    = color.New(color.BgRed, color.FgWhite, color.Bold)
	GreenBg  = color.New(color.BgGreen, color.FgBlack, color.Bold)
	YellowBg = color.New(color.BgYellow, color.FgBlack, color.Bold)
	BlueBg   = color.New(color.BgBlue, color.FgWhite, color.Bold)

	// Faint colors for less important text
	Faint = color.New(color.Faint)
	Gray  = color.New(color.FgHiBlack)
	
	// Italic text
	Italic = color.New(color.Italic)
)

// Themed color functions
func KnightMessage(text string) string {
	return Gold.Sprint(text)
}

func SuccessMessage(text string) string {
	return Green.Sprint(text)
}

func ErrorMessage(text string) string {
	return Red.Sprint(text)
}

func WarningMessage(text string) string {
	return Yellow.Sprint(text)
}

func InfoMessage(text string) string {
	return Blue.Sprint(text)
}

func AIMessage(text string) string {
	return Purple.Sprint(text)
}

func CommandText(text string) string {
	return Cyan.Sprint(text)
}

func TimestampText(text string) string {
	return Gray.Sprint(text)
}

func HighlightText(text string) string {
	return Gold.Sprint(text)
}

func CommentText(text string) string {
	return Purple.Add(color.Italic).Sprint(text)
}