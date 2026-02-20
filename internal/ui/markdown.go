package ui

import (
	"bufio"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	boldRe   = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRe = regexp.MustCompile(`\*(.+?)\*`)
	codeRe   = regexp.MustCompile("`(.+?)`")
)

// RenderMarkdown renders markdown-formatted text with lipgloss styles for
// display in a dark terminal with cyan accent theme.
func RenderMarkdown(s string, width int) string {
	if width <= 0 {
		width = 80
	}

	h1Style := lipgloss.NewStyle().Bold(true).Foreground(Cyan)
	h2Style := lipgloss.NewStyle().Bold(true).Foreground(White)
	bulletStyle := lipgloss.NewStyle().Foreground(Cyan)
	quoteStyle := lipgloss.NewStyle().Foreground(DimGray).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(DimGray).
		PaddingLeft(1)
	defaultStyle := lipgloss.NewStyle().Foreground(White)
	boldStyle := lipgloss.NewStyle().Bold(true)
	italicStyle := lipgloss.NewStyle().Italic(true)
	codeStyle := lipgloss.NewStyle().Background(DimGray)

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "## "):
			text := strings.TrimPrefix(line, "## ")
			text = wrapText(text, width)
			lines = append(lines, h2Style.Render(text))

		case strings.HasPrefix(line, "# "):
			text := strings.TrimPrefix(line, "# ")
			text = wrapText(text, width)
			lines = append(lines, h1Style.Render(text))

		case strings.HasPrefix(line, "> "):
			text := strings.TrimPrefix(line, "> ")
			text = wrapText(text, width-4)
			lines = append(lines, quoteStyle.Render(text))

		case strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* "):
			text := line[2:]
			text = renderInlineStyles(text, boldStyle, italicStyle, codeStyle)
			text = wrapText(text, width-4)
			lines = append(lines, bulletStyle.Render("  * ")+defaultStyle.Render(text))

		case strings.TrimSpace(line) == "":
			lines = append(lines, "")

		default:
			text := renderInlineStyles(line, boldStyle, italicStyle, codeStyle)
			text = wrapText(text, width)
			lines = append(lines, defaultStyle.Render(text))
		}
	}

	return strings.Join(lines, "\n")
}

// renderInlineStyles applies bold, italic, and code inline formatting.
func renderInlineStyles(s string, boldStyle, italicStyle, codeStyle lipgloss.Style) string {
	// Process code first to avoid conflicts with bold/italic inside code spans.
	s = codeRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[1 : len(match)-1]
		return codeStyle.Render(inner)
	})

	s = boldRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[2 : len(match)-2]
		return boldStyle.Render(inner)
	})

	// Italic: only match single asterisks that are not part of bold.
	s = italicRe.ReplaceAllStringFunc(s, func(match string) string {
		inner := match[1 : len(match)-1]
		return italicStyle.Render(inner)
	})

	return s
}

// wrapText wraps text at the given width, breaking on spaces.
func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	// Use lipgloss width-aware measurement for ANSI strings.
	if lipgloss.Width(s) <= width {
		return s
	}

	var result strings.Builder
	words := strings.Fields(s)
	lineLen := 0

	for i, word := range words {
		wLen := lipgloss.Width(word)
		if i > 0 && lineLen+1+wLen > width {
			result.WriteByte('\n')
			lineLen = 0
		} else if i > 0 {
			result.WriteByte(' ')
			lineLen++
		}
		result.WriteString(word)
		lineLen += wLen
	}

	return result.String()
}
