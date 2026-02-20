# TamboUI Core Concepts

## Overview

TamboUI is a Java terminal UI framework (version 0.2.0-SNAPSHOT) that closely follows ratatui's architecture. The framework uses immediate-mode rendering with an intermediate buffer system for terminal applications.

## Rendering Model

TamboUI implements a three-stage rendering pipeline:

1. Widgets render to a Buffer (not directly to terminal)
2. Buffer is diffed (only changed cells sent to terminal)
3. Each frame is a full redraw (stateless widgets, simple state management)

This design ensures "the UI is always a pure function of your application state."

## Buffer System

### Buffer
A 2D grid of Cell objects representing the terminal screen:

```java
Buffer buffer = Buffer.empty(new Rect(0, 0, width, height));
buffer.set(x, y, new Cell("A", Style.EMPTY.fg(Color.RED)));
Cell cell = buffer.get(x, y);
```

### Cell
Represents a single character with styling:

```java
Cell cell = new Cell("X", Style.EMPTY.bold().fg(Color.CYAN));
String symbol = cell.symbol();
Style style = cell.style();
```

### Frame
Wraps a buffer and provides rendering helpers:

```java
terminal.draw(frame -> {
    Rect area = frame.area();
    frame.renderWidget(myWidget, area);

    Rect subArea = new Rect(0, 0, 40, 10);
    frame.renderWidget(headerWidget, subArea);
});
```

## Layout System

### Rect
Defines rectangular regions with position and size:

```java
Rect rect = new Rect(x, y, width, height);
int rx = rect.x();
int ry = rect.y();
int rw = rect.width();
int rh = rect.height();
int right = rect.right();    // x + width
int bottom = rect.bottom();  // y + height
```

### Constraints
Space allocation types:

| Constraint | Purpose | Example |
|-----------|---------|---------|
| Length(n) | Fixed size | `Constraint.length(20)` |
| Percentage(n) | Percentage of space | `Constraint.percentage(50)` |
| Ratio(num, denom) | Fractional size | `Constraint.ratio(1, 3)` |
| Min(n) | Minimum cells | `Constraint.min(10)` |
| Max(n) | Maximum cells | `Constraint.max(50)` |
| Fill(weight) | Remaining space | `Constraint.fill()` or `Constraint.fill(2)` |

### Layout
Divides rectangular areas into smaller sections using constraints:

```java
Layout layout = Layout.vertical()
    .constraints(
        Constraint.length(3),   // 3 cells tall
        Constraint.fill()       // remaining space
    );

Rect area = new Rect(0, 0, 80, 24);
List<Rect> areas = layout.split(area);
// areas.get(0) = Rect(0, 0, 80, 3)
// areas.get(1) = Rect(0, 3, 80, 21)
```

**Complete rendering example:**

```java
Layout layout = Layout.vertical()
    .constraints(
        Constraint.length(3),
        Constraint.fill()
    );

List<Rect> areas = layout.split(frame.area());
Rect headerArea = areas.get(0);
Rect contentArea = areas.get(1);

Paragraph header = Paragraph.from("Header Text");
header.render(headerArea, frame.buffer());

Paragraph content = Paragraph.from("Main content goes here...");
content.render(contentArea, frame.buffer());
```

**Nested layouts:**

```java
Layout horizontalLayout = Layout.horizontal()
    .constraints(
        Constraint.percentage(30),
        Constraint.percentage(70)
    );

List<Rect> columns = horizontalLayout.split(contentArea);

ListWidget sidebar = ListWidget.builder()
    .items(ListItem.from("Item 1"), ListItem.from("Item 2"))
    .build();
sidebar.render(columns.get(0), frame.buffer(), new ListState());

Paragraph mainArea = Paragraph.from("Main content");
mainArea.render(columns.get(1), frame.buffer());
```

#### Flex Positioning
Controls how remaining space is distributed when children don't fill containers:

```java
Layout toolbar = Layout.horizontal()
    .constraints(
        Constraint.length(10),  // Button 1
        Constraint.length(10),  // Button 2
        Constraint.length(10)   // Button 3
    )
    .flex(Flex.CENTER);  // Center buttons, space on both sides

List<Rect> buttonAreas = toolbar.split(toolbarArea);
renderButton("Save", buttonAreas.get(0), frame.buffer());
renderButton("Cancel", buttonAreas.get(1), frame.buffer());
renderButton("Help", buttonAreas.get(2), frame.buffer());
```

**Navigation bar with spacing:**

```java
Layout navbar = Layout.horizontal()
    .constraints(
        Constraint.length(8),   // "File"
        Constraint.length(8)    // "Edit"
    )
    .flex(Flex.SPACE_BETWEEN);  // Push to edges with gap

List<Rect> menuAreas = navbar.split(navbarArea);
renderMenuItem("File", menuAreas.get(0), frame.buffer());
renderMenuItem("Edit", menuAreas.get(1), frame.buffer());
```

