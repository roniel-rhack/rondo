# TamboUI Widgets Reference

## Overview

TamboUI version 0.2.0-SNAPSHOT offers "a comprehensive set of widgets for building terminal UIs." The library organizes components across multiple categories based on their function and state management requirements.

## Widget Categories

### Container Widgets

**Block**
A fundamental container supporting borders, titles, and padding. Configuration includes:
- Border types: PLAIN, ROUNDED, DOUBLE, THICK
- Padding specification: `new Padding(top, right, bottom, left)`
- Inner area calculation after borders and padding applied

**Clear**
"Clears an area by filling it with spaces. Useful before rendering overlays." Creates blank space for layering widgets.

### Text Widgets

**Paragraph**
Multi-line text display with wrapping modes (NONE, CHARACTER, WORD) and alignment options. Supports overflow handling and optional block containers.

### Selection Widgets

**ListWidget**
A stateful scrollable list requiring external `ListState` management. Features include:
- Selection navigation (next, previous, first, last)
- Highlight symbols and custom styling
- Item-based rendering

**ListElement (Toolkit DSL)**
Higher-level alternative managing state internally, accepting any `StyledElement` as items rather than just text.

**Table**
Grid structure supporting rows, columns, and selection. Configuration involves:
- Header row definition
- Width constraints per column
- Highlight styling for selected rows
- External `TableState` for selection tracking

**Tabs**
Tab navigation bar with titles, highlight styling, and divider customization. Uses `TabsState` for current selection.

**Tree**
Hierarchical view with keyboard navigation, expand/collapse functionality, and lazy loading support. Key features:

- Model/View separation via `TreeNode<T>` and custom `nodeRenderer`
- Navigation: arrow keys for movement, Enter/Space for toggling
- Lazy loading through `childrenLoader()`
- CSS styling with selectors: `TreeElement`, `TreeElement-node`, `TreeElement-node:selected`, `TreeElement-guide`

### Data Visualization Widgets

**Gauge**
Progress bar displaying completion percentage with label and custom styling.

**LineGauge**
"Compact single-line progress indicator" with filled/unfilled style configuration and line set options.

**Sparkline**
"Mini chart showing data trends" from integer array data with color customization.

**BarChart**
Vertical bar visualization supporting multiple bars, width/gap configuration, and custom coloring.

**Chart**
Line and scatter plots with:
- Dataset support with custom markers (DOT, BRAILLE, BLOCK)
- X/Y axis configuration with titles and bounds
- Multiple data series rendering

**Canvas**
Drawing surface supporting shapes:
- Circle (center, radius, color)
- Line (endpoints, color)
- Rectangle (position, size, color)
- Points (collection of coordinates)

**Calendar**
Monthly calendar widget with event styling via `CalendarEventStore`, customizable headers, and date range display options.

### Animated Widgets

**Spinner**
Loading indicator with built-in styles and custom frame support:
- Styles: DOTS, LINE, ARC, CIRCLE, BOUNCING_BAR, TOGGLE, GAUGE, VERTICAL_GAUGE, ARROWS, CLOCK, MOON, SQUARE_CORNERS, GROWING_DOTS, BOUNCING_BALL
- Custom frames via `frameSet()` or explicit frame strings
- CSS properties: `spinner-style`, `spinner-frames`
- State advancement required each tick via `state.advance()`

**WaveText**
"Animated text widget with a wave brightness effect." Configuration:
- Color specification
- Dim factor (0.0-1.0, default 0.3)
- Peak width and count
- Speed multiplier
- Mode: LOOP or OSCILLATE
- Inverted option for effect direction

### Input Widgets

**TextInput**
Single-line text field with state management:
- Placeholder text
- Cursor styling
- Methods: `insert(char)`, `deleteBackward()`, `deleteForward()`, `moveCursorLeft()`/`Right()`

**TextArea**
Multi-line editor supporting:
- Line number display
- Scrolling with `scrollUp()`/`scrollDown()`
- Multi-character insertion
- Cursor positioning methods

**Checkbox**
Toggleable checkbox with customizable symbols and colors for checked/unchecked states.

**Toggle**
Switch widget supporting two modes:
- Single symbol: displays ON/OFF text
- Inline choice: shows labeled options with selection indicators

**Select**
"Inline select widget showing current selection with navigation indicators." Wrapping navigation between options.

### Utility Widgets

**Scrollbar**
Visual scroll position indicator with orientation options (VERTICAL_RIGHT, etc.) and custom styling.

**Logo**
TamboUI branding in two sizes: TINY (2 lines, braille characters) and NORMAL (4 lines, box-drawing characters).

**ErrorDisplay**
Exception visualization with formatted stack trace, customizable title/footer, and scrollable content.

## Toolkit DSL Integration

Factory methods provide fluent widget construction:

```java
text("Hello").bold().cyan()
panel("Title", text("child")).rounded()
list("Item 1", "Item 2").highlightColor(Color.YELLOW)
spinner().cyan()
tree(rootNode).highlightColor(Color.CYAN).scrollbar()
```

Stateful widgets like Spinner and ListElement manage state internally without external state objects.

## Key Architectural Patterns

**State Management**: Stateless widgets render directly; stateful widgets require external state objects passed to render methods.

**CSS Styling**: Component-specific selectors enable terminal UI customization through style sheets.

**Model/View Separation**: TreeElement exemplifies clean architecture through data-agnostic rendering via custom node renderers.
