package dev.todoapp.model;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class TaskStatusTest {

    @Test
    void hasThreeValues() {
        assertEquals(3, TaskStatus.values().length);
    }

    @Test
    void displaysHumanReadableLabel() {
        assertEquals("Pending", TaskStatus.PENDING.label());
        assertEquals("In Progress", TaskStatus.IN_PROGRESS.label());
        assertEquals("Done", TaskStatus.DONE.label());
    }

    @Test
    void cyclesForward() {
        assertEquals(TaskStatus.IN_PROGRESS, TaskStatus.PENDING.next());
        assertEquals(TaskStatus.DONE, TaskStatus.IN_PROGRESS.next());
        assertEquals(TaskStatus.PENDING, TaskStatus.DONE.next());
    }

    @Test
    void hasSymbol() {
        assertEquals("○", TaskStatus.PENDING.symbol());
        assertEquals("◐", TaskStatus.IN_PROGRESS.symbol());
        assertEquals("●", TaskStatus.DONE.symbol());
    }
}
