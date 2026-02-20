# TUI Redesign v2 - Full Task Forms & Subtask Management

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the single-field task input with a full multi-field form (title, description, priority, due date), fix the space-submits bug, and add subtask management from the UI.

**Architecture:** Use TamboUI's `FormState` + `FormElement` for the add/edit dialog — it handles arrow/tab navigation, text input, select dropdowns, and Enter-only submission natively. Add `ADD_SUBTASK` mode for a lightweight subtask dialog. Fix the search handler to use only `isConfirm()`.

**Tech Stack:** TamboUI 0.2.0-SNAPSHOT Toolkit DSL (`FormState`, `FormElement`, `SelectFieldState`, `Validators`), Java 25

---

### Task 1: Fix space-submits bug in search handler

**Files:**
- Modify: `src/main/java/dev/todoapp/TodoApp.java:409`

**Step 1: Fix the search event handler**

In `handleSearchEvent`, line 409, change:
```java
if (event.isSelect() || event.isConfirm()) {
```
to:
```java
if (event.isConfirm()) {
```

**Step 2: Compile and verify**

Run: `mvn compile`
Expected: BUILD SUCCESS

**Step 3: Commit**

```bash
git add src/main/java/dev/todoapp/TodoApp.java
git commit -m "fix: search handler no longer submits on Space"
```

---

### Task 2: Replace single-field input dialog with FormState-based multi-field form

**Files:**
- Modify: `src/main/java/dev/todoapp/TodoApp.java`

This is the main task. Replace the `TextInputState inputState` + manual `handleInputEvent` approach with a `FormState` + `FormElement`.

**Step 1: Replace state fields**

Remove:
```java
private final TextInputState inputState = new TextInputState();
```

Add:
```java
private FormState formState;
private final TextInputState subtaskInputState = new TextInputState();
```

Add new imports:
```java
import dev.tamboui.widgets.form.FormState;
import dev.tamboui.widgets.form.FieldType;
import dev.tamboui.widgets.form.Validators;
```

Add AppMode entry:
```java
private enum AppMode { NORMAL, SEARCH, ADD_TASK, EDIT_TASK, CONFIRM_DELETE, HELP, ADD_SUBTASK }
```

**Step 2: Create helper to build FormState**

```java
private FormState createFormState(String title, String description, int priorityIndex, String dueDate) {
    return FormState.builder()
            .textField("title", title)
            .textField("description", description)
            .selectField("priority", List.of("Low", "Medium", "High", "Urgent"), priorityIndex)
            .textField("dueDate", dueDate)
            .build();
}
```

**Step 3: Replace renderInputDialog()**

```java
private Element renderInputDialog() {
    String title = mode == AppMode.ADD_TASK ? "New Task" : "Edit Task";
    return dialog(title,
            text(""),
            form(formState)
                    .field("title", "Title", Validators.required("Title is required"))
                    .field("description", "Desc", FieldType.TEXT_AREA)
                    .field("priority", "Priority", FieldType.SELECT)
                    .field("dueDate", "Due Date", "yyyy-mm-dd")
                    .labelWidth(12)
                    .spacing(1)
                    .rounded()
                    .borderColor(Color.GRAY)
                    .focusedBorderColor(Color.CYAN)
                    .errorBorderColor(Color.RED)
                    .showInlineErrors(true)
                    .arrowNavigation(true)
                    .submitOnEnter(true)
                    .validateOnSubmit(true)
                    .onSubmit(this::onFormSubmit),
            text(""),
            markupText("  [dim]Tab[/]: next field  [dim]Enter[/]: save  [dim]Esc[/]: cancel")
    ).width(65).minWidth(50).padding(1).doubleBorder().borderColor(Color.CYAN)
            .focusable()
            .onKeyEvent(this::handleFormDialogEvent);
}
```

**Step 4: Add form submit handler**

