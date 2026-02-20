package dev.todoapp.model;

public enum TaskStatus {
    PENDING("Pending", "○"),
    IN_PROGRESS("In Progress", "◐"),
    DONE("Done", "●");

    private final String label;
    private final String symbol;

    TaskStatus(String label, String symbol) {
        this.label = label;
        this.symbol = symbol;
    }

    public String label() { return label; }
    public String symbol() { return symbol; }

    public TaskStatus next() {
        TaskStatus[] values = values();
        return values[(ordinal() + 1) % values.length];
    }
}
