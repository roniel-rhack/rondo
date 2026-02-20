# TamboUI Layouts Reference

## Overview

TamboUI provides layout widgets for arranging child components into structured compositions. These widgets are stateless and located in `dev.tamboui.layout.*` subpackages within the `tamboui-core` module. Version: 0.2.0-SNAPSHOT

## Layout Widgets Summary

| Widget | Type | Purpose |
|--------|------|---------|
| Columns | Stateless | "Multi-column grid layout" |
| Grid | Stateless | "CSS Grid-inspired layout with constraints and gutter" |
| Dock | Stateless | "5-region layout (top, bottom, left, right, center)" |
| Stack | Stateless | "Overlapping layers (painter's algorithm)" |
| Flow | Stateless | "Wrap layout (flow left-to-right, wrap on overflow)" |

---

## Columns Layout

A grid widget that arranges children into a fixed number of columns with position determined by ordering mode.

### Basic Usage

```java
Columns columns = Columns.builder()
    .children(widget1, widget2, widget3, widget4)
    .columnCount(2)
    .spacing(1)
    .order(ColumnOrder.ROW_FIRST)
    .build();

columns.render(area, buffer);
```

### Ordering Modes

- **ColumnOrder.ROW_FIRST** (default): "Items fill left-to-right, then top-to-bottom (like reading text)"
- **ColumnOrder.COLUMN_FIRST**: "Items fill top-to-bottom, then left-to-right (like newspaper columns)"

### Builder Options

- `children(Widget...)` / `children(List<Widget>)` - Child widgets to arrange
- `columnCount(int)` - Number of columns (required, defaults to 1)
- `spacing(int)` - Gap between columns in cells (default: 0)
- `flex(Flex)` - Space distribution (default: `Flex.START`)
- `order(ColumnOrder)` - Child ordering mode (default: `ROW_FIRST`)
- `columnWidths(Constraint...)` - Per-column width constraints (default: `fill()` for all)
- `rowHeights(int...)` - Explicit row heights (default: equal distribution)

### Advanced Examples

```java
// Custom column widths and row-first ordering
Columns grid = Columns.builder()
    .children(headerWidget, contentWidget, sidebarWidget, footerWidget)
    .columnCount(2)
    .columnWidths(Constraint.length(20), Constraint.fill())
    .rowHeights(3, 10)
    .spacing(1)
    .build();

// Column-first ordering: fills top-to-bottom per column
Columns newspaperLayout = Columns.builder()
    .children(article1, article2, article3, article4)
    .columnCount(2)
    .order(ColumnOrder.COLUMN_FIRST)
    .build();
```

**Note**: The widget-level `Columns` requires explicit column count. Use `ColumnsElement` in the Toolkit DSL for auto-detection based on child widths.

---

## Grid Layout

A CSS Grid-inspired layout with explicit control over dimensions, sizing constraints, and gutter spacing. Unlike `Columns`, `Grid` provides independent horizontal and vertical gutters, per-column width constraints, per-row height constraints, and constraint cycling.

Grid supports two mutually exclusive modes:
- **Children mode**: Sequential placement using `children()` and `columnCount()`
- **Area mode**: Template-based placement using `gridAreas()` (CSS grid-template-areas style)

### Children Mode

#### Basic Usage

```java
Grid grid = Grid.builder()
    .children(widget1, widget2, widget3, widget4, widget5, widget6)
    .columnCount(3)
    .horizontalGutter(1)
    .verticalGutter(1)
    .build();

grid.render(area, buffer);
```

#### ChildrenBuilder Options

- `children(Widget...)` / `children(List<Widget>)` - Child widgets to arrange
- `columnCount(int)` - Number of columns (defaults to 1)
- `rowCount(int)` - Optional maximum rows (validates children fit)
- `horizontalGutter(int)` - Gap between columns (default: 0)
- `verticalGutter(int)` - Gap between rows (default: 0)
- `flex(Flex)` - Space distribution (default: `Flex.START`)
- `columnConstraints(Constraint...)` - Per-column width constraints (cycles when fewer than column count)
- `rowConstraints(Constraint...)` - Per-row height constraints (cycles when fewer than row count)
- `rowHeights(int...)` - Explicit row heights (default: equal distribution)

#### Children Mode Examples

