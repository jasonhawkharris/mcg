package tui

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Regular expression to find Markdown code blocks with language specification
	// Example: ```go\npackage main\n```
	codeBlockRegexp = regexp.MustCompile("```([a-zA-Z0-9_+-]+)?\n([\\s\\S]*?)\n```")
	
	// Regular expression to find inline code
	// Example: `fmt.Println("Hello")`
	inlineCodeRegexp = regexp.MustCompile("`([^`\n]+)`")
)

// Highlight applies syntax highlighting to Markdown content
func Highlight(content string) string {
	// First, handle code blocks (multiline code with language specification)
	result := codeBlockRegexp.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the language and code content
		submatches := codeBlockRegexp.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match // Unexpected match format, return as is
		}
		
		lang := submatches[1]
		code := submatches[2]
		
		return formatCodeBlock(code, lang)
	})
	
	// Then, handle inline code
	result = inlineCodeRegexp.ReplaceAllStringFunc(result, func(match string) string {
		// Extract the code content (without the backticks)
		code := match[1 : len(match)-1]
		return formatInlineCode(code)
	})
	
	return result
}

// formatCodeBlock formats a code block with syntax highlighting
func formatCodeBlock(code, lang string) string {
	// Trim any trailing whitespace that might have been included in the match
	code = strings.TrimSpace(code)
	
	var lexer chroma.Lexer
	if lang != "" {
		// Get lexer for specified language
		lexer = lexers.Get(lang)
	}
	
	if lexer == nil {
		// If no language specified or language not found, try to guess
		lexer = lexers.Analyse(code)
		if lexer == nil {
			// Fall back to plain text
			lexer = lexers.Get("plaintext")
		}
	}
	
	// Use a style suitable for terminal output
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}
	
	// Use terminal formatter
	formatter := formatters.Get("terminal")
	if formatter == nil {
		formatter = formatters.Fallback
	}
	
	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// Fall back to plain text if tokenization fails
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#2d2d2d")).
			Foreground(lipgloss.Color("#f8f8f2")).
			Padding(1).
			Render(code)
	}
	
	// Create a buffer for the formatter output
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		// Fall back to unstyled code if formatting fails
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#2d2d2d")).
			Foreground(lipgloss.Color("#f8f8f2")).
			Padding(1).
			Render(code)
	}

	// Add a border and padding around the code block
	return "\n" + lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#696969")).
		Padding(1).
		Render(buf.String()) + "\n"
}

// formatInlineCode formats inline code
func formatInlineCode(code string) string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color("#2d2d2d")).
		Foreground(lipgloss.Color("#f8f8f2")).
		Padding(0, 1, 0, 1).  // Padding on left and right only
		Render(code)
}