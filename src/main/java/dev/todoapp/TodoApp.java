package dev.todoapp;

import dev.tamboui.style.Color;
import dev.tamboui.toolkit.app.ToolkitApp;
import dev.tamboui.toolkit.element.Element;
import dev.tamboui.toolkit.elements.ListElement;
import dev.tamboui.toolkit.event.EventResult;
import dev.tamboui.tui.event.KeyCode;
import dev.tamboui.tui.event.KeyEvent;
import dev.todoapp.controller.TaskController;
import dev.todoapp.model.*;
import dev.todoapp.storage.JsonTaskStore;

import java.nio.file.Path;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;

import static dev.tamboui.toolkit.Toolkit.*;

public class TodoApp extends ToolkitApp {

    private static final DateTimeFormatter DATE_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd");
    private static final DateTimeFormatter DATETIME_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm");

    private final TaskController controller = new TaskController();
    private final JsonTaskStore store;
    private final ListElement<?> taskList = list().highlightColor(Color.CYAN).autoScroll();

    private boolean searchMode = false;
    private boolean showHelp = false;
    private boolean confirmDelete = false;
    private boolean addMode = false;
    private boolean editMode = false;
    private String inputBuffer = "";

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

        if (showHelp) {
            return stack(mainContent, renderHelpDialog());
        }
        if (confirmDelete) {
            return stack(mainContent, renderConfirmDialog());
        }
        if (addMode || editMode) {
            return stack(mainContent, renderInputDialog());
        }

