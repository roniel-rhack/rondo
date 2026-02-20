# TamboUI Developer Guide

## Overview

Version 0.2.0-SNAPSHOT

This guide covers creating custom widgets and components for TamboUI, a terminal UI framework.

---

## Creating Custom Widgets

TamboUI provides two widget interfaces based on state requirements.

### Stateless Widgets

For widgets producing consistent output from identical inputs:

```java
public interface WidgetInterface {
    void render(Rect area, Buffer buffer);
}
```

**Example: Separator Widget**

```java
public class Separator implements Widget {
    private final Style style;

    public Separator(Style style) {
        this.style = style;
    }

    @Override
    public void render(Rect area, Buffer buffer) {
        for (int x = area.x(); x < area.right(); x++) {
            buffer.set(x, area.y(), new Cell("\u2500", style));
        }
    }
}
```

### Stateful Widgets

For widgets requiring selection, scroll position, or other state tracking:

```java
public interface StatefulWidgetInterface<S> {
    void render(Rect area, Buffer buffer, S state);
}
```

**Example: Counter Widget**

```java
public class Counter implements StatefulWidget<Counter.State> {
    private final Style style;

    public static class State {
        private int value = 0;

        public int value() { return value; }
        public void increment() { value++; }
        public void decrement() { if (value > 0) value--; }
    }

    public Counter(Style style) {
        this.style = style;
    }

    @Override
    public void render(Rect area, Buffer buffer, State state) {
        String text = "Count: " + state.value();
        int x = area.x();
        for (char c : text.toCharArray()) {
            if (x < area.right()) {
                buffer.set(x++, area.y(), new Cell(String.valueOf(c), style));
            }
        }
    }
}
```

**Usage:**

```java
Counter.State counterState = new Counter.State();
Counter counter = new Counter(Style.EMPTY.bold());
counter.render(area, buffer, counterState);
```

---

## Buffer Operations

### Setting Cells

```java
// Single cell
buffer.set(x, y, new Cell("X", style));

// String
int col = x;
for (char c : "Hello".toCharArray()) {
    buffer.set(col++, y, new Cell(String.valueOf(c), style));
}
```

### Reading Cells

```java
Cell cell = buffer.get(x, y);
String symbol = cell.symbol();
Style cellStyle = cell.style();
```

### Filling Areas

```java
for (int row = area.y(); row < area.bottom(); row++) {
    for (int col = area.x(); col < area.right(); col++) {
        buffer.set(col, row, new Cell(" ", bgStyle));
    }
}
```

### Bounds Checking

Always verify coordinates before writing:

```java
if (x >= area.x() && x < area.right() &&
    y >= area.y() && y < area.bottom()) {
    buffer.set(x, y, cell);
}
```

---

## Exporting Buffer Content

Export buffers to SVG, HTML, or text formats for documentation or sharing.

### Fluent API

```java
Buffer buffer = Buffer.empty(new Rect(0, 0, 80, 24));

// SVG export shorthand
export(buffer).svg().toFile(Path.of("output.svg"));

// Configure options
export(buffer).svg()
    .options(o -> o.title("My App"))
    .toFile(Path.of("output.svg"));

// Export to string or bytes
String svgString = export(buffer).svg().toString();
byte[] htmlBytes = export(buffer).html()
    .options(o -> o.inlineStyles(true))
    .toBytes();

// Write to stream or writer
export(buffer).text().to(outputStream);
```

### Format Selection

Built-in formats:

- `export(buffer).svg()` -- SVG graphics
- `export(buffer).html()` -- HTML document
- `export(buffer).text()` -- Plain or ANSI text

Custom formats:

```java
export(buffer).as(Formats.SVG).toFile(path);
```

File extension mapping:
- `.svg` -> SVG
- `.html` / `.htm` -> HTML
- `.txt` / `.asc` -> plain text
- `.ans` / `.ansi` -> ANSI text
- Unknown -> defaults to SVG

### Output Methods

After format selection:

