# TamboUI API Levels

## Overview

TamboUI is organized in layers, progressing from low-level terminal control to high-level declarative interfaces. As stated in the docs: "At the bottom sits the immediate-mode API for direct terminal control. On top of that, TuiRunner adds a managed event loop."

## API Layers

| Layer | Purpose | Best For |
|-------|---------|----------|
| **Immediate Mode** | Direct terminal and buffer control | Custom backends, game engines, learning |
| **TuiRunner** | Managed event loop with callbacks | Custom event handling, animations |
| **Toolkit DSL** | Declarative component-based UI | Most applications |
| **Inline Display** | Fixed status area preserving scroll | Progress bars, logs |
| **Inline TuiRunner** | Event loop for inline displays | Interactive inline apps |
| **Inline Toolkit** | Declarative inline elements | Most inline applications |

---

## Immediate Mode

Direct control over terminal, backend, and event loop.

### Terminal Setup

```java
try (var backend = new JLineBackend()) {
    backend.enableRawMode();
    backend.enterAlternateScreen();
    backend.hideCursor();
    var terminal = new Terminal<>(backend);
}
```

### Drawing

```java
terminal.draw(frame -> {
    Rect area = frame.area();
    var paragraph = Paragraph.builder()
        .text(Text.from("Hello!"))
        .style(Style.EMPTY.bold().fg(Color.CYAN))
        .build();
    frame.renderWidget(paragraph, area);
});
```

---

## TuiRunner

Manages terminal setup and provides an event loop with two callbacks: event handling and rendering.

### Basic Usage

```java
try (var tui = TuiRunner.create()) {
    tui.run(
        (event, runner) -> {
            if (event instanceof KeyEvent key && key.isQuit()) {
                runner.quit();
                return false;
            }
            return handleEvent(event);
        },
        frame -> renderUI(frame)
    );
}
```

Event handlers return `true` to request redraw, `false` otherwise.

### Configuration

```java
var config = TuiConfig.builder()
    .mouseCapture(true)
    .tickRate(Duration.ofMillis(16))
    .pollTimeout(Duration.ofMillis(50))
    .build();

try (var tui = TuiRunner.create(config)) {
    tui.run(handler, renderer);
}
```

### Disabling Ticks

For event-driven UIs without animations:

```java
var config = TuiConfig.builder()
    .noTick()
    .resizeGracePeriod(Duration.ofMillis(100))
    .build();
```

### Error Handling

Built-in error handlers include `displayAndQuit()` (default), `logAndQuit(PrintStream)`, `writeToFile(Path)`, and `suppress()`.

```java
var config = TuiConfig.builder()
    .errorHandler(RenderErrorHandlers.logAndQuit(
        new PrintStream("/tmp/tui-errors.log")))
    .build();
```

### Threading Model

TamboUI uses a dedicated render thread. All UI modifications must occur on this thread.

**Checking thread context:**
```java
if (runner.isRenderThread()) {
    // Safe to modify UI state
}
```

**Updates from background threads:**
```java
runner.runOnRenderThread(() -> {
    // Queued to run on render thread
    status = "Download complete";
});
```

**Always-deferred actions:**
```java
runner.runLater(() -> {
    // Runs after current event handling
    processNextItem();
});
```

### Scheduled Actions

```java
runner.schedule(() -> {
    runner.runOnRenderThread(() -> {
        countdown--;
    });
}, Duration.ofSeconds(1));

runner.scheduleRepeating(() -> {
    runner.runOnRenderThread(() -> animationFrame++);
}, Duration.ofMillis(16));
```

### Semantic Key Checks

"KeyEvent provides semantic methods that respect the configured bindings":