        return mainContent;
    }

    private Element renderHeader() {
        String allLabel = "All (" + controller.totalCount() + ")";
        String progressLabel = "In Progress (" + controller.inProgressCount() + ")";
        String doneLabel = "Done (" + controller.doneCount() + ")";

        int activeIndex;
        TaskStatus tab = controller.activeTab();
        if (tab == null) activeIndex = 0;
        else if (tab == TaskStatus.IN_PROGRESS) activeIndex = 1;
        else activeIndex = 2;

        Element tabsRow = tabs(allLabel, progressLabel, doneLabel)
                .selected(activeIndex)
                .highlightColor(Color.CYAN);

        if (searchMode) {
            return column(
                    tabsRow,
                    panel(text("[/] " + inputBuffer + "█").cyan())
                            .id("search-input")
                            .focusable()
                            .onKeyEvent(this::handleSearchEvent)
            );
        }

        return panel(tabsRow).id("header");
    }

    private Element renderTaskList() {
        List<Task> visible = controller.filteredTasks();

        if (visible.isEmpty()) {
            return panel("Tasks",
                    spacer(),
                    text("No tasks found").dim(),
                    text(controller.searchQuery().isEmpty()
                            ? "Press 'a' to add a task"
                            : "Try a different search").dim(),
                    spacer()
            ).id("task-list").focusable().onKeyEvent(this::handleKeyEvent);
        }

        String[] items = visible.stream()
                .map(this::formatTaskTitle)
                .toArray(String[]::new);

        return panel("Tasks",
                taskList.items(items).selected(controller.selectedIndex())
        ).id("task-list")
                .focusable()
                .onKeyEvent(this::handleKeyEvent);
    }

    private String formatTaskTitle(Task task) {
        String status = task.status().symbol();
        String priority = task.priority().symbol();
        String overdue = task.isOverdue() ? " !" : "";
        String subtaskInfo = task.subtasks().isEmpty() ? ""
                : " [" + task.subtasks().stream().filter(SubTask::completed).count()
                + "/" + task.subtasks().size() + "]";
        return status + " " + priority + " " + task.title() + subtaskInfo + overdue;
    }

    private Element renderTaskDetail() {
        Task task = controller.selectedTask();

        if (task == null) {
            return panel("Details",
                    spacer(),
                    text("Select a task to view details").dim(),
                    spacer()
            ).id("task-detail");
        }

        List<Element> content = new ArrayList<>();

        content.add(text(task.title()).bold().cyan());
        content.add(text(""));

        content.add(text("Status:   " + task.status().symbol() + " " + task.status().label()));
        content.add(text("Priority: " + task.priority().symbol() + " " + task.priority().label()));

        content.add(text("Created:  " + task.createdAt().format(DATETIME_FMT)).dim());
        if (task.dueDate() != null) {
            String dueText = "Due:      " + task.dueDate().format(DATE_FMT);
            if (task.isOverdue()) {
                content.add(text(dueText + " (OVERDUE)").fg(Color.RED).bold());
            } else {
                content.add(text(dueText));
            }
        }

        if (!task.tags().isEmpty()) {
            content.add(text("Tags:     " + String.join(", ", task.tags())).dim());
        }

        if (task.description() != null && !task.description().isEmpty()) {
            content.add(text(""));
            content.add(text("Description").bold());
            content.add(text(task.description()));
        }

        if (!task.subtasks().isEmpty()) {
            content.add(text(""));
            long done = task.subtasks().stream().filter(SubTask::completed).count();
            content.add(text("Subtasks (" + done + "/" + task.subtasks().size() + ")").bold());
            for (SubTask sub : task.subtasks()) {
                String check = sub.completed() ? "✓" : "○";
                Color color = sub.completed() ? Color.GREEN : Color.WHITE;
                content.add(text("  " + check + " " + sub.title()).fg(color));
            }
        }

        content.add(spacer());

        return panel("Details", content.toArray(Element[]::new)).id("task-detail");
    }

    private Element renderStatusBar() {
        long total = controller.totalCount();
        long done = controller.doneCount();
        long progress = controller.inProgressCount();
        String today = LocalDate.now().format(DATE_FMT);

        String left = total + " tasks | " + progress + " active | " + done + " done | " + today;
        String right = "[a]dd [e]dit [d]el [s]tatus [/]search [Tab]tab [?]help [q]uit";

        return row(
                text(" " + left).dim(),
                spacer(),
                text(right + " ").dim()
        ).id("status-bar");
    }

    private Element renderHelpDialog() {
        return dialog("Keyboard Shortcuts",
                text(""),
                text("  j / ↓       Move down").cyan(),
                text("  k / ↑       Move up").cyan(),
                text("  Enter       Select / toggle subtask"),
                text("  a           Add new task"),
                text("  e           Edit selected task"),
                text("  d           Delete selected task"),
                text("  s           Cycle task status"),
                text("  Tab         Switch tab"),
                text("  /           Search tasks"),
                text("  F1          Sort by date"),
                text("  F2          Sort by priority"),
                text("  F3          Sort by due date"),
                text("  Escape      Close dialog / cancel"),
                text("  q           Quit"),
                text(""),
                text("  Press any key to close").dim()
        ).focusable()
                .onKeyEvent(e -> { showHelp = false; return EventResult.HANDLED; });
    }

    private Element renderConfirmDialog() {
        Task task = controller.selectedTask();
        String title = task != null ? task.title() : "this task";
        return dialog("Confirm Delete",
                text(""),
                text("  Delete \"" + title + "\"?").bold(),
                text(""),
                text("  [y] Yes   [n/Esc] No").dim(),
                text("")
        ).focusable()
                .onKeyEvent(e -> {
                    if (e.isChar('y') || e.isChar('Y')) {
                        controller.deleteSelected();
                        saveAndRefresh();
                    }
                    confirmDelete = false;
                    return EventResult.HANDLED;
                });
    }

    private Element renderInputDialog() {
        String title = addMode ? "New Task" : "Edit Task";
        return dialog(title,
                text(""),
                text("  Title:").bold(),
                text("  " + inputBuffer + "█").cyan(),
                text(""),
                text("  [Enter] Confirm   [Esc] Cancel").dim(),
                text("")
        ).focusable()
                .onKeyEvent(this::handleInputEvent);
    }

    private EventResult handleKeyEvent(KeyEvent event) {
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
            addMode = true;
            inputBuffer = "";
            return EventResult.HANDLED;
        }
        if (event.isChar('e')) {
            Task selected = controller.selectedTask();
            if (selected != null) {
                editMode = true;
                inputBuffer = selected.title();
            }
            return EventResult.HANDLED;
        }
        if (event.isChar('d')) {
            if (controller.selectedTask() != null) {
                confirmDelete = true;
            }
            return EventResult.HANDLED;
        }
        if (event.isChar('s')) {
            controller.cycleSelectedStatus();
            saveAndRefresh();
            return EventResult.HANDLED;
        }

        if (event.isChar('/')) {
            searchMode = true;
            inputBuffer = controller.searchQuery();
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
            showHelp = true;
            return EventResult.HANDLED;
        }

        return EventResult.UNHANDLED;
    }

    private EventResult handleSearchEvent(KeyEvent event) {
        if (event.isCancel()) {
            searchMode = false;
            if (inputBuffer.isEmpty()) {
                controller.clearSearch();
            }
            inputBuffer = "";
            return EventResult.HANDLED;
        }
        if (event.isSelect() || event.isConfirm()) {
            searchMode = false;
            inputBuffer = "";
            return EventResult.HANDLED;
        }
        if (event.isDeleteBackward()) {
            if (!inputBuffer.isEmpty()) {
                inputBuffer = inputBuffer.substring(0, inputBuffer.length() - 1);
            }
            controller.setSearchQuery(inputBuffer);
            return EventResult.HANDLED;
        }
        char ch = event.character();
        if (ch >= 32 && ch < 127) {
            inputBuffer += ch;
            controller.setSearchQuery(inputBuffer);
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    }

    private EventResult handleInputEvent(KeyEvent event) {
        if (event.isCancel()) {
            addMode = false;
            editMode = false;
            inputBuffer = "";
            return EventResult.HANDLED;
        }
        if ((event.isSelect() || event.isConfirm()) && !inputBuffer.isBlank()) {
            if (addMode) {
                controller.addTask(inputBuffer.trim());
            } else if (editMode) {
                Task task = controller.selectedTask();
                if (task != null) task.setTitle(inputBuffer.trim());
            }
            addMode = false;
            editMode = false;
            inputBuffer = "";
            saveAndRefresh();
            return EventResult.HANDLED;
        }
        if (event.isDeleteBackward()) {
            if (!inputBuffer.isEmpty()) {
                inputBuffer = inputBuffer.substring(0, inputBuffer.length() - 1);
            }
            return EventResult.HANDLED;
        }
        char ch = event.character();
        if (ch >= 32 && ch < 127) {
            inputBuffer += ch;
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    }

    private void saveAndRefresh() {
        store.save(controller.allTasks());
    }

    public static void main(String[] args) throws Exception {
        new TodoApp().run();
    }
}
