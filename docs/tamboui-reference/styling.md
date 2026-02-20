# TamboUI CSS Styling Documentation

## Overview

TamboUI v0.2.0-SNAPSHOT provides CSS-based styling for terminal UIs through `.tcss` (TamboUI CSS) files. This approach offers separation of concerns, runtime theming, and designer-friendly configuration compared to programmatic styling chains.

## Why CSS?

Rather than chaining method calls like `.bold().cyan().onBlue()`, CSS styling enables:

- **Separation of concerns** - styling defined outside application code
- **Runtime theming** - switch between light/dark themes dynamically
- **Non-developer friendly** - designers can adjust colors and spacing
- **Consistency** - define styles once, reuse everywhere via classes

## StyleEngine

The `StyleEngine` manages stylesheets and style resolution:

```java
StyleEngine engine = StyleEngine.create();
engine.addStylesheet("Panel { border-type: rounded; }");
engine.loadStylesheet("dark", "/themes/dark.tcss");
engine.setActiveStylesheet("dark");
```

## TCSS Format

### Variables

Define reusable values with `$` prefix:

```tcss
$bg-primary: black;
$fg-primary: white;
$accent: cyan;

Panel {
    background: $bg-primary;
    color: $fg-primary;
}
```

### Selectors

**Type selectors** match element classes:
```tcss
Panel { border-type: rounded; }
```

**Class selectors** target CSS classes:
```tcss
.primary { color: cyan; text-style: bold; }
```

**ID selectors** target specific elements:
```tcss
#sidebar { width: 30; background: dark-gray; }
```

**Pseudo-class selectors** match element state:
```tcss
Panel:focus { border-color: cyan; }
Button:hover { text-style: bold; }
ListElement-item:selected { background: blue; }
```

Supported pseudo-classes: `:focus`, `:hover`, `:disabled`, `:active`, `:selected`, `:first-child`, `:last-child`, `:nth-child(even)`, `:nth-child(odd)`

**Compound selectors** combine conditions without spaces:
```tcss
Panel.primary#main { border-color: cyan; }
```

**Descendant and child combinators**:
```tcss
Panel Button { color: white; }      /* Any Button inside Panel */
Panel > Button { text-style: bold; } /* Direct children only */
```

**Selector lists** apply styles to multiple selectors:
```tcss
.error, .warning, .danger { text-style: bold; }
```

**Attribute selectors** match based on element attributes:
```tcss
Panel[title="Settings"] { border-color: cyan; }
Panel[title] { border-type: double; }
Panel[title^="Test"] { border-color: yellow; }     /* starts with */
Panel[title$="Output"] { border-color: green; }    /* ends with */
Panel[title*="Tree"] { border-color: magenta; }    /* contains */
```

Available attributes by element type:

| Element | Attributes |
|---------|-----------|
| Panel | title, bottom-title |
| DialogElement | title |
| ListElement | title |
| TableElement | title |
| TabsElement | title |
| GaugeElement | title, label |
| LineGaugeElement | label |
| TextInputElement | title, placeholder |
| TextAreaElement | title, placeholder |

### Nesting

Use `&` for nested rules:

```tcss
Panel {
    border-type: rounded;
    border-color: gray;

    &:focus {
        border-color: cyan;
    }

    &.primary {
        border-color: blue;
    }
}
```

### Style Properties

| Property | Values | Example |
|----------|--------|---------|
| color | Named colors, hex, rgb | `color: cyan;` |
| background | Named colors, hex, rgb | `background: black;` |
| text-style | bold, dim, italic, underlined, reversed | `text-style: bold;` |
| border-type | plain, rounded, double, thick | `border-type: rounded;` |
| border-color | Colors | `border-color: cyan;` |
| padding | Single or 4 values | `padding: 1;` |
| text-align | left, center, right | `text-align: center;` |

Named colors: black, red, green, yellow, blue, magenta, cyan, white, gray, dark-gray, and bright variants.

### Layout Properties

| Property | Values | Example |
|----------|--------|---------|
| height | 5, fill, 50%, fill(2) | `height: fill;` |
| width | 10, fit, 25% | `width: fit;` |
| flex | start, center, end, space-between, space-around, space-evenly | `flex: center;` |
| spacing | Gap between children | `spacing: 1;` |
| margin | Single or 4 values | `margin: 2;` |
| direction | horizontal, vertical | `direction: vertical;` |
| column-count | Number of columns | `column-count: 3;` |
| grid-size | Columns or columns/rows | `grid-size: 3 4;` |
| grid-columns | Column constraints | `grid-columns: fill fill(2) 20;` |
| grid-rows | Row constraints | `grid-rows: 2 3;` |

#### Constraint Values

| Value | Description |
|-------|-------------|
| `<number>` | Fixed size in cells |
| `<number>%` | Percentage of space |
| `fill` | Fill available space (weight 1) |
| `fill(<weight>)` | Fill with specified weight |
| `<n>fr` | Fractional unit (1fr = fill(1)) |
| `fit` | Size to content |
| `min(<value>)` | Minimum size |
| `max(<value>)` | Maximum size |
| `<n>/<d>` | Ratio |