Available flex modes: `START`, `CENTER`, `END`, `SPACE_BETWEEN`, `SPACE_AROUND`, `SPACE_EVENLY`.

### Direction
Layout orientation types:

- `Direction.VERTICAL` - stack top to bottom
- `Direction.HORIZONTAL` - place left to right

### Margin
Add spacing around layouts:

```java
Layout.vertical()
    .margin(new Margin(1, 2, 1, 2))  // top, right, bottom, left
    .constraints(Constraint.fill());
```

## Styling

### Style
Immutable object defining text appearance:

```java
Style style = Style.EMPTY
    .fg(Color.CYAN)
    .bg(Color.BLACK)
    .bold()
    .underlined();

Style dimStyle = style.dim();  // returns new instance
```

### Color
Multiple color specification methods:

```java
// Named ANSI colors
Color red = Color.RED;
Color green = Color.GREEN;
Color cyan = Color.CYAN;
Color white = Color.WHITE;
Color gray = Color.GRAY;

// Indexed colors (0-255)
Color indexed = Color.indexed(196);

// RGB colors (true color)
Color rgb = Color.rgb(255, 128, 0);
```

### Modifiers
Text display modifications:

```java
Style.EMPTY
    .bold()       // Bold text
    .dim()        // Dimmed/faint
    .italic()     // Italic (terminal support varies)
    .underlined() // Underlined
    .slowBlink()  // Slow blinking
    .rapidBlink() // Rapid blinking
    .reversed()   // Swap fg/bg colors
    .hidden()     // Hidden text
    .crossedOut(); // Strikethrough
```

## Text System

Hierarchical text model for styled terminal output: `Text` > `Line` > `Span`

### Text
Multi-line styled text representation:

```java
Text text = Text.from("Hello, World!");

Text multiLine = Text.from(
    Line.from("First line"),
    Line.from("Second line")
);

Text centered = Text.from("Centered")
    .alignment(Alignment.CENTER);
```

### Line
Single line composed of Span objects:

```java
Line line = Line.from(
    Span.styled("Bold", Style.EMPTY.bold()),
    Span.raw(" and "),
    Span.styled("Red", Style.EMPTY.fg(Color.RED))
);
```

### Span
Styled piece of text:

```java
Span plain = Span.raw("plain text");

Span styled = Span.styled("styled",
    Style.EMPTY.fg(Color.CYAN).bold());
```

## Widget Interfaces

### Widget (Stateless)
Implements the Widget interface:

```java
public interface WidgetInterface {
    void render(Rect area, Buffer buffer);
}
```

Examples: Paragraph, Gauge, Block, Clear

### StatefulWidget (Stateful)
Carries external state:

```java
public interface StatefulWidgetInterface<S> {
    void render(Rect area, Buffer buffer, S state);
}
```

Usage example:

```java
ListState listState = new ListState();
listWidget.render(area, buffer, listState);
listState.selectNext(items.size());
```

Examples: ListWidget, Table, Tabs, TextInput

## Event System

### KeyEvent
Keyboard input handling:

```java
if (event.code() == KeyCode.ENTER) { /* handle enter */ }
if (event.code() == KeyCode.CHAR &&
    event.character() == 'q') { /* handle q */ }

if (event.modifiers().ctrl()) { /* Ctrl held */ }

if (event.isQuit()) { /* quit action */ }
if (event.isUp()) { /* move up action */ }
```

### MouseEvent
Mouse input handling:

```java
int x = event.x();
int y = event.y();
MouseEventKind kind = event.kind();

if (kind == MouseEventKind.PRESS &&
    event.button() == MouseButton.LEFT) {
    handleClick(x, y);
}
```

### TickEvent
Periodic events for animations:

```java
TuiConfig config = TuiConfig.builder()
    .tickRate(Duration.ofMillis(16))  // ~60fps
    .build();

if (event instanceof TickEvent) {
    updateAnimation();
    // return true to trigger redraw
}
```

### Exception Hierarchy
Clear exception structure for framework errors:

- `TamboUIException` - base framework exception (RuntimeException)
- `RuntimeIOException` - terminal I/O errors
- `BackendException` - backend non-I/O errors
- `TuiException` - TUI framework errors
- `IllegalArgumentException`, `IllegalStateException` - invalid parameters/state
- Domain-specific: `SolverException`, `CssParseException`, etc.

The `TuiRunner` provides centralized error handling with configurable `RenderErrorHandler`. Default behavior displays errors (with stack traces) in the UI for debugging.

## Next Steps

- Explore the Widgets Reference for available components
- Learn about API Levels in detail
- Use CSS Styling for external style definitions
- Understand Bindings and Actions for input handling
- Build maintainable apps with Application Structure (MVC)
