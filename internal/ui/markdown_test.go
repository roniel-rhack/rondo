package ui

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_H1(t *testing.T) {
	result := RenderMarkdown("# Hello World", 80)
	if !strings.Contains(result, "Hello World") {
		t.Errorf("expected heading text in output, got: %q", result)
	}
	// Should not contain the markdown prefix.
	if strings.Contains(result, "# ") {
		t.Errorf("expected markdown prefix to be stripped, got: %q", result)
	}
}

func TestRenderMarkdown_H2(t *testing.T) {
	result := RenderMarkdown("## Subheading", 80)
	if !strings.Contains(result, "Subheading") {
		t.Errorf("expected subheading text in output, got: %q", result)
	}
	if strings.Contains(result, "## ") {
		t.Errorf("expected markdown prefix to be stripped, got: %q", result)
	}
}

func TestRenderMarkdown_BoldText(t *testing.T) {
	result := RenderMarkdown("This is **bold** text", 80)
	if !strings.Contains(result, "bold") {
		t.Errorf("expected bold text in output, got: %q", result)
	}
	// The ** markers should be removed.
	if strings.Contains(result, "**") {
		t.Errorf("expected ** markers to be removed, got: %q", result)
	}
}

func TestRenderMarkdown_ItalicText(t *testing.T) {
	result := RenderMarkdown("This is *italic* text", 80)
	if !strings.Contains(result, "italic") {
		t.Errorf("expected italic text in output, got: %q", result)
	}
}

func TestRenderMarkdown_BulletList(t *testing.T) {
	input := "- First item\n- Second item\n* Third item"
	result := RenderMarkdown(input, 80)
	if !strings.Contains(result, "First item") {
		t.Errorf("expected first item in output, got: %q", result)
	}
	if !strings.Contains(result, "Second item") {
		t.Errorf("expected second item in output, got: %q", result)
	}
	if !strings.Contains(result, "Third item") {
		t.Errorf("expected third item in output, got: %q", result)
	}
	// Bullets should be rendered with * character.
	if !strings.Contains(result, "*") {
		t.Errorf("expected bullet markers in output, got: %q", result)
	}
}

func TestRenderMarkdown_Blockquote(t *testing.T) {
	result := RenderMarkdown("> This is a quote", 80)
	if !strings.Contains(result, "This is a quote") {
		t.Errorf("expected quote text in output, got: %q", result)
	}
}

func TestRenderMarkdown_CodeInline(t *testing.T) {
	result := RenderMarkdown("Use `fmt.Println` here", 80)
	if !strings.Contains(result, "fmt.Println") {
		t.Errorf("expected code text in output, got: %q", result)
	}
	// Backticks should be removed.
	if strings.Contains(result, "`") {
		t.Errorf("expected backticks to be removed, got: %q", result)
	}
}

func TestRenderMarkdown_EmptyInput(t *testing.T) {
	result := RenderMarkdown("", 80)
	if result != "" {
		t.Errorf("expected empty output for empty input, got: %q", result)
	}
}

func TestRenderMarkdown_PlainText(t *testing.T) {
	result := RenderMarkdown("Just plain text", 80)
	if !strings.Contains(result, "Just plain text") {
		t.Errorf("expected plain text in output, got: %q", result)
	}
}

func TestRenderMarkdown_MultipleLines(t *testing.T) {
	input := "# Title\n\nSome text\n\n- Item 1\n- Item 2"
	result := RenderMarkdown(input, 80)
	lines := strings.Split(result, "\n")
	if len(lines) < 4 {
		t.Errorf("expected at least 4 lines, got %d", len(lines))
	}
}