```java
private void onFormSubmit(FormState state) {
    String title = state.textValue("title");
    if (title == null || title.isBlank()) return;

    String description = state.textValue("description");
    String priorityStr = state.selectValue("priority");
    String dueDateStr = state.textValue("dueDate");

    Priority priority = switch (priorityStr) {
        case "Low" -> Priority.LOW;
        case "High" -> Priority.HIGH;
        case "Urgent" -> Priority.URGENT;
        default -> Priority.MEDIUM;
    };

    LocalDate dueDate = null;
    if (dueDateStr != null && !dueDateStr.isBlank()) {
        try {
            dueDate = LocalDate.parse(dueDateStr.trim());
        } catch (Exception ignored) {}
    }

    if (mode == AppMode.ADD_TASK) {
        Task task = controller.addTask(title.trim());
        if (description != null && !description.isBlank()) task.setDescription(description.trim());
        task.setPriority(priority);
        if (dueDate != null) task.setDueDate(dueDate);
    } else if (mode == AppMode.EDIT_TASK) {
        Task task = controller.selectedTask();
        if (task != null) {
            task.setTitle(title.trim());
            task.setDescription(description != null ? description.trim() : null);
            task.setPriority(priority);
            task.setDueDate(dueDate);
        }
    }

    mode = AppMode.NORMAL;
    formState = null;
    saveAndRefresh();
}
```

**Step 5: Add form dialog key handler (for Esc only)**

```java
private EventResult handleFormDialogEvent(KeyEvent event) {
    if (event.isCancel()) {
        mode = AppMode.NORMAL;
        formState = null;
        return EventResult.HANDLED;
    }
    return EventResult.UNHANDLED;
}
```

**Step 6: Update 'a' and 'e' key handlers to create FormState**

In `handleKeyEvent`, change the `'a'` handler:
```java
if (event.isChar('a')) {
    formState = createFormState("", "", 1, "");
    mode = AppMode.ADD_TASK;
    return EventResult.HANDLED;
}
```

Change the `'e'` handler:
```java
if (event.isChar('e')) {
    Task selected = controller.selectedTask();
    if (selected != null) {
        int priorityIndex = selected.priority().ordinal();
        String dueStr = selected.dueDate() != null ? selected.dueDate().format(DATE_FMT) : "";
        String desc = selected.description() != null ? selected.description() : "";
        formState = createFormState(selected.title(), desc, priorityIndex, dueStr);
        mode = AppMode.EDIT_TASK;
    }
    return EventResult.HANDLED;
}
```

**Step 7: Update render() switch for ADD_SUBTASK**

```java
return switch (mode) {
    case HELP -> stack(mainContent, renderHelpDialog());
    case CONFIRM_DELETE -> stack(mainContent, renderConfirmDialog());
    case ADD_TASK, EDIT_TASK -> stack(mainContent, renderInputDialog());
    case ADD_SUBTASK -> stack(mainContent, renderSubtaskDialog());
    default -> mainContent;
};
```

**Step 8: Update search handler to use a separate TextInputState**

The search still needs a `TextInputState`. Keep a dedicated one:
```java
private final TextInputState searchInputState = new TextInputState();
```

Update all search references from `inputState` to `searchInputState`.

**Step 9: Remove old handleInputEvent method**

Delete the `handleInputEvent` method entirely — `FormElement.onSubmit()` handles submission now.

**Step 10: Compile and verify**

Run: `mvn compile`
Expected: BUILD SUCCESS

**Step 11: Commit**

```bash
git add src/main/java/dev/todoapp/TodoApp.java
git commit -m "feat: replace single-field input with full multi-field form"
```

---

### Task 3: Add subtask management

**Files:**
- Modify: `src/main/java/dev/todoapp/TodoApp.java`

**Step 1: Add subtask dialog renderer**

```java
private Element renderSubtaskDialog() {
    return dialog("Add Subtask",
            text(""),
            formField("Subtask", subtaskInputState)
                    .placeholder("Enter subtask title...")
                    .labelWidth(12)
                    .rounded()
                    .borderColor(Color.GRAY)
                    .focusedBorderColor(Color.CYAN)
                    .onSubmit(this::onSubtaskSubmit)
                    .arrowNavigation(false),
            text(""),
            markupText("  [dim]Enter[/]: add  [dim]Esc[/]: cancel")
    ).width(60).padding(1).doubleBorder().borderColor(Color.CYAN)
            .focusable()
            .onKeyEvent(this::handleSubtaskDialogEvent);
}
```

