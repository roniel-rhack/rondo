package dev.todoapp.model;

public enum Priority {
    LOW("Low", "↓"),
    MEDIUM("Medium", "→"),
    HIGH("High", "↑"),
    URGENT("Urgent", "⚡");

    private final String label;
    private final String symbol;

    Priority(String label, String symbol) {
        this.label = label;
        this.symbol = symbol;
    }

    public String label() { return label; }
    public String symbol() { return symbol; }
}
