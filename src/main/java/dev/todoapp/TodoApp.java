package dev.todoapp;

import dev.tamboui.style.Color;
import dev.tamboui.text.Line;
import dev.tamboui.text.Span;
import dev.tamboui.text.Text;
import dev.tamboui.toolkit.app.ToolkitApp;
import dev.tamboui.toolkit.element.Element;
import dev.tamboui.toolkit.elements.ListElement;
import dev.tamboui.toolkit.elements.MarkupTextElement;
import dev.tamboui.toolkit.event.EventResult;
import dev.tamboui.tui.event.KeyCode;
import dev.tamboui.tui.event.KeyEvent;
import dev.tamboui.widgets.form.FieldType;
import dev.tamboui.widgets.form.FormState;
import dev.tamboui.widgets.form.Validators;
import dev.tamboui.widgets.input.TextInputState;
import dev.todoapp.controller.TaskController;
import dev.todoapp.model.*;
import dev.todoapp.storage.JsonTaskStore;

import java.nio.file.Path;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;

import static dev.tamboui.toolkit.Toolkit.*;

public class TodoApp extends ToolkitApp {

    private static final DateTimeFormatter DATE_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd");
    private static final DateTimeFormatter DATETIME_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm");

    private enum AppMode { NORMAL, SEARCH, ADD_TASK, EDIT_TASK, CONFIRM_DELETE, HELP, ADD_SUBTASK }

    private final TaskController controller = new TaskController();
    private final JsonTaskStore store;
    private final ListElement<?> taskList = list();
    private final TextInputState searchInputState = new TextInputState();
    private final TextInputState subtaskInputState = new TextInputState();
    private FormState formState;
    private AppMode mode = AppMode.NORMAL;

    public TodoApp() {
        Path dataDir = Path.of(System.getProperty("user.home"), ".todo-app");
        this.store = new JsonTaskStore(dataDir.resolve("tasks.json"));
        controller.setTasks(store.loadOrCreateSamples());
    }

    @Override
    protected Element render() {
        Element mainContent = dock()
                .top(renderHeader(), length(3))
                .center(row(renderTaskList(), renderTaskDetail()))
                .bottom(renderStatusBar(), length(1));

        return switch (mode) {
            case HELP -> stack(mainContent, renderHelpDialog());
            case CONFIRM_DELETE -> stack(mainContent, renderConfirmDialog());
            case ADD_TASK, EDIT_TASK -> stack(mainContent, renderInputDialog());
            case ADD_SUBTASK -> stack(mainContent, renderSubtaskDialog());
            default -> mainContent;
        };
    }

    // ─── Header ────────────────────────────────────────────

    private Element renderHeader() {
        String allLabel = " All (" + controller.totalCount() + ") ";
        String progressLabel = " In Progress (" + controller.inProgressCount() + ") ";
        String doneLabel = " Done (" + controller.doneCount() + ") ";

        int activeIndex;
        TaskStatus tab = controller.activeTab();
        if (tab == null) activeIndex = 0;
        else if (tab == TaskStatus.IN_PROGRESS) activeIndex = 1;
        else activeIndex = 2;

        Element tabsRow = tabs(allLabel, progressLabel, doneLabel)
                .selected(activeIndex)
                .highlightColor(Color.CYAN);

        if (mode == AppMode.SEARCH) {
            return column(
                    tabsRow,
                    panel(
                            textInput(searchInputState)
                                    .placeholder("Type to search...")
                                    .rounded()
                                    .borderColor(Color.GRAY)
                                    .focusedBorderColor(Color.CYAN)
                                    .showCursor(true)
                                    .cursorRequiresFocus(false)
                                    .focusable(false)
                    ).id("search-input").focusable().onKeyEvent(this::handleSearchEvent)
            );
        }

        return panel(tabsRow).id("header");
    }

    // ─── Task List ─────────────────────────────────────────

    private Element renderTaskList() {
        List<Task> visible = controller.filteredTasks();

        if (visible.isEmpty()) {
            String msg = controller.searchQuery().isEmpty()
                    ? "No tasks yet" : "No matching tasks";
            String hint = controller.searchQuery().isEmpty()
                    ? "Press [cyan]a[/] to create one" : "Try a different search";
            return panel("Tasks",
                    spacer(),
                    markupText("[dim]  " + msg + "[/]"),
                    markupText("  " + hint),
                    spacer()
            ).id("task-list").focusable().onKeyEvent(this::handleKeyEvent);
        }

        MarkupTextElement[] items = visible.stream()
                .map(this::renderTaskItem)
                .toArray(MarkupTextElement[]::new);

        taskList.elements(items)
                .selected(controller.selectedIndex())
                .highlightSymbol("▸ ")
                .highlightColor(Color.CYAN)
                .scrollbar()
                .autoScroll();

        return panel("Tasks", taskList)
                .id("task-list")
                .focusable()
                .onKeyEvent(this::handleKeyEvent);
    }

    private MarkupTextElement renderTaskItem(Task task) {
        String sc = colorName(task.status());
        String pc = colorName(task.priority());
        String title = esc(task.title());

        StringBuilder sb = new StringBuilder();
        sb.append("[").append(sc).append("]").append(task.status().symbol()).append("[/] ");
        sb.append("[").append(pc).append("]").append(task.priority().symbol()).append("[/] ");

        if (task.status() == TaskStatus.DONE) {
            sb.append("[dim crossed-out]").append(title).append("[/]");
        } else {
            sb.append(title);
        }

        if (!task.subtasks().isEmpty()) {
            long done = task.subtasks().stream().filter(SubTask::completed).count();
            sb.append(" [dim][[").append(done).append("/").append(task.subtasks().size()).append("]][/]");
        }

        if (task.isOverdue()) {
            sb.append(" [bold red]OVERDUE[/]");
        }

        return markupText(sb.toString());
    }

    // ─── Task Detail ───────────────────────────────────────

    private Element renderTaskDetail() {
        Task task = controller.selectedTask();

        if (task == null) {
            return panel("Details",
                    spacer(),
                    markupText("[dim]  Select a task[/]"),
                    markupText("[dim]  to view details[/]"),
                    spacer()
            ).id("task-detail");
        }

        String sc = colorName(task.status());
        String pc = colorName(task.priority());

        List<Element> content = new ArrayList<>();

        content.add(markupText("[bold " + sc + "]" + esc(task.title()) + "[/]"));
        content.add(text(""));

        content.add(markupText("[dim]Status    [/][" + sc + "]" + task.status().symbol() + " " + task.status().label() + "[/]"));
        content.add(markupText("[dim]Priority  [/][" + pc + "]" + task.priority().symbol() + " " + task.priority().label() + "[/]"));
        content.add(markupText("[dim]Created   [/][dim]" + task.createdAt().format(DATETIME_FMT) + "[/]"));

        if (task.dueDate() != null) {
            if (task.isOverdue()) {
                long days = ChronoUnit.DAYS.between(task.dueDate(), LocalDate.now());
                content.add(markupText("[dim]Due       [/][bold red]" + task.dueDate().format(DATE_FMT) + " (" + days + "d overdue)[/]"));
            } else {
                long days = ChronoUnit.DAYS.between(LocalDate.now(), task.dueDate());
                content.add(markupText("[dim]Due       [/][yellow]" + task.dueDate().format(DATE_FMT) + "[/] [dim](" + days + "d left)[/]"));
            }
        }

        if (!task.tags().isEmpty()) {
            content.add(markupText("[dim]Tags      [/][cyan]" + String.join("[/], [cyan]", task.tags()) + "[/]"));
        }

        if (task.description() != null && !task.description().isEmpty()) {
            content.add(text(""));
            content.add(markupText("[dim]─── Description ───────────────[/]"));
            content.add(text(task.description()));
        }

        if (!task.subtasks().isEmpty()) {
            long done = task.subtasks().stream().filter(SubTask::completed).count();
            double ratio = (double) done / task.subtasks().size();

            content.add(text(""));
            content.add(markupText("[dim]─── Subtasks (" + done + "/" + task.subtasks().size() + ") ─────────────[/]"));
            content.add(lineGauge(ratio)
                    .filledColor(Color.CYAN)
                    .label(Math.round(ratio * 100) + "%"));

            for (SubTask sub : task.subtasks()) {
                if (sub.completed()) {
                    content.add(markupText("  [green]✓[/] [dim crossed-out]" + esc(sub.title()) + "[/]"));
                } else {
                    content.add(markupText("  [yellow]○[/] " + esc(sub.title())));
                }
            }
        }

        content.add(spacer());
        return panel("Details", content.toArray(Element[]::new)).id("task-detail");
    }