#### Flex Layout

Flex controls positioning of remaining space in Row/Column containers:

```tcss
.toolbar { flex: space-between; spacing: 1; }
.sidebar { flex: center; }
#footer { flex: end; }
```

Flex modes:
- `start` - pack at start
- `center` - center items
- `end` - pack at end
- `space-between` - spread items with gaps
- `space-around` - equal space around items
- `space-evenly` - equal space everywhere

#### Grid Template Areas

Named regions that span multiple cells:

```tcss
.dashboard {
    grid-template-areas: "header header header; nav main main; footer footer footer";
    grid-gutter: 1;
}
```

Assignment is programmatic:
```java
grid().area("header", text("Title")).area("main", content)
```

### Importance

Override specificity with `!important`:

```tcss
.error { color: red !important; }
```

## Property System

### PropertyDefinition

Defines a CSS property with type and converter:

```java
PropertyDefinition<Color> COLOR =
    PropertyDefinition.builder("color", ColorConverter.INSTANCE)
        .inheritable()
        .build();
```

### PropertyRegistry

Register custom properties:

```java
PropertyRegistry.register(MY_PROPERTY);
PropertyRegistry.registerAll(PROP_A, PROP_B);
```

Widgets should register properties in static blocks:

```java
public static class MyWidget {
    public static final PropertyDefinition<Color> BAR_COLOR =
        PropertyDefinition.of("bar-color", ColorConverter.INSTANCE);

    static {
        PropertyRegistry.registerAll(BAR_COLOR);
    }
}
```

### StandardProperties

Core properties all widgets can use:

| Property | Type | Inherits | Description |
|----------|------|----------|-------------|
| color | Color | Yes | Foreground/text color |
| text-style | Set<Modifier> | Yes | Bold, dim, italic, etc. |
| border-type | BorderType | Yes | Border style |
| background | Color | No | Background color |
| border-color | Color | No | Border color |
| padding | Padding | No | Inner spacing |
| margin | Margin | No | Outer spacing |
| width, height | Constraint | No | Size constraints |

### Property Inheritance

**Inherited properties** flow from parent to children:

```tcss
Panel.sidebar {
    color: cyan;      /* Inherited */
    text-style: dim;  /* Inherited */
}
```

All descendants receive these values unless overridden.

**Non-inherited properties** apply only to the element:

```tcss
Panel.sidebar {
    background: dark-gray;  /* Not inherited */
    padding: 1;             /* Not inherited */
}
```

Inherited: `color`, `text-style`, `border-type`
Non-inherited: `background`, `padding`, `margin`, `width`, `height`, `flex`, `direction`, `spacing`

## Applying Styles

### With Toolkit DSL

Elements automatically support CSS via classes and IDs:

```java
Element panel = panel("Settings",
    () -> column(
        text("Username").addClass("label"),
        textInput(state).id("username-input")
    )
).id("settings-panel").addClass("primary");
```

### With ToolkitRunner

```java
StyleEngine engine = StyleEngine.create();
engine.loadStylesheet("dark", "/themes/dark.tcss");
engine.setActiveStylesheet("dark");

try (var runner = ToolkitRunner.builder()
        .styleEngine(engine)
        .build()) {
    runner.run(() -> myApp());
}
```

### Implementing Styleable

For custom widgets, implement `Styleable`:

```java
public class MyWidget implements Styleable {
    private String id;
    private Set<String> classes = new HashSet<>();

    @Override
    public String styleType() { return "MyWidget"; }

    @Override
    public Optional<String> cssId() { return Optional.ofNullable(id); }

    @Override
    public Set<String> cssClasses() { return classes; }

    @Override
    public Optional<Styleable> cssParent() { return Optional.empty(); }
}
```

Resolve and apply styles:

```java
CssStyleResolver resolved = engine.resolve(widget);
Style style = Style.EMPTY;
if (resolved.foreground().isPresent()) {
    style = style.fg(resolved.foreground().get());
}
```

## Theme Switching

Switch themes at runtime:

```java
String current = engine.getActiveStylesheet().orElse("dark");
String next = "dark".equals(current) ? "light" : "dark";
engine.setActiveStylesheet(next);
```

Listen for changes:

```java
engine.addChangeListener(() -> requestRedraw());
```

## Example Theme Files

**dark.tcss**

```tcss
$bg-primary: black;
$fg-primary: white;
$accent: cyan;
$border-color: dark-gray;

* {
    color: $fg-primary;
    background: $bg-primary;
}

Panel {
    border-type: rounded;
    border-color: $border-color;
}

Panel:focus {
    border-color: $accent;
    border-type: double;
}

.primary {
    color: $accent;
    text-style: bold;
}

.danger {
    color: red;
}
```

**light.tcss**

```tcss
$bg-primary: white;
$fg-primary: black;
$accent: blue;
$border-color: gray;

* {
    color: $fg-primary;
    background: $bg-primary;
}

Panel {
    border-type: rounded;
    border-color: $border-color;
}

Panel:focus {
    border-color: $accent;
    border-type: double;
}

.primary {
    color: $accent;
    text-style: bold;
}
```