- `toFile(Path path)` -- write file (UTF-8)
- `to(OutputStream out)` -- write stream (UTF-8)
- `to(Writer out)` -- write using caller's charset
- `toString()` -- return string
- `toBytes()` -- return UTF-8 bytes

### Cropping to Region

```java
Rect titleRect = new Rect(0, 0, 80, 3);
export(buffer).crop(titleRect).svg().toFile(Path.of("title.svg"));

// Union of multiple rectangles
Rect tableRect = new Rect(22, 3, 58, 12);
Rect footerRect = new Rect(0, 15, 80, 2);
Rect combined = tableRect.union(footerRect);
export(buffer).crop(combined).svg().toFile(Path.of("combined.svg"));
```

**Toolkit Integration:**

```java
Rect area = myPanel.renderedArea();
if (area != null && !area.isEmpty()) {
    export(buffer).crop(area).svg().toFile(Path.of("panel.svg"));
}
```

### Default Export Colors

Export formats use default foreground and background when cell styles lack color specification.

```java
// Use property defaults
export(buffer).svg().toFile(path);

// Custom defaults via resolver
StylePropertyResolver resolver = StylePropertyResolver.empty();
export(buffer).svg()
    .options(o -> o.styles(resolver))
    .toFile(path);
```

### SVG Export

Options via `.options(o -> ...)`:

- `title(String)` -- window title (default: "TamboUI")
- `chrome(boolean)` -- include frame/buttons (default: true)
- `styles(StylePropertyResolver)` -- style resolver
- `fontAspectRatio(double)` -- width-to-height ratio (default: 0.61)
- `uniqueId(String)` -- CSS class prefix or null for auto

```java
export(buffer).svg()
    .options(o -> o.title("Dashboard"))
    .toFile(Path.of("dashboard.svg"));
```

### HTML Export

Options:

- `styles(StylePropertyResolver)` -- style resolver
- `inlineStyles(boolean)` -- inline styles on spans (default: false)

```java
// External stylesheet (smaller)
export(buffer).html().toFile(Path.of("output.html"));

// Inline styles (self-contained)
export(buffer).html()
    .options(o -> o.inlineStyles(true))
    .toFile(Path.of("output_inline.html"));
```

### Text Export

Options:

- `styles(boolean)` -- include ANSI codes (default: false)

```java
// Plain text
String plain = export(buffer).text().toString();
export(buffer).text().toFile(Path.of("dump.txt"));

// ANSI-styled
String ansi = export(buffer).text()
    .options(o -> o.styles(true))
    .toString();
export(buffer).text()
    .options(o -> o.styles(true))
    .toFile(Path.of("styled.ansi"));
```

---

## Creating Toolkit Components

Extend `Component` class for DSL integration. Use `@OnAction` annotation for input handling:

```java
public class CounterCard extends Component<CounterCard> {
    private int count = 0;

    @OnAction("increment")
    void onIncrement(Event event) {
        count++;
    }

    @OnAction("decrement")
    void onDecrement(Event event) {
        count--;
    }

    @Override
    protected Element render() {
        var borderColor = isFocused() ? Color.CYAN : Color.GRAY;

        return panel(() -> column(
                text("Count: " + count).bold(),
                text("Press +/- to change").dim()
        ))
        .rounded()
        .borderColor(borderColor)
        .fill();
    }
}
```

### Using Components

Components require IDs:

```java
var counter1 = new CounterCard().id("counter-1");
var counter2 = new CounterCard().id("counter-2");

var bindings = BindingSets.standard()
    .toBuilder()
    .bind(KeyTrigger.ch('+'), "increment")
    .bind(KeyTrigger.ch('='), "increment")
    .bind(KeyTrigger.ch('-'), "decrement")
    .bind(KeyTrigger.ch('_'), "decrement")
    .build();

try (var runner = ToolkitRunner.builder()
        .bindings(bindings)
        .build()) {
    runner.run(() -> row(counter1, counter2));
}
```

---

## CSS Property Support

