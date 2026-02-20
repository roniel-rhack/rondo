package dev.todoapp.controller;

import dev.todoapp.model.*;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

public class TaskController {
    private final List<Task> tasks = new ArrayList<>();
    private int selectedIndex = 0;
    private String searchQuery = "";
    private TaskStatus activeTab = null;

    public List<Task> allTasks() { return tasks; }

    public void setTasks(List<Task> newTasks) {
        tasks.clear();
        tasks.addAll(newTasks);
        clampSelectedIndex();
    }

    public Task addTask(String title) {
        Task task = Task.create(title);
        tasks.add(task);
        return task;
    }

    public void deleteSelected() {
        List<Task> visible = filteredTasks();
        if (visible.isEmpty()) return;
        int idx = Math.min(selectedIndex, visible.size() - 1);
        Task toRemove = visible.get(idx);
        tasks.remove(toRemove);
        clampSelectedIndex();
    }

    public void cycleSelectedStatus() {
        Task task = selectedTask();
        if (task != null) task.cycleStatus();
    }

    public Task selectedTask() {
        List<Task> visible = filteredTasks();
        if (visible.isEmpty()) return null;
        int idx = Math.min(selectedIndex, visible.size() - 1);
        return visible.get(idx);
    }

    public int selectedIndex() { return selectedIndex; }

    public void setSelectedIndex(int index) {
        this.selectedIndex = index;
        clampSelectedIndex();
    }

    public void moveUp() { setSelectedIndex(selectedIndex - 1); }
    public void moveDown() { setSelectedIndex(selectedIndex + 1); }

    public List<Task> tasksByStatus(TaskStatus status) {
        return tasks.stream().filter(t -> t.status() == status).toList();
    }

    public List<Task> filteredTasks() {
        return tasks.stream()
                .filter(t -> activeTab == null || t.status() == activeTab)
                .filter(t -> searchQuery.isEmpty() || t.matchesSearch(searchQuery))
                .toList();
    }

    public String searchQuery() { return searchQuery; }

    public void setSearchQuery(String query) {
        this.searchQuery = query != null ? query : "";
        clampSelectedIndex();
    }

    public void clearSearch() { setSearchQuery(""); }

    public TaskStatus activeTab() { return activeTab; }

    public void setActiveTab(TaskStatus tab) {
        this.activeTab = tab;
        this.selectedIndex = 0;
    }

    public void cycleTab() {
        if (activeTab == null) activeTab = TaskStatus.IN_PROGRESS;
        else if (activeTab == TaskStatus.IN_PROGRESS) activeTab = TaskStatus.DONE;
        else activeTab = null;
        this.selectedIndex = 0;
    }

    public void sortByCreatedAt() {
        tasks.sort(Comparator.comparing(Task::createdAt));
    }

    public void sortByDueDate() {
        tasks.sort(Comparator.comparing(Task::dueDate, Comparator.nullsLast(Comparator.naturalOrder())));
    }

    public void sortByPriority() {
        tasks.sort(Comparator.comparing(Task::priority).reversed());
    }

    public int totalCount() { return tasks.size(); }
    public long doneCount() { return tasks.stream().filter(t -> t.status() == TaskStatus.DONE).count(); }
    public long inProgressCount() { return tasks.stream().filter(t -> t.status() == TaskStatus.IN_PROGRESS).count(); }

    private void clampSelectedIndex() {
        List<Task> visible = filteredTasks();
        if (visible.isEmpty()) {
            selectedIndex = 0;
        } else {
            selectedIndex = Math.max(0, Math.min(selectedIndex, visible.size() - 1));
        }
    }
}