```java
// Grid with custom column widths and gutter
Grid dashboard = Grid.builder()
    .children(cpuPanel, memoryPanel, diskPanel,
              netUpPanel, netDownPanel, uptimePanel)
    .columnCount(3)
    .columnConstraints(
        Constraint.length(16),  // fixed first column
        Constraint.fill(),      // remaining columns fill
        Constraint.fill()
    )
    .horizontalGutter(2)
    .verticalGutter(1)
    .build();

// Grid with row constraints
Grid sized = Grid.builder()
    .children(header, content, sidebar, footer)
    .columnCount(2)
    .rowConstraints(Constraint.length(3), Constraint.fill())
    .build();

// Column constraint cycling: single constraint applied to all columns
Grid uniform = Grid.builder()
    .children(a, b, c, d, e, f)
    .columnCount(3)
    .columnConstraints(Constraint.length(10))  // all 3 cols get length(10)
    .build();
```

### Area Mode (Grid Template Areas)

Area mode uses CSS `grid-template-areas` style templates where cells can span multiple rows and columns. "Named areas must form contiguous rectangles -- L-shapes and disconnected regions are rejected with a `LayoutException`."

#### Basic Usage

```java
// "Holy grail" layout with spanning regions
Grid layout = Grid.builder()
    .gridAreas("header header header",
               "nav    main   main",
               "nav    main   main",
               "footer footer footer")
    .area("header", headerWidget)
    .area("nav", navWidget)
    .area("main", mainWidget)
    .area("footer", footerWidget)
    .horizontalGutter(1)
    .verticalGutter(1)
    .build();

layout.render(area, buffer);
```

#### AreaBuilder Options

- `gridAreas(String...)` - Row templates defining named areas (use `.` for empty cells)
- `area(String, Widget)` - Assign widget to named area
- `horizontalGutter(int)` - Gap between columns (default: 0)
- `verticalGutter(int)` - Gap between rows (default: 0)
- `flex(Flex)` - Space distribution (default: `Flex.START`)
- `columnConstraints(Constraint...)` - Per-column width constraints
- `rowConstraints(Constraint...)` - Per-row height constraints

#### Template Rules

- Each row is a space-separated list of area names
- All rows must have the same number of columns
- Area names must start with a letter (alphanumeric and underscores allowed)
- Use `.` (dot) for empty cells
- "Named areas must form contiguous rectangles"

#### Area Mode Examples

```java
// Dashboard with 2x2 spanning main area
Grid dashboard = Grid.builder()
    .gridAreas("A A B",
               "A A C",
               "D D D")
    .area("A", mainPanel)    // 2x2 span
    .area("B", sidePanel1)
    .area("C", sidePanel2)
    .area("D", statusBar)    // full-width span
    .horizontalGutter(1)
    .verticalGutter(1)
    .build();

// Empty cells with dot notation
Grid sparse = Grid.builder()
    .gridAreas("A . B",
               ". C .")
    .area("A", widget1)
    .area("B", widget2)
    .area("C", widget3)
    .build();
```

**Important**: Areas without assigned widgets render as empty space. Assigning a widget to an undefined area throws `LayoutException`.

**Note**: The widget-level `Grid` requires explicit column count or grid areas template. Use `GridElement` in the Toolkit DSL for auto-sizing based on child count.

---

## Dock Layout

A 5-region dock layout that arranges children into top, bottom, left, right, and center regions -- "the most common TUI application structure (header + sidebar + content + footer)."

### Basic Usage

```java
Dock dock = Dock.builder()
    .top(headerWidget)
    .bottom(statusBarWidget)
    .left(sidebarWidget)
    .right(outlineWidget)
    .center(editorWidget)
    .topHeight(Constraint.length(3))
    .bottomHeight(Constraint.length(1))
    .leftWidth(Constraint.length(20))
    .rightWidth(Constraint.length(20))
    .build();

dock.render(area, buffer);
```

### Builder Options

- `.top(Widget)` / `.bottom(Widget)` / `.left(Widget)` / `.right(Widget)` / `.center(Widget)` - Set region widgets (all optional)
- `.topHeight(Constraint)` - Height constraint for top region (default: `length(1)`)
- `.bottomHeight(Constraint)` - Height constraint for bottom region (default: `length(1)`)
- `.leftWidth(Constraint)` - Width constraint for left region (default: `length(10)`)
- `.rightWidth(Constraint)` - Width constraint for right region (default: `length(10)`)