```java
if (key.isQuit()) { }      // q, Q, Ctrl+C
if (key.isUp()) { }        // Up, k (vim), Ctrl+P
if (key.isDown()) { }      // Down, j (vim), Ctrl+N
if (key.isSelect()) { }    // Enter, Space
if (key.isCancel()) { }    // Escape
if (key.isPageUp()) { }    // PageUp, Ctrl+B
if (key.isPageDown()) { }  // PageDown, Ctrl+F
```

### Event Handling Patterns

```java
EventHandler handler = (event, runner) -> {
    if (event instanceof KeyEvent k) return handleKey(k);
    if (event instanceof MouseEvent m) return handleMouse(m);
    if (event instanceof TickEvent) { animationFrame++; return true; }
    if (event instanceof ResizeEvent) return true;
    return false;
};
```

### Layout in Renderer

```java
Renderer renderer = frame -> {
    var areas = Layout.vertical()
        .constraints(
            Constraint.length(3),
            Constraint.fill(),
            Constraint.length(1)
        )
        .split(frame.area());

    frame.renderWidget(header, areas.get(0));
    frame.renderWidget(content, areas.get(1));
    frame.renderWidget(statusBar, areas.get(2));
};
```

### Example: Counter with Animation

```java
public class CounterDemo {
    private int counter = 0;
    private int ticks = 0;

    public void run() throws Exception {
        var config = TuiConfig.builder()
            .tickRate(Duration.ofMillis(100))
            .build();

        try (var tui = TuiRunner.create(config)) {
            tui.run(this::handleEvent, this::render);
        }
    }

    private boolean handleEvent(Event event, TuiRunner runner) {
        if (event instanceof KeyEvent key && key.isQuit()) {
            runner.quit();
            return false;
        }
        if (event instanceof TickEvent) {
            ticks++;
            return true;
        }
        if (event instanceof KeyEvent key) {
            if (key.isChar('+')) { counter++; return true; }
            if (key.isChar('-')) { counter--; return true; }
        }
        return false;
    }

    private void render(Frame frame) {
        var text = String.format(
            "Counter: %d (ticks: %d)%n%nPress +/- to change, q to quit",
            counter, ticks);

        var widget = Paragraph.builder()
            .text(Text.from(text))
            .block(Block.builder()
                .title("Demo")
                .borders(Borders.ALL)
                .borderType(BorderType.ROUNDED)
                .build())
            .build();

        frame.renderWidget(widget, frame.area());
    }
}
```

---

## Toolkit DSL

Declarative, component-based UI using fluent builders and element factories.

### Basic App Structure

```java
public static class MyApp extends ToolkitApp {
    @Override
    protected Element render() {
        return panel("My App",
            text("Hello!").bold().cyan(),
            spacer(),
            text("Press 'q' to quit").dim()
        ).rounded();
    }

    public static void main(String[] args) throws Exception {
        new MyApp().run();
    }
}
```

### Layout: Rows and Columns

```java
row(
    panel("Left").fill(),
    panel("Right").fill()
);

column(
    text("Header"),
    panel("Content").fill(),
    text("Footer")
);
```

### Multi-Column Grid

```java
// Auto-detect columns
columns(item1, item2, item3, item4, item5, item6)
    .spacing(1);

// Explicit column count
columns(item1, item2, item3, item4)
    .columnCount(2);

// Column-first ordering
columns(item1, item2, item3, item4)
    .columnCount(2)
    .columnFirst();
```

### Spacing and Flex

```java
row(
    text("Left"),
    spacer(),
    text("Right")
);

column(
    panel("Top"),
    panel("Bottom")
).flex(Flex.SPACE_BETWEEN);
```

### Styling

```java
text("Styled").bold().italic().cyan().onBlue();

panel("Title", text("content"))
    .rounded()
    .borderColor(Color.CYAN)
    .focusedBorderColor(Color.YELLOW);
```

### Sizing

```java
panel("Fixed width").length(30);
panel("Take what's left").fill();
panel("Twice the weight").fill(2);
panel("At least 10").min(10);
panel("At most 50").max(50);
```