    // ─── Status Bar ────────────────────────────────────────

    private Element renderStatusBar() {
        long total = controller.totalCount();
        long done = controller.doneCount();
        long progress = controller.inProgressCount();

        return row(
                richText(Text.from(Line.from(
                        Span.raw(" "),
                        Span.raw(String.valueOf(total)).bold(),
                        Span.raw(" tasks").dim(),
                        Span.raw(" │ ").dim(),
                        Span.raw(String.valueOf(progress)).fg(Color.CYAN).bold(),
                        Span.raw(" active").dim(),
                        Span.raw(" │ ").dim(),
                        Span.raw(String.valueOf(done)).fg(Color.GREEN).bold(),
                        Span.raw(" done").dim()
                ))),
                spacer(),
                richText(Text.from(Line.from(
                        Span.raw("a").fg(Color.CYAN), Span.raw(":add ").dim(),
                        Span.raw("e").fg(Color.CYAN), Span.raw(":edit ").dim(),
                        Span.raw("d").fg(Color.CYAN), Span.raw(":del ").dim(),
                        Span.raw("s").fg(Color.CYAN), Span.raw(":status ").dim(),
                        Span.raw("t").fg(Color.CYAN), Span.raw(":sub ").dim(),
                        Span.raw("/").fg(Color.CYAN), Span.raw(":find ").dim(),
                        Span.raw("?").fg(Color.CYAN), Span.raw(":help ").dim()
                )))
        ).id("status-bar");
    }

    // ─── Dialogs ───────────────────────────────────────────

    private Element renderHelpDialog() {
        return dialog("Keyboard Shortcuts",
                text(""),
                markupText("  [bold cyan]Navigation[/]"),
                markupText("  [cyan]j / ↓       [/]Move down"),
                markupText("  [cyan]k / ↑       [/]Move up"),
                markupText("  [cyan]Tab         [/]Switch tab"),
                text(""),
                markupText("  [bold cyan]Actions[/]"),
                markupText("  [cyan]a           [/]Add new task"),
                markupText("  [cyan]e           [/]Edit selected task"),
                markupText("  [cyan]d           [/]Delete task"),
                markupText("  [cyan]s           [/]Cycle task status"),
                markupText("  [cyan]t           [/]Add subtask"),
                markupText("  [cyan]x           [/]Toggle subtask"),
                markupText("  [cyan]/           [/]Search tasks"),
                text(""),
                markupText("  [bold cyan]Sorting[/]"),
                markupText("  [cyan]F1          [/]Sort by date"),
                markupText("  [cyan]F2          [/]Sort by priority"),
                markupText("  [cyan]F3          [/]Sort by due date"),
                text(""),
                markupText("  [cyan]Esc[/] Close   [cyan]q[/] Quit"),
                text("")
        ).width(50).padding(1).doubleBorder().borderColor(Color.CYAN)
                .focusable()
                .onKeyEvent(e -> { mode = AppMode.NORMAL; return EventResult.HANDLED; });
    }

    private Element renderConfirmDialog() {
        Task task = controller.selectedTask();
        String title = task != null ? esc(task.title()) : "this task";
        return dialog("Confirm Delete",
                text(""),
                markupText("  Delete [bold red]\"" + title + "\"[/] ?"),
                text(""),
                markupText("  [bold green]y[/] [dim]Yes[/]   [bold red]n / Esc[/] [dim]No[/]"),
                text("")
        ).width(50).padding(1).doubleBorder().borderColor(Color.RED)
                .focusable()
                .onKeyEvent(e -> {
                    if (e.isChar('y') || e.isChar('Y')) {
                        controller.deleteSelected();
                        saveAndRefresh();
                    }
                    mode = AppMode.NORMAL;
                    return EventResult.HANDLED;
                });
    }

