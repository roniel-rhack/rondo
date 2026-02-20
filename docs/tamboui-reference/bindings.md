# TamboUI Bindings and Actions Documentation

## Overview

TamboUI version 0.2.0-SNAPSHOT separates physical input from semantic actions. Rather than checking for specific keys throughout code, developers define bindings mapping inputs to actions like "moveUp" or "delete".

## Predefined Binding Sets

Three preset binding configurations are available:

- `BindingSets.standard()` - Arrow keys, Enter, Escape
- `BindingSets.vim()` - hjkl navigation, Ctrl+u/d
- `BindingSets.emacs()` - Ctrl+n/p/f/b navigation

## Matching Actions in Events

Events provide both semantic and low-level checking methods:

**Semantic approaches** (work across binding sets):
```java
if (event.isUp()) { }
if (event.isSelect()) { }
if (event.matches("delete")) { }
```

**Low-level approaches** (specific keys):
```java
if (event.isChar('x')) { }
if (event.isKey(KeyCode.F1)) { }
```

The `is*()` methods delegate to current bindings. With vim bindings, `isUp()` matches the Up arrow, 'k', or Ctrl+p.

## Key Triggers

`KeyTrigger` defines input that activates actions:

```java
KeyTrigger.key(KeyCode.UP);           // Arrow key
KeyTrigger.ch('j');                   // Character
KeyTrigger.chIgnoreCase('j');         // j or J
KeyTrigger.ctrl('u');                 // Ctrl+U
KeyTrigger.alt('x');                  // Alt+X
```

### Terminal Compatibility

**Ctrl key limitations**: Terminals encode key combinations as single ASCII control characters (from the 1970s). Ctrl+a and Ctrl+Shift+A send identical bytes -- case information is lost. Ctrl+number combinations are unreliable, with Ctrl+2, Ctrl+@, and Ctrl+Space all sending the same byte (0).

**Alt key advantages**: Terminals transmit ESC followed by the actual character, preserving case and distinguishing numbers reliably.

**Compatibility recommendations**:

| Modifier | Reliability | Notes |
|----------|------------|-------|
| Alt+letter | Excellent | Preserves case, consistent |
| Alt+number | Excellent | Fully distinguishable |
| Ctrl+letter | Good | Works but loses shift info |
| Ctrl+number | Poor | Avoid -- ambiguous |

For custom bindings, prefer `Alt+key` combinations. Use `Ctrl+key` only for established conventions (Ctrl+C, Ctrl+S, Ctrl+Z).

## Mouse Triggers

```java
MouseTrigger.click();                 // Left click
MouseTrigger.rightClick();            // Right click
MouseTrigger.ctrlClick();             // Ctrl+click
MouseTrigger.scrollUp();              // Scroll wheel
MouseTrigger.drag(MouseButton.LEFT);  // Dragging
```

## Custom Bindings

Start from presets and modify:

```java
Bindings custom = BindingSets.standard()
    .toBuilder()
    .bind(KeyTrigger.ch('d'), "delete")
    .bind(KeyTrigger.key(KeyCode.DELETE), "delete")
    .bind(KeyTrigger.ctrl('s'), "save")
    .bind(MouseTrigger.rightClick(), "contextMenu")
    .build();
```

## ActionHandler

`ActionHandler` provides centralized action handling:

```java
ActionHandler actions = new ActionHandler(BindingSets.vim())
    .on(Actions.QUIT, e -> runner.quit())
    .on("save", e -> save())
    .on("delete", e -> deleteSelected());

boolean handleEvent(Event event, TuiRunner runner) {
    ActionHandler actions = new ActionHandler(BindingSets.vim());
    if (actions.dispatch(event)) {
        return true;  // Action was handled
    }
    return false;
}
```

Handlers can receive action names:

```java
ActionHandler actions = new ActionHandler(bindings)
    .on("red", (event, action) -> setColor(action))
    .on("blue", (event, action) -> setColor(action))
    .on("green", (event, action) -> setColor(action));
```

## @OnAction Annotation