### Stateful Widgets

Tables and text inputs require state objects; lists manage state internally:

```java
private ListElement<?> myList = list("Apple", "Banana", "Cherry")
    .highlightColor(Color.CYAN)
    .autoScroll();

private TableState tableState = new TableState();
private TextInputState inputState = new TextInputState();

Element statefulExample() {
    return column(
        myList,
        table()
            .header("Name", "Age")
            .row("Alice", "30")
            .row("Bob", "25")
            .state(tableState),
        textInput(inputState)
            .placeholder("Type here...")
    );
}
```

**List navigation:**
```java
myList.selectNext(itemCount);
myList.selectPrevious();
myList.selected();      // Get current index
myList.selected(0);     // Set selection
```

### Event Handling

```java
panel("Interactive")
    .id("main")
    .focusable()
    .onKeyEvent(event -> {
        if (event.isChar('a')) {
            addItem();
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    });
```

### Focus Management

Elements need both ID and focusable flag for Tab/Shift+Tab navigation:

```java
column(
    panel("First").id("first").focusable(),
    panel("Second").id("second").focusable()
);
```

### Data Visualization

```java
gauge(0.75).label("Progress").gaugeColor(Color.GREEN);
sparkline(1, 4, 2, 8, 5, 7).color(Color.CYAN);
barChart(10, 20, 30).barColor(Color.BLUE);
```

### Wrapping Low-Level Widgets

```java
widget(myCustomWidget)
    .addClass("custom")
    .fill();

row(
    widget(customWidget1).fill(),
    widget(customWidget2).fill()
);
```

**Limitations:** Styling doesn't propagate to wrapped widgets; no CSS child selectors; no preferred size inference.

### Using ToolkitRunner Directly

```java
var config = TuiConfig.builder()
    .mouseCapture(true)
    .tickRate(Duration.ofMillis(50))
    .build();

try (var runner = ToolkitRunner.create(config)) {
    runner.run(() -> panel("App", content));
}
```

### Fault-Tolerant Rendering

```java
try (var runner = ToolkitRunner.builder()
        .faultTolerant(true)
        .build()) {
    runner.run(() -> render());
}
```

With fault-tolerant mode enabled, element exceptions display as error placeholders rather than crashing the full application.

### Example: Todo List

```java
public static class TodoApp extends ToolkitApp {
    private final List<String> items = new ArrayList<>(List.of(
        "Learn TamboUI",
        "Build something cool"
    ));
    private final ListElement<?> todoList = list()
        .highlightColor(Color.CYAN)
        .autoScroll();

    @Override
    protected Element render() {
        return panel("Todo",
            items.isEmpty()
                ? text("Empty - press 'a' to add").dim()
                : todoList.items(items.toArray(new String[0])),
            spacer(),
            text("[a]dd [d]elete [q]uit").dim()
        )
        .rounded()
        .id("main")
        .focusable()
        .onKeyEvent(this::handleKey);
    }

    private EventResult handleKey(KeyEvent event) {
        if (event.isChar('a')) {
            items.add("New item");
            return EventResult.HANDLED;
        }
        if (event.isChar('d') && !items.isEmpty()) {
            items.remove(todoList.selected());
            return EventResult.HANDLED;
        }
        if (event.isDown()) {
            todoList.selectNext(items.size());
            return EventResult.HANDLED;
        }
        if (event.isUp()) {
            todoList.selectPrevious();
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    }

    public static void main(String[] args) throws Exception {
        new TodoApp().run();
    }
}
```

---

## Inline Display Mode

Fixed status area that preserves terminal scroll history, as used by build tools and package managers.

### When to Use

- Build tools showing compilation progress
- Package managers displaying download status
- Long-running scripts with progress indicators
- Tools preserving output history

### Basic Usage