Widgets support CSS styling through `StylePropertyResolver` and `PropertyDefinition`.

### Defining Custom Properties

```java
public class MyGauge implements Widget {
    public static final PropertyDefinition<Color> BAR_COLOR =
        PropertyDefinition.of("bar-color", ColorConverter.INSTANCE);

    static {
        PropertyRegistry.register(BAR_COLOR);
    }

    @Override
    public void render(Rect area, Buffer buffer) {
        // Implementation
    }
}
```

Multiple properties:

```java
PropertyRegistry.registerAll(PROP1, PROP2);
```

Registration enables:
- Style resolver validation
- Property converter availability
- CSS styling support

Example CSS:

```css
MyGauge {
    bar-color: green;
}
```

### Using StylePropertyResolver

```java
public static class MyGaugeWithResolver implements Widget {
    public static final PropertyDefinition<Color> BAR_COLOR =
        PropertyDefinition.of("bar-color", ColorConverter.INSTANCE);

    private final Color barColor;

    private MyGaugeWithResolver(Builder builder) {
        this.barColor = builder.resolveBarColor();
    }

    @Override
    public void render(Rect area, Buffer buffer) {
        // Use barColor
    }

    public static final class Builder {
        private Color barColor;
        private StylePropertyResolver styleResolver = StylePropertyResolver.empty();

        public Builder barColor(Color color) {
            this.barColor = color;
            return this;
        }

        public Builder styleResolver(StylePropertyResolver resolver) {
            this.styleResolver = resolver != null ? resolver : StylePropertyResolver.empty();
            return this;
        }

        private Color resolveBarColor() {
            return styleResolver.resolve(BAR_COLOR, barColor);
        }

        public MyGaugeWithResolver build() {
            return new MyGaugeWithResolver(this);
        }
    }
}
```

**Resolution order:**

1. Programmatic value (builder method)
2. CSS value (style resolver)
3. Property default

### Using Standard Properties

```java
Color bg = styleResolver.resolve(StandardProperties.BACKGROUND, background);
Color fg = styleResolver.resolve(StandardProperties.COLOR, foreground);
```

### Inheritable vs Non-Inheritable

**Inheritable** properties propagate to children:

```css
Panel {
    color: cyan;  /* All Text children inherit */
}
```

**Non-inheritable** properties apply to target only:

```css
Panel {
    background: gray;  /* Only Panel, not children */
}
```

Define as:

- Non-inheritable: `PropertyDefinition.of(...)` (default)
- Inheritable: `PropertyDefinition.builder(...).inheritable().build()`

Most widget-specific properties should be non-inheritable.

---

## Unicode and Display Width Handling

Terminal display width differs from Java string length. Use `CharWidth` for all text width calculations.

### The Problem

| Type | Example | `length()` | Display Width |
|------|---------|-----------|---------------|
| ASCII | "A" | 1 | 1 |
| CJK | "world" | 1 | 2 |
| Simple Emoji | "fire" | 2 | 2 |
| ZWJ Emoji | "man-bald" | 5 | 2 |
| Flag Emoji | "flag-GL" | 4 | 2 |

Using `length()` causes:
- ZWJ emoji truncation mid-sequence
- CJK character overflow
- Incorrect cursor positioning

### CharWidth Utilities

```java
// Display width of string
int width = CharWidth.of("Hello world fire");  // Returns 14, not 11

// Display width of code point
int cpWidth = CharWidth.of(0x4E16);  // Returns 2 (CJK)

// Truncate to fit width (preserves graphemes)
String truncated = CharWidth.substringByWidth("Hello man-bald World", 8);

// Truncate from end
String suffix = CharWidth.substringByWidthFromEnd("Hello World", 5);

// Truncate with ellipsis
String ellipsized = CharWidth.truncateWithEllipsis(
    "Very long text here",
    10,
    CharWidth.TruncatePosition.END
);
```

### Common Patterns

**Width Calculation:**

```java
// WRONG - breaks emoji and CJK
int wrongWidth = text.length();

// CORRECT
int correctWidth = CharWidth.of(text);
```

