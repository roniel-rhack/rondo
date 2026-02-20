# TamboUI Markup Text Documentation

## Overview

"TamboUI provides a BBCode-style markup parser for creating styled Text objects from strings."

## Basic Usage

```java
// Parse markup into styled Text
Text text = MarkupParser.parse("This is [bold]bold[/bold] and [red]red[/red].");

// Use with widgets
Paragraph paragraph = Paragraph.from(
    MarkupParser.parse("[cyan]Status:[/cyan] [bold green]Ready[/]")
);
```

## Style Tags

### Modifier Tags

| Tag | Effect | Aliases |
|-----|--------|---------|
| `[bold]` | Bold text | `[b]` |
| `[italic]` | Italic text | `[i]` |
| `[underlined]` | Underlined text | `[u]` |
| `[dim]` | Dimmed/faint text | -- |
| `[reversed]` | Swap foreground and background | -- |
| `[crossed-out]` | Strikethrough text | `[strikethrough]`, `[s]` |

### Color Support

"Colors can be specified the ways ColorConverter supports: Named colors, hex colors, RGB colors, and indexed colors."

Supported formats:
- Named: `red`, `green`, `blue`, `yellow`, `cyan`, `magenta`, `white`, `black`, `gray`
- Hex: `#RGB`, `#RRGGBB`
- RGB: `rgb(r,g,b)`
- Indexed: `indexed(index)`

## Tag Closure

### Implicit Close
"Close the most recent tag without naming it" using `[/]`:

```java
[bold]Hello[/] World
```

### Compound Styles

Combine multiple styles in one tag:

```java
[bold red]Error message[/]
[italic cyan]Hint text[/]
[bold underlined yellow]Warning![/]
```

## Background Colors

"Use on to specify a background color":

```java
[white on blue]Highlighted text[/]
[black on yellow]Warning banner[/]
[bold red on white]Alert![/]
```

## Hyperlinks

```java
[link=https://example.com]Click here[/link]
```

## Escape Mechanisms

### Double Brackets
```
[[tag]] renders as: [tag]
]] renders as: ]
```

### Backslash Escapes
```
\[tag\] renders as: [tag]
\\ renders as: \
```

"Supported backslash escapes: `\[`, `\]`, `\\`. Other sequences (like `\n`) are preserved as-is."

## Nesting

```java
[red]This is red and [bold]this is red and bold[/bold][/red]
```

## Custom Style Resolver

Define custom tag mappings:

```java
MarkupParser.StyleResolver resolver = tagName -> {
    switch (tagName) {
        case "keyword": return Style.EMPTY.fg(Color.CYAN).bold();
        case "string": return Style.EMPTY.fg(Color.GREEN);
        case "comment": return Style.EMPTY.fg(Color.GRAY).italic();
        case "error": return Style.EMPTY.fg(Color.WHITE).bg(Color.RED).bold();
        default: return null;
    }
};

Text code = MarkupParser.parse(
    "[keyword]public[/] [keyword]void[/] main([string]\"Hello\"[/])",
    resolver
);
```

### Resolver Priority

"The resolver has priority over built-in styles. This means you can redefine what built-in color names mean."

### Merging Styles

When combining compound styles with a resolver, "inline styles from the compound override the resolver's base style, similar to how CSS inline styles override class styles."

## CSS Class Targeting

"All tokens in a tag become CSS classes that can be targeted with stylesheets."

```java
Text text = MarkupParser.parse("[error]message[/]");
// Span has Tags extension containing "error"

Text compound = MarkupParser.parse("[bold underlined yellow]Warning![/]");
// Tags: "bold", "underlined", "yellow"

Text background = MarkupParser.parse("[white on blue]text[/]");
// Tags: "white", "blue" (not "on")
```

Example CSS:
```css
.bold { text-style: bold; }
.yellow { color: #ffcc00; }
.error { color: red; text-style: bold; }
```

## Unknown Tags

"Tags that aren't recognized as built-in styles and aren't resolved by a custom resolver are still tracked for CSS class targeting, but have no visual effect."

## Multi-line Text

```java
Text multiline = MarkupParser.parse("""
    [bold]Header[/]

    [dim]This is a paragraph with [cyan]highlighted[/] text.[/]

    [italic]Footer note[/]
    """);
```

## Examples

### Status Messages
```java
MarkupParser.parse("[green]SUCCESS[/] Operation completed");
MarkupParser.parse("[yellow]WARNING[/] Disk space low");
MarkupParser.parse("[bold red]ERROR[/] Connection failed");
```

### Syntax Highlighting
```java
MarkupParser.StyleResolver syntax = tag -> switch (tag) {
    case "kw" -> Style.EMPTY.fg(Color.MAGENTA).bold();
    case "str" -> Style.EMPTY.fg(Color.GREEN);
    case "num" -> Style.EMPTY.fg(Color.CYAN);
    case "cmt" -> Style.EMPTY.fg(Color.GRAY).italic();
    default -> null;
};

Text code = MarkupParser.parse(
    "[kw]int[/] x = [num]42[/]; [cmt]// answer[/]",
    syntax
);
```

### Rich Notifications
```java
MarkupParser.parse("""
    [bold white on blue] NOTICE [/]

    Your session will expire in [bold yellow]5 minutes[/].
    Please [link=https://example.com/save]save your work[/link].
    """);
```

## Syntax Summary

| Syntax | Description | Example |
|--------|-------------|---------|
| `[tag]...[/tag]` | Explicit close | `[bold]text[/bold]` |
| `[tag]...[/]` | Implicit close | `[bold]text[/]` |
| `[tag1 tag2]...[/]` | Compound styles | `[bold red]text[/]` |
| `[color on bgcolor]` | Background color | `[white on blue]text[/]` |
| `[link=url]...[/link]` | Hyperlink | `[link=https://x.com]click[/link]` |
| `[[` / `]]` | Double brackets | `Use [[tag]]` |
| `\[` / `\]` / `\\` | Backslash escapes | `\[tag\]` |