```java
try (var display = InlineDisplay.create(3)) {
    for (int i = 0; i <= 100; i += 10) {
        final int progress = i;
        display.render((area, buffer) -> {
            var gauge = Gauge.builder()
                .ratio(progress / 100.0)
                .label("Processing: " + progress + "%")
                .build();
            gauge.render(area, buffer);
        });
        Thread.sleep(100);
    }
}
```

### Logging Above Status

```java
try (var display = InlineDisplay.create(4)) {
    for (var task : tasks) {
        display.render((area, buffer) -> {
            renderProgress(area, buffer, task, progress);
        });
        processTask(task);
        display.println(Text.from(Line.from(
            Span.styled("OK ", Style.EMPTY.fg(Color.GREEN)),
            Span.raw(task.name())
        )));
    }
}
```

### Configuration

```java
var display = InlineDisplay.create(4);           // Fixed height
var display2 = InlineDisplay.create(4, 80);      // Fixed height and width
var display3 = InlineDisplay.create(4)
    .clearOnClose();                              // Clear on exit
```

### Setting Lines Directly

```java
display.setLine(0, "Building module: core");
display.setLine(1, "Progress: 45%");
display.setLine(2, Text.from(
    Span.styled("No errors", Style.EMPTY.fg(Color.GREEN))));
```

### Inline vs Alternate Screen

| Feature | InlineDisplay | TuiRunner/Toolkit |
|---------|---------------|-------------------|
| Buffer | Main (preserved) | Alternate (replaced) |
| Cursor | Visible | Hidden |
| Terminal | Partial | Full |
| Previous content | Preserved | Restored |
| Best for | Progress, logs | Interactive apps |

---

## Inline TUI Runner

Managed event loop for inline displays with keyboard and mouse event handling.

### Basic Usage

```java
try (var runner = InlineTuiRunner.create(4)) {
    runner.run(
        (event, r) -> {
            if (event instanceof KeyEvent key && key.character() == 'q') {
                r.quit();
                return true;
            }
            return false;
        },
        frame -> {
            var gauge = Gauge.builder()
                .ratio(progress / 100.0)
                .build();
            gauge.render(frame.area(), frame.buffer());
        }
    );
}
```

### Configuration

```java
var config = InlineTuiConfig.builder(4)
    .tickRate(Duration.ofMillis(50))
    .clearOnClose(true)
    .build();

try (var runner = InlineTuiRunner.create(config)) {
    // ...
}
```

### Printing Above Viewport

```java
runner.println("Task completed!");

runner.println(Text.from(Line.from(
    Span.styled("OK ", Style.EMPTY.fg(Color.GREEN)),
    Span.raw("Build successful")
)));
```

### Thread-Safe Updates

```java
runner.runOnRenderThread(() -> {
    // Update state on render thread
});

runner.runLater(() -> {
    cleanup();
});
```

---

## Inline Toolkit

Declarative element-based API for inline displays.

### InlineApp Base Class

```java
public static class MyProgressApp extends InlineApp {
    private double progress = 0.0;

    public static void main(String[] args) throws Exception {
        new MyProgressApp().run();
    }

    @Override
    protected int height() {
        return 3;
    }

    @Override
    protected Element render() {
        return column(
            text("Installing packages...").bold(),
            gauge(progress).green(),
            text(String.format("%.0f%% complete", progress * 100)).dim()
        );
    }

    @Override
    protected void onStart() {
        runner().scheduleRepeating(() -> {
            runner().runOnRenderThread(() -> {
                progress += 0.01;
                if (progress >= 1.0) quit();
            });
        }, Duration.ofMillis(50));
    }
}
```

### Configuration

```java
public static class MyConfiguredInlineApp extends InlineApp {
    @Override
    protected int height() { return 3; }

    @Override
    protected Element render() { return text(""); }

    @Override
    protected InlineTuiConfig configure(int height) {
        return InlineTuiConfig.builder(height)
            .tickRate(Duration.ofMillis(30))
            .clearOnClose(false)
            .build();
    }
}
```

