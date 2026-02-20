package dev.todoapp.model;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class PriorityTest {

    @Test
    void hasFourValues() {
        assertEquals(4, Priority.values().length);
    }

    @Test
    void displaysLabel() {
        assertEquals("Low", Priority.LOW.label());
        assertEquals("Medium", Priority.MEDIUM.label());
        assertEquals("High", Priority.HIGH.label());
        assertEquals("Urgent", Priority.URGENT.label());
    }

    @Test
    void hasSymbol() {
        assertEquals("↓", Priority.LOW.symbol());
        assertEquals("→", Priority.MEDIUM.symbol());
        assertEquals("↑", Priority.HIGH.symbol());
        assertEquals("⚡", Priority.URGENT.symbol());
    }

    @Test
    void orderedBySeverity() {
        assertTrue(Priority.LOW.ordinal() < Priority.MEDIUM.ordinal());
        assertTrue(Priority.MEDIUM.ordinal() < Priority.HIGH.ordinal());
        assertTrue(Priority.HIGH.ordinal() < Priority.URGENT.ordinal());
    }
}