**Step 2: Add subtask submit handler**

```java
private void onSubtaskSubmit() {
    String title = subtaskInputState.text();
    if (title != null && !title.isBlank()) {
        Task task = controller.selectedTask();
        if (task != null) {
            task.addSubTask(title.trim());
            saveAndRefresh();
        }
    }
    mode = AppMode.NORMAL;
    subtaskInputState.clear();
}
```

**Step 3: Add subtask dialog key handler**

```java
private EventResult handleSubtaskDialogEvent(KeyEvent event) {
    if (event.isCancel()) {
        mode = AppMode.NORMAL;
        subtaskInputState.clear();
        return EventResult.HANDLED;
    }
    return EventResult.UNHANDLED;
}
```

**Step 4: Add 't' and 'x' keybindings in handleKeyEvent**

After the `'s'` handler:
```java
if (event.isChar('t')) {
    if (controller.selectedTask() != null) {
        subtaskInputState.clear();
        mode = AppMode.ADD_SUBTASK;
    }
    return EventResult.HANDLED;
}
if (event.isChar('x')) {
    Task task = controller.selectedTask();
    if (task != null && !task.subtasks().isEmpty()) {
        task.subtasks().stream()
                .filter(st -> !st.completed())
                .findFirst()
                .ifPresentOrElse(
                        st -> st.toggle(),
                        () -> task.subtasks().getFirst().toggle()
                );
        saveAndRefresh();
    }
    return EventResult.HANDLED;
}
```

**Step 5: Update help dialog with new keybindings**

Add after the "Cycle task status" line:
```java
markupText("  [cyan]t           [/]Add subtask"),
markupText("  [cyan]x           [/]Toggle subtask"),
```

**Step 6: Update status bar with new key hints**

Add to the status bar richText:
```java
Span.raw("t").fg(Color.CYAN), Span.raw(":sub ").dim(),
```

**Step 7: Compile and verify**

Run: `mvn compile`
Expected: BUILD SUCCESS

**Step 8: Commit**

```bash
git add src/main/java/dev/todoapp/TodoApp.java
git commit -m "feat: add subtask management with t to add and x to toggle"
```

---

### Task 4: Update TCSS theme for form styling

**Files:**
- Modify: `src/main/resources/themes/dark.tcss`

**Step 1: Add form-specific styles**

Append to dark.tcss:
```css
FormElement {
    spacing: 1;
}

FormFieldElement-label {
    color: gray;
    text-style: dim;
}

FormFieldElement-error {
    color: red;
    text-style: italic;
}

Dialog {
    border-type: double;
    border-color: cyan;
    background: black;
}
```

**Step 2: Commit**

```bash
git add src/main/resources/themes/dark.tcss
git commit -m "style: add form and dialog TCSS styles"
```

---

### Task 5: Rebuild native image and verify

**Files:**
- No new files

**Step 1: Run tests**

Run: `mvn test`
Expected: 28 tests pass

**Step 2: Build native image**

Run: `mvn -Pnative package -DskipTests`
Expected: BUILD SUCCESS, `target/todo-app` binary generated

**Step 3: Smoke test**

Run: `./target/todo-app` in a real terminal
Verify:
- Main UI renders with colors
- Press `a` -> full form dialog appears with Title, Desc, Priority, Due Date
- Tab moves between fields
- Priority field cycles with left/right arrows
- Enter submits (Space does NOT submit)
- Esc cancels
- Press `e` on a task -> form pre-fills with task data
- Press `t` -> subtask dialog appears
- Press `x` -> toggles next incomplete subtask
- Press `/` -> search works, Space doesn't dismiss
- All existing features still work

**Step 4: Commit**

```bash
git add -A
git commit -m "chore: rebuild native image with form-based task dialogs"
```