This annotation works with Toolkit Components. For TuiRunner or immediate mode, use `ActionHandler` instead.

In Toolkit components, annotate methods:

```java
public class EditorComponent extends Component<EditorComponent> {
    private List<String> lines = new ArrayList<>();
    private int cursor = 0;

    @OnAction(Actions.MOVE_UP)
    void moveCursorUp(Event event) {
        if (cursor > 0) cursor--;
    }

    @OnAction(Actions.MOVE_DOWN)
    void moveCursorDown(Event event) {
        if (cursor < lines.size() - 1) cursor++;
    }

    @OnAction("delete")
    void deleteLine(Event event) {
        if (!lines.isEmpty()) {
            lines.remove(cursor);
        }
    }

    @Override
    protected Element render() {
        return text("Editor");
    }
}
```

The component framework automatically discovers and dispatches to these methods.

### Annotation Processing

By default, `@OnAction` methods are discovered via runtime reflection. For better startup performance and GraalVM native image compatibility, use the annotation processor for compile-time generation.

**Gradle (Kotlin DSL)**:
```kotlin
dependencies {
    implementation("dev.tamboui:tamboui-toolkit:0.2.0-SNAPSHOT")
    annotationProcessor("dev.tamboui:tamboui-processor:0.2.0-SNAPSHOT")
}
```

**Gradle (Groovy DSL)**:
```groovy
dependencies {
    implementation 'dev.tamboui:tamboui-toolkit:0.2.0-SNAPSHOT'
    annotationProcessor 'dev.tamboui:tamboui-processor:0.2.0-SNAPSHOT'
}
```

**Maven**:
```xml
<dependencies>
    <dependency>
        <groupId>dev.tamboui</groupId>
        <artifactId>tamboui-toolkit</artifactId>
        <version>0.2.0-SNAPSHOT</version>
    </dependency>
</dependencies>

<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-compiler-plugin</artifactId>
            <version>3.13.0</version>
            <configuration>
                <annotationProcessorPaths>
                    <path>
                        <groupId>dev.tamboui</groupId>
                        <artifactId>tamboui-processor</artifactId>
                        <version>0.2.0-SNAPSHOT</version>
                    </path>
                </annotationProcessorPaths>
            </configuration>
        </plugin>
    </plugins>
</build>
```

The processor generates `_ActionHandlerRegistrar` classes for each component with `@OnAction` methods. These are discovered via `ServiceLoader` and register handlers without reflection.

## Loading Bindings from Files

Bindings can be externalized to properties files:

```java
// From classpath
Bindings bindings = BindingSets.loadResource("/my-bindings.properties");

// From filesystem
Bindings bindingsFromFile = BindingSets.load(Paths.get("~/.config/myapp/bindings.properties"));
```

**Property file format**:
```properties
# Navigation
moveUp = Up, k
moveDown = Down, j
pageUp = PageUp, Ctrl+u
pageDown = PageDown, Ctrl+d

# Actions
select = Enter, Space
cancel = Escape
delete = d, Delete

# Mouse
click = Mouse.Left.Press
contextMenu = Mouse.Right.Press
```

Multiple triggers for the same action are comma-separated. Modifiers use `Ctrl+`, `Alt+`, `Shift+` prefixes.

## Standard Actions

The `Actions` class defines common action constants:

| Action | Description |
|--------|-------------|
| `MOVE_UP`, `MOVE_DOWN`, `MOVE_LEFT`, `MOVE_RIGHT` | Navigation |
| `PAGE_UP`, `PAGE_DOWN`, `HOME`, `END` | Page navigation |
| `SELECT`, `CONFIRM`, `CANCEL` | Selection and confirmation |
| `FOCUS_NEXT`, `FOCUS_PREVIOUS` | Focus navigation (Tab/Shift+Tab) |
| `DELETE_BACKWARD`, `DELETE_FORWARD` | Text editing |
| `QUIT` | Application exit |

## Next Steps

- **API Levels** - choosing between immediate mode, TuiRunner, and Toolkit
- **Application Structure** - patterns for larger applications