**Text Truncation:**

```java
// WRONG - may break mid-grapheme
String wrongTruncated = text.substring(0, Math.min(text.length(), maxWidth));

// CORRECT
String correctTruncated = CharWidth.substringByWidth(text, maxWidth);
```

**Position Tracking:**

```java
// WRONG - position drift
int col = x;
buffer.setString(col, y, text, style);
col += text.length();

// CORRECT
int col2 = x;
buffer.setString(col2, y, text, style);
col2 += CharWidth.of(text);
```

**Centering Text:**

```java
// WRONG - misaligned
int wrongLabelX = x + (width - label.length()) / 2;

// CORRECT
int correctLabelX = x + (width - CharWidth.of(label)) / 2;
```

**Truncation with Ellipsis:**

```java
// WRONG
String wrongEllipsis = text;
if (text.length() > maxWidth) {
    wrongEllipsis = text.substring(0, maxWidth - 3) + "...";
}

// CORRECT
String correctEllipsis = CharWidth.truncateWithEllipsis(
    text, maxWidth, CharWidth.TruncatePosition.END
);

// Manual approach
String manualEllipsis = text;
if (CharWidth.of(text) > maxWidth) {
    int ellipsisWidth = CharWidth.of("...");
    manualEllipsis = CharWidth.substringByWidth(text, maxWidth - ellipsisWidth) + "...";
}
```

### Ellipsis Truncation Positions

```java
String text = "Hello World Example";

// END: "Hello Wor..."
String endTruncate = CharWidth.truncateWithEllipsis(
    text, 12, CharWidth.TruncatePosition.END
);

// START: "...d Example"
String startTruncate = CharWidth.truncateWithEllipsis(
    text, 12, CharWidth.TruncatePosition.START
);

// MIDDLE: "Hell...ample"
String middleTruncate = CharWidth.truncateWithEllipsis(
    text, 12, CharWidth.TruncatePosition.MIDDLE
);

// Custom ellipsis
String customEllipsis = CharWidth.truncateWithEllipsis(
    text, 12, "\u2026", CharWidth.TruncatePosition.END
);
```

### ZWJ Sequence Safety

Zero-Width Joiner sequences combine multiple code points into single glyphs. `CharWidth.substringByWidth()` preserves these:

```java
// "man-bald" = man + ZWJ + bald = 5 code units, 2 display columns

// Safe truncation - won't break mid-sequence
String result = CharWidth.substringByWidth("A man-bald B", 3);
// Returns "A man-bald" (width 3), not broken sequence
```

### Reference Implementation

See `Paragraph.java` for complete CharWidth usage example with text wrapping and truncation.

---

## Best Practices

### Widget Design

- Maintain single responsibility focus
- Implement builder pattern for complex configuration
- Respect provided `Rect` bounds
- Handle edge cases (empty area, missing data)
- Always use `CharWidth` for text calculations

### Performance

- Minimize allocations in `render()` methods
- Pre-compute strings and styles when possible
- Use primitive arrays over collections for large data

### State Management

- Keep state classes simple
- Provide methods for all modifications
- Consider immutable state for thread safety

---

## Next Steps

- Review Widgets Reference for existing implementations
- Learn Bindings and Actions for input handling
- Study Application Structure for organizing larger applications

---

## Further Reading

Source code locations:

- `tamboui-widgets/src/main/java/dev/tamboui/widgets/` -- Widget implementations
- `tamboui-toolkit/src/main/java/dev/tamboui/toolkit/` -- Toolkit components
- `tamboui-core/src/main/java/dev/tamboui/buffer/` -- Buffer and Cell
- `tamboui-core/src/main/java/dev/tamboui/export/` -- Export API
- `tamboui-core/src/main/java/dev/tamboui/text/CharWidth.java` -- Unicode utilities
- `tamboui-widgets/src/main/java/dev/tamboui/widgets/paragraph/Paragraph.java` -- Reference implementation