### Rendering Algorithm

"Top and bottom take full width, then the remaining middle area is split horizontally into left, center, and right. Omitted regions are skipped -- for example, setting only `center` gives a full-area layout."

---

## Stack Layout

An overlapping layers widget where children render on top of each other using "painter's algorithm (last child on top). Essential for dialogs, popups, floating overlays, and any scenario where UI elements need to overlap."

### Basic Usage

```java
Stack stack = Stack.builder()
    .children(backgroundWidget, dialogWidget)
    .alignment(ContentAlignment.STRETCH)
    .build();

stack.render(area, buffer);
```

### Builder Options

- `.children(Widget...)` / `.children(List<Widget>)` - Child widgets to stack
- `.alignment(ContentAlignment)` - How children are positioned (default: `STRETCH`)

### Alignment Modes

`TOP_LEFT`, `TOP_CENTER`, `TOP_RIGHT`, `CENTER_LEFT`, `CENTER`, `CENTER_RIGHT`, `BOTTOM_LEFT`, `BOTTOM_CENTER`, `BOTTOM_RIGHT`, `STRETCH`

---

## Flow Layout

A wrap layout widget where "items flow left-to-right and wrap to the next line when exceeding the available width. Useful for tag clouds, button groups, chip lists, and similar layouts where items should wrap naturally."

### Basic Usage

```java
Flow flow = Flow.builder()
    .item(tag1Widget, 8)
    .item(tag2Widget, 12)
    .item(tag3Widget, 6)
    .horizontalSpacing(1)
    .verticalSpacing(1)
    .build();

flow.render(area, buffer);
```

### Builder Options

- `.item(Widget, int width)` / `.item(Widget, int width, int height)` - Add item with explicit size
- `.items(List<FlowItem>)` - Set items from a list
- `.horizontalSpacing(int)` - Gap between items on same row (default: 0)
- `.verticalSpacing(int)` - Gap between rows (default: 0)

**Note**: "The widget-level `Flow` requires explicit item widths via `FlowItem` since the `Widget` interface has no `preferredWidth()` method. For auto-measurement, use `FlowElement` in the Toolkit DSL (via the `flow()` factory method), which auto-measures children via `Element.preferredWidth()`."

---

## Using Layouts with Toolkit DSL

The Toolkit DSL provides fluent factories for all layout widgets:

```java
import static dev.tamboui.toolkit.Toolkit.*;

// Multi-column grid (auto-detects column count from child widths)
columns(item1, item2, item3, item4, item5, item6)
    .spacing(1)

// Explicit column count with column-first ordering
columns(child1, child2, child3, child4)
    .columnCount(2)
    .columnFirst()

// CSS Grid-inspired layout with gutter and constraints
grid(item1, item2, item3, item4, item5, item6)
    .gridSize(3)
    .gutter(1)

// Grid with explicit dimensions and column constraints
grid(header, content, sidebar, footer)
    .gridSize(2, 2)
    .gridColumns(Constraint.length(20), Constraint.fill())
    .gutter(1, 0)

// Grid with template areas (CSS grid-template-areas style)
grid()
    .gridAreas("header header header",
               "nav    main   main",
               "nav    main   main",
               "footer footer footer")
    .area("header", text("Header").bold())
    .area("nav", list("Nav 1", "Nav 2"))
    .area("main", text("Main Content"))
    .area("footer", text("Footer").dim())
    .gutter(1)

// 5-region dock layout
dock()
    .top(text("Header").bold())
    .bottom(text("Footer").dim())
    .left(list("Nav 1", "Nav 2", "Nav 3"))
    .center(text("Main Content"))
    .topHeight(Constraint.length(3))
    .leftWidth(Constraint.length(20))

// Overlapping layers (last child on top)
stack(backgroundElement, dialogElement)
    .alignment(ContentAlignment.CENTER)

// Wrap layout (items flow and wrap)
flow(tag1, tag2, tag3, tag4, tag5)
    .spacing(1)
    .rowSpacing(1)
```

See the API Levels documentation for additional details on the Toolkit DSL.
