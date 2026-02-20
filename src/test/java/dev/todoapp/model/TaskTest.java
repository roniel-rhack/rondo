package dev.todoapp.model;

import org.junit.jupiter.api.Test;
import java.time.LocalDate;
import static org.junit.jupiter.api.Assertions.*;

class TaskTest {

    @Test
    void createsWithDefaults() {
        Task task = Task.create("Buy groceries");
        assertNotNull(task.id());
        assertEquals("Buy groceries", task.title());
        assertNull(task.description());
        assertEquals(TaskStatus.PENDING, task.status());
        assertEquals(Priority.MEDIUM, task.priority());
        assertNotNull(task.createdAt());
        assertNull(task.dueDate());
        assertTrue(task.subtasks().isEmpty());
        assertTrue(task.tags().isEmpty());
    }

    @Test
    void addSubTask() {
        Task task = Task.create("Project");
        task.addSubTask("Step 1");
        task.addSubTask("Step 2");
        assertEquals(2, task.subtasks().size());
        assertEquals("Step 1", task.subtasks().get(0).title());
        assertFalse(task.subtasks().get(0).completed());
    }

    @Test
    void toggleSubTask() {
        Task task = Task.create("Project");
        task.addSubTask("Step 1");
        SubTask sub = task.subtasks().get(0);
        assertFalse(sub.completed());
        sub.toggle();
        assertTrue(sub.completed());
        sub.toggle();
        assertFalse(sub.completed());
    }

    @Test
    void subtaskProgress() {
        Task task = Task.create("Project");
        task.addSubTask("Step 1");
        task.addSubTask("Step 2");
        task.addSubTask("Step 3");
        assertEquals(0.0, task.subtaskProgress(), 0.01);
        task.subtasks().get(0).toggle();
        assertEquals(1.0 / 3.0, task.subtaskProgress(), 0.01);
        task.subtasks().get(1).toggle();
        task.subtasks().get(2).toggle();
        assertEquals(1.0, task.subtaskProgress(), 0.01);
    }

    @Test
    void subtaskProgressWithNoSubtasks() {
        Task task = Task.create("Simple");
        assertEquals(0.0, task.subtaskProgress(), 0.01);
    }

    @Test
    void cycleStatus() {
        Task task = Task.create("Test");
        assertEquals(TaskStatus.PENDING, task.status());
        task.cycleStatus();
        assertEquals(TaskStatus.IN_PROGRESS, task.status());
        task.cycleStatus();
        assertEquals(TaskStatus.DONE, task.status());
        task.cycleStatus();
        assertEquals(TaskStatus.PENDING, task.status());
    }

    @Test
    void isOverdue() {
        Task task = Task.create("Overdue");
        assertFalse(task.isOverdue());
        task.setDueDate(LocalDate.now().minusDays(1));
        assertTrue(task.isOverdue());
        task.setStatus(TaskStatus.DONE);
        assertFalse(task.isOverdue());
    }

    @Test
    void matchesSearch() {
        Task task = Task.create("Buy groceries");
        task.setDescription("Get milk and eggs");
        task.addTag("shopping");
        assertTrue(task.matchesSearch("buy"));
        assertTrue(task.matchesSearch("GROCERIES"));
        assertTrue(task.matchesSearch("milk"));
        assertTrue(task.matchesSearch("shopping"));
        assertFalse(task.matchesSearch("workout"));
    }
}
