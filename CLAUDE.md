# Todo App TamboUI - Project Guide

## Project Overview

A modern, beautiful terminal user interface (TUI) task management application built with **Java** and the **TamboUI** framework (https://tamboui.dev/docs/main/index.html).

### Tech Stack
- **Language**: Java 25+
- **TUI Framework**: TamboUI 0.2.0-SNAPSHOT (Toolkit DSL - Level 3 API)
- **Build Tool**: Maven
- **Backend**: JLine 3 (`tamboui-jline`)
- **Styling**: TCSS (TamboUI CSS)
- **Native**: GraalVM native image compilation (target deliverable)

### Key Dependencies (`pom.xml`)
```xml
<repositories>
    <repository>
        <id>ossrh-snapshots</id>
        <url>https://central.sonatype.com/repository/maven-snapshots/</url>
        <releases><enabled>false</enabled></releases>
        <snapshots><enabled>true</enabled></snapshots>
    </repository>
</repositories>

<dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>dev.tamboui</groupId>
            <artifactId>tamboui-bom</artifactId>
            <version>0.2.0-SNAPSHOT</version>
            <type>pom</type>
            <scope>import</scope>
        </dependency>
    </dependencies>
</dependencyManagement>

<dependencies>
    <dependency>
        <groupId>dev.tamboui</groupId>
        <artifactId>tamboui-toolkit</artifactId>
    </dependency>
    <dependency>
        <groupId>dev.tamboui</groupId>
        <artifactId>tamboui-jline</artifactId>
    </dependency>
    <dependency>
        <groupId>dev.tamboui</groupId>
        <artifactId>tamboui-css</artifactId>
    </dependency>
</dependencies>

<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-compiler-plugin</artifactId>
            <version>3.13.0</version>
            <configuration>
                <source>25</source>
                <target>25</target>
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

<profiles>
    <profile>
        <id>native</id>
        <build>
            <plugins>
                <plugin>
                    <groupId>org.graalvm.buildtools</groupId>
                    <artifactId>native-maven-plugin</artifactId>
                    <version>0.10.4</version>
                    <extensions>true</extensions>
                    <executions>
                        <execution>
                            <id>build-native</id>
                            <goals><goal>compile-no-fork</goal></goals>
                            <phase>package</phase>
                        </execution>
                    </executions>
                    <configuration>
                        <mainClass>dev.todoapp.TodoApp</mainClass>
                        <imageName>todo-app</imageName>
                        <buildArgs>
                            <buildArg>--no-fallback</buildArg>
                        </buildArgs>
                    </configuration>
                </plugin>
            </plugins>
        </build>
    </profile>
</profiles>
```

---

## Application Features

### Core Functionality
- **Task Management**: Create, view, edit, and delete tasks
- **Subtask Support**: Tasks can have subtasks with independent state management
- **Status Tracking**: Toggle tasks between states (Pending, In Progress, Done)
- **Tab Navigation**: Separate tabs for "In Progress" and "Completed" tasks
- **Task Details**: Select a task to view its description and subtasks
- **Date Tracking**: Automatic creation date + optional due date
- **Sorting**: Sort by creation date, due date, or status
- **Search**: Filter/search tasks by title
- **Persistence**: Save tasks to JSON file for persistence across sessions

### Additional Features
- Priority levels (Low, Medium, High, Urgent)
- Task categories/tags for organization
- Keyboard-driven navigation (vim-style bindings)
- Status bar with task counts and current date
- Inline editing of task titles
- Confirmation dialogs for destructive actions (delete)
- Empty state with helpful prompts
- Auto-refresh on data changes

### UI Layout (Dock-based)
```
┌─────────────────── Header / Tabs ───────────────────┐
│  [In Progress]  [Completed]  [All]                   │
├──────────────────┬──────────────────────────────────┤
│  Task List       │  Task Detail                      │
│  (left panel)    │  (right panel)                    │
│                  │  - Title                           │
│  > Task 1        │  - Description                    │
│    Task 2        │  - Due Date                        │
│    Task 3        │  - Priority                        │
│                  │  - Subtasks []                     │
│                  │                                    │
├──────────────────┴──────────────────────────────────┤
│  Status Bar: 3 tasks | 1 done | [h]elp [a]dd [q]uit │
└─────────────────────────────────────────────────────┘
```

---

## Architecture

### MVC Pattern (TamboUI recommended)
- **Controller**: Plain Java classes managing task state, validation, and mutations
- **View**: Pure functions `Controller -> Element`, no side effects
- **Event Handling**: `@OnAction` annotations + `onKeyEvent` handlers returning `EventResult`

### Project Structure
```
src/
├── main/
│   ├── java/dev/todoapp/
│   │   ├── TodoApp.java              # Main entry point, ToolkitApp subclass
│   │   ├── controller/
│   │   │   ├── TaskController.java   # Task CRUD operations, state management
│   │   │   └── SearchController.java # Search/filter logic
│   │   ├── model/
│   │   │   ├── Task.java             # Task entity (title, desc, status, dates, priority)
│   │   │   ├── SubTask.java          # Subtask entity
│   │   │   ├── TaskStatus.java       # Enum: PENDING, IN_PROGRESS, DONE
│   │   │   └── Priority.java         # Enum: LOW, MEDIUM, HIGH, URGENT
│   │   ├── view/
│   │   │   ├── TaskListView.java     # Left panel - task list component
│   │   │   ├── TaskDetailView.java   # Right panel - task details component
│   │   │   ├── TaskFormView.java     # Add/edit task dialog
│   │   │   └── HeaderView.java       # Tab bar + search
│   │   └── storage/
│   │       └── JsonTaskStore.java    # JSON file persistence
│   └── resources/
│       ├── themes/
│       │   └── dark.tcss             # Dark theme stylesheet
│       └── bindings.properties       # Key bindings config
├── test/
│   └── java/dev/todoapp/
│       ├── controller/
│       │   └── TaskControllerTest.java
│       └── model/
│           └── TaskTest.java
```

### Data Model
```java
// Task.java
public class Task {
    private String id;           // UUID
    private String title;
    private String description;
    private TaskStatus status;   // PENDING, IN_PROGRESS, DONE
    private Priority priority;   // LOW, MEDIUM, HIGH, URGENT
    private LocalDateTime createdAt;    // auto-assigned
    private LocalDate dueDate;          // optional
    private List<SubTask> subtasks;
    private List<String> tags;
}

// SubTask.java
public class SubTask {
    private String id;
    private String title;
    private boolean completed;
}
```

---

## TamboUI Framework Reference

### API Level: Toolkit DSL (Level 3 - Recommended)
```java
public class TodoApp extends ToolkitApp {
    @Override
    protected Element render() {
        return dock()
            .top(headerView.render())
            .center(row(taskListView, taskDetailView))
            .bottom(statusBar())
            .topHeight(3)
            .bottomHeight(1);
    }
}
```

### Key Widgets Used
- **Tabs** + **TabsState**: Tab navigation between In Progress / Completed / All
- **ListWidget** + **ListState**: Task list with highlight and selection
- **Paragraph**: Task description display
- **TextInput** + **TextInputState**: Search bar, inline editing
- **Block**: Panels with borders (`.rounded()`)
- **Gauge**: Progress indicator for subtask completion
- **Checkbox** + **CheckboxState**: Subtask completion toggles
- **Stack**: Modal dialogs overlaid on main UI
- **Select** + **SelectState**: Priority and status dropdowns

### Layout System
- **Dock**: Main app layout (header, center, footer)
- **Row/Column**: Flex-based split panels
- **Grid**: Form layouts
- **Stack**: Modal overlays

### Styling (TCSS)
```tcss
$accent: cyan;
$bg-panel: black;
$border-focused: cyan;
$border-unfocused: gray;
$priority-high: red;
$priority-medium: yellow;
$priority-low: green;

Panel {
    border-type: rounded;
    border-color: $border-unfocused;
    &:focus { border-color: $border-focused; }
}

TabsElement-tab:selected {
    text-style: bold reversed;
    color: $accent;
}

ListElement-item:selected {
    background: $accent;
    color: black;
    text-style: bold;
}
```

### Event Handling
```java
// Vim-style + standard bindings
Bindings bindings = BindingSets.vim().toBuilder()
    .bind(KeyTrigger.ch('a'), "addTask")
    .bind(KeyTrigger.ch('d'), "deleteTask")
    .bind(KeyTrigger.ch('e'), "editTask")
    .bind(KeyTrigger.ch('/'), "search")
    .bind(KeyTrigger.ch('s'), "toggleStatus")
    .bind(KeyTrigger.key(KeyCode.TAB), "nextTab")
    .bind(KeyTrigger.ctrl('s'), "save")
    .build();

@OnAction("addTask")
void onAddTask(Event event) { /* show form */ }

@OnAction("deleteTask")
void onDeleteTask(Event event) { /* confirm + delete */ }
```

### Key Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Navigate task list |
| `Enter`/`Space` | Select/expand task |
| `a` | Add new task |
| `e` | Edit selected task |
| `d` | Delete selected task (with confirmation) |
| `s` | Cycle task status |
| `Tab` | Switch tabs |
| `/` | Focus search |
| `Escape` | Cancel/close dialog |
| `q` | Quit application |

---

## Workflow Orchestration

### 1. Plan Mode Default
- Enter plan mode for ANY non-trivial task (3+ steps or architectural decisions)
- If something goes sideways, STOP and re-plan immediately - don't keep pushing
- Use plan mode for verification steps, not just building
- Write detailed specs upfront to reduce ambiguity

### 2. Subagent Strategy
- Use subagents liberally to keep main context window clean
- Offload research, exploration, and parallel analysis to subagents
- For complex problems, throw more compute at it via subagents
- One task per subagent for focused execution

### 3. Self-Improvement Loop
- After ANY correction from the user: update `tasks/lessons.md` with the pattern
- Write rules for yourself that prevent the same mistake
- Ruthlessly iterate on these lessons until mistake rate drops
- Review lessons at session start for relevant project

### 4. Verification Before Done
- Never mark a task complete without proving it works
- Diff behavior between main and your changes when relevant
- Ask yourself: "Would a staff engineer approve this?"
- Run tests, check logs, demonstrate correctness

### 5. Demand Elegance (Balanced)
- For non-trivial changes: pause and ask "is there a more elegant way?"
- If a fix feels hacky: "Knowing everything I know now, implement the elegant solution"
- Skip this for simple, obvious fixes - don't over-engineer
- Challenge your own work before presenting it

### 6. Autonomous Bug Fixing
- When given a bug report: just fix it. Don't ask for hand-holding
- Point at logs, errors, failing tests - then resolve them
- Zero context switching required from the user
- Go fix failing CI tests without being told how

---

## Task Management

1. **Plan First**: Write plan to `tasks/todo.md` with checkable items
2. **Verify Plan**: Check in before starting implementation
3. **Track Progress**: Mark items complete as you go
4. **Explain Changes**: High-level summary at each step
5. **Document Results**: Add review section to `tasks/todo.md`
6. **Capture Lessons**: Update `tasks/lessons.md` after corrections

---

## Core Principles

- **Simplicity First**: Make every change as simple as possible. Impact minimal code.
- **No Laziness**: Find root causes. No temporary fixes. Senior developer standards.
- **Minimal Impact**: Changes should only touch what's necessary. Avoid introducing bugs.
- **Test Everything**: Unit tests for controllers and models. Verify TUI behavior manually.
- **Consistent Style**: Follow Java conventions, use the MVC pattern, keep views pure.
- **No AI Attribution**: Never reference Claude, AI, or any assistant in commit messages, code comments, or any project file. No `Co-Authored-By` headers.

---

## Build & Run

```bash
# Build
mvn clean compile

# Package
mvn package

# Run
mvn exec:java -Dexec.mainClass="dev.todoapp.TodoApp"
# or
java -jar target/todo-app-1.0-SNAPSHOT.jar

# Run tests
mvn test

# Native image (requires GraalVM)
mvn -Pnative native:compile
./target/todo-app

# Clean
mvn clean
```

---

## TamboUI Documentation Links
- Main docs: https://tamboui.dev/docs/main/index.html
- Getting started: https://tamboui.dev/docs/main/getting-started.html
- Widgets: https://tamboui.dev/docs/main/widgets.html
- Layouts: https://tamboui.dev/docs/main/layouts.html
- Styling: https://tamboui.dev/docs/main/styling.html
- Bindings: https://tamboui.dev/docs/main/bindings.html
- MVC: https://tamboui.dev/docs/main/mvc-architecture.html
- Forms: https://tamboui.dev/docs/main/forms.html
- GitHub: https://github.com/tamboui/tamboui