### Printing Elements

```java
println(row(
    text("OK ").green().fit(),
    text("lodash").fit(),
    text("@4.17.21").dim().fit()
).flex(Flex.START));

println("Step completed");
```

### Handling Key Events

```java
public static class MyInteractiveInlineApp extends InlineApp {
    @Override
    protected int height() { return 3; }

    @Override
    protected Element render() {
        return column(
            text("Continue? [Y/n]").cyan(),
            spacer()
        )
        .focusable()
        .onKeyEvent(event -> {
            if (event.character() == 'y' || event.character() == 'Y') {
                startNextPhase();
                return EventResult.HANDLED;
            } else if (event.character() == 'n') {
                quit();
                return EventResult.HANDLED;
            }
            return EventResult.UNHANDLED;
        });
    }

    void startNextPhase() {}
}
```

### Text Input

```java
public static class MyFormInlineApp extends InlineApp {
    private final TextInputState nameState = new TextInputState();

    @Override
    protected int height() { return 3; }

    @Override
    protected Element render() {
        return column(
            row(
                text("Name: ").bold().fit(),
                textInput(nameState)
                    .id("name-input")
                    .placeholder("Enter name...")
                    .constraint(Constraint.length(20))
                    .onSubmit(() -> handleSubmit(nameState.text()))
            ).flex(Flex.START),
            text("[Enter] Submit  [Tab] Next field").dim()
        );
    }

    private void handleSubmit(String value) {}
}
```

### Using InlineToolkitRunner Directly

```java
try (var runner = InlineToolkitRunner.create(3)) {
    runner.run(() -> column(
        waveText("Processing...").cyan(),
        gauge(progress),
        text("Please wait").dim()
    ));
}
```

### Scopes for Dynamic Regions

```java
public static class MyScopedInlineApp extends InlineApp {
    private boolean downloading = true;
    private double progress1 = 0;
    private double progress2 = 0;

    @Override
    protected int height() { return 5; }

    @Override
    protected Element render() {
        return column(
            text("Package Installation").bold(),
            scope(downloading,
                row(text("file1.zip: "), gauge(progress1)),
                row(text("file2.zip: "), gauge(progress2))
            ),
            text(downloading ? "Downloading..." : "Complete!").dim()
        );
    }
}
```

### Inline vs Full-Screen Toolkit

| Feature | InlineApp | App |
|---------|-----------|-----|
| Screen | Inline (preserves scroll) | Alternate (full takeover) |
| Height | Fixed | Dynamic |
| Output | `println()` scrolls above | No scrolling |
| Use | Progress, status, forms | Interactive apps |

---

## Scheduler Management

### External Scheduler Injection

```java
ScheduledExecutorService myScheduler =
    Executors.newSingleThreadScheduledExecutor();

var config = TuiConfig.builder()
    .scheduler(myScheduler)
    .build();

try (var tui = TuiRunner.create(config)) {
    tui.run(handler, renderer);
}

myScheduler.shutdown();
```

**Closing semantics:**
- Internally-created scheduler: shut down automatically
- External scheduler: NOT shut down (caller retains ownership)

### Multiple Runners with Shared Scheduler

Provide externally-managed scheduler to all runners to prevent premature shutdown of one runner affecting others.

---

## Choosing a Level

**Start with Toolkit DSL** for most applications -- clearest API, best productivity.

**Drop to TuiRunner** when needing custom event handling, animations, or direct widget control without Toolkit abstractions.

**Use Immediate Mode** for unusual scenarios -- custom backends, game engines, or complete understanding of terminal mechanics.

The levels compose well; mix as needed.

---

## Next Steps

- Bindings and Actions -- key bindings and action handling
- CSS Styling -- external stylesheets and theming
- Application Structure -- patterns for larger applications
- Widgets Reference -- all available widgets
- Developer Guide -- building custom widgets