    private Element renderInputDialog() {
        String title = mode == AppMode.ADD_TASK ? "New Task" : "Edit Task";
        return dialog(title,
                text(""),
                form(formState)
                        .field("title", "Title", Validators.required("Title is required"))
                        .field("description", "Description", FieldType.TEXT_AREA)
                        .field("priority", "Priority", FieldType.SELECT)
                        .field("dueDate", "Due Date", "yyyy-MM-dd")
                        .labelWidth(14)
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
                markupText("  [dim]Tab/↑↓[/]: fields  [dim]Enter[/]: save  [dim]Esc[/]: cancel")
        ).width(65).minWidth(50).padding(1).doubleBorder().borderColor(Color.CYAN)
                .focusable()
                .onKeyEvent(this::handleFormDialogEvent);
    }

    private FormState createFormState(String title, String description, int priorityIndex, String dueDate) {
        return FormState.builder()
                .textField("title", title)
                .textField("description", description)
                .selectField("priority", List.of("Low", "Medium", "High", "Urgent"), priorityIndex)
                .textField("dueDate", dueDate)
                .build();
    }

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

    private EventResult handleFormDialogEvent(KeyEvent event) {
        if (event.isCancel()) {
            mode = AppMode.NORMAL;
            formState = null;
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    }

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

    private EventResult handleSubtaskDialogEvent(KeyEvent event) {
        if (event.isCancel()) {
            mode = AppMode.NORMAL;
            subtaskInputState.clear();
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    }

    // ─── Event Handlers ────────────────────────────────────

    private EventResult handleKeyEvent(KeyEvent event) {
        if (mode != AppMode.NORMAL) return EventResult.UNHANDLED;

        if (event.isUp() || event.isChar('k')) {
            controller.moveUp();
            return EventResult.HANDLED;
        }
        if (event.isDown() || event.isChar('j')) {
            controller.moveDown();
            return EventResult.HANDLED;
        }

        if (event.isKey(KeyCode.TAB)) {
            controller.cycleTab();
            return EventResult.HANDLED;
        }

        if (event.isChar('a')) {
            formState = createFormState("", "", 1, "");
            mode = AppMode.ADD_TASK;
            return EventResult.HANDLED;
        }
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
        if (event.isChar('d')) {
            if (controller.selectedTask() != null) {
                mode = AppMode.CONFIRM_DELETE;
            }
            return EventResult.HANDLED;
        }
        if (event.isChar('s')) {
            controller.cycleSelectedStatus();
            saveAndRefresh();
            return EventResult.HANDLED;
        }
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
                        .ifPresent(st -> {
                            st.toggle();
                            saveAndRefresh();
                        });
            }
            return EventResult.HANDLED;
        }

        if (event.isChar('/')) {
            mode = AppMode.SEARCH;
            searchInputState.setText(controller.searchQuery());
            return EventResult.HANDLED;
        }

        if (event.isKey(KeyCode.F1)) {
            controller.sortByCreatedAt();
            return EventResult.HANDLED;
        }
        if (event.isKey(KeyCode.F2)) {
            controller.sortByPriority();
            return EventResult.HANDLED;
        }
        if (event.isKey(KeyCode.F3)) {
            controller.sortByDueDate();
            return EventResult.HANDLED;
        }

        if (event.isChar('?')) {
            mode = AppMode.HELP;
            return EventResult.HANDLED;
        }

        return EventResult.UNHANDLED;
    }

    private EventResult handleSearchEvent(KeyEvent event) {
        if (event.isCancel()) {
            controller.clearSearch();
            mode = AppMode.NORMAL;
            searchInputState.clear();
            return EventResult.HANDLED;
        }
        if (event.isConfirm()) {
            controller.setSearchQuery(searchInputState.text());
            mode = AppMode.NORMAL;
            searchInputState.clear();
            return EventResult.HANDLED;
        }
        if (handleTextInputKey(searchInputState, event)) {
            controller.setSearchQuery(searchInputState.text());
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    }

    // ─── Helpers ───────────────────────────────────────────

    private String colorName(TaskStatus status) {
        return switch (status) {
            case PENDING -> "yellow";
            case IN_PROGRESS -> "cyan";
            case DONE -> "green";
        };
    }

    private String colorName(Priority priority) {
        return switch (priority) {
            case LOW -> "green";
            case MEDIUM -> "yellow";
            case HIGH -> "red";
            case URGENT -> "bright-magenta";
        };
    }

    private String esc(String text) {
        return text.replace("[", "[[").replace("]", "]]");
    }

    private void saveAndRefresh() {
        store.save(controller.allTasks());
    }

    public static void main(String[] args) throws Exception {
        new TodoApp().run();
    }
}
