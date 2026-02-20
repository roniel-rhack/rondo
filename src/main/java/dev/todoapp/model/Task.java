package dev.todoapp.model;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

public class Task {
    private final String id;
    private String title;
    private String description;
    private TaskStatus status;
    private Priority priority;
    private final LocalDateTime createdAt;
    private LocalDate dueDate;
    private final List<SubTask> subtasks;
    private final List<String> tags;

    private Task(String title) {
        this.id = UUID.randomUUID().toString().substring(0, 8);
        this.title = title;
        this.status = TaskStatus.PENDING;
        this.priority = Priority.MEDIUM;
        this.createdAt = LocalDateTime.now();
        this.subtasks = new ArrayList<>();
        this.tags = new ArrayList<>();
    }

    public static Task create(String title) {
        return new Task(title);
    }

    public String id() { return id; }
    public String title() { return title; }
    public void setTitle(String title) { this.title = title; }
    public String description() { return description; }
    public void setDescription(String description) { this.description = description; }
    public TaskStatus status() { return status; }
    public void setStatus(TaskStatus status) { this.status = status; }
    public Priority priority() { return priority; }
    public void setPriority(Priority priority) { this.priority = priority; }
    public LocalDateTime createdAt() { return createdAt; }
    public LocalDate dueDate() { return dueDate; }
    public void setDueDate(LocalDate dueDate) { this.dueDate = dueDate; }
    public List<SubTask> subtasks() { return subtasks; }
    public List<String> tags() { return tags; }

    public void addSubTask(String title) {
        subtasks.add(new SubTask(title));
    }

    public void removeSubTask(String id) {
        subtasks.removeIf(s -> s.id().equals(id));
    }

    public void addTag(String tag) {
        if (!tags.contains(tag)) tags.add(tag);
    }

    public void removeTag(String tag) {
        tags.remove(tag);
    }

    public void cycleStatus() {
        this.status = this.status.next();
    }

    public double subtaskProgress() {
        if (subtasks.isEmpty()) return 0.0;
        long done = subtasks.stream().filter(SubTask::completed).count();
        return (double) done / subtasks.size();
    }

    public boolean isOverdue() {
        if (dueDate == null || status == TaskStatus.DONE) return false;
        return LocalDate.now().isAfter(dueDate);
    }

    public boolean matchesSearch(String query) {
        String q = query.toLowerCase();
        if (title.toLowerCase().contains(q)) return true;
        if (description != null && description.toLowerCase().contains(q)) return true;
        return tags.stream().anyMatch(t -> t.toLowerCase().contains(q));
    }
}
