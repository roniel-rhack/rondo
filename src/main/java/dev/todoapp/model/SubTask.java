package dev.todoapp.model;

import java.util.UUID;

public class SubTask {
    private final String id;
    private String title;
    private boolean completed;

    public SubTask(String title) {
        this.id = UUID.randomUUID().toString().substring(0, 8);
        this.title = title;
        this.completed = false;
    }

    public String id() { return id; }
    public String title() { return title; }
    public void setTitle(String title) { this.title = title; }
    public boolean completed() { return completed; }
    public void toggle() { this.completed = !this.completed; }
}
