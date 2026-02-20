package dev.todoapp.controller;

import dev.todoapp.model.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import java.util.List;
import static org.junit.jupiter.api.Assertions.*;

class TaskControllerTest {

    private TaskController controller;

    @BeforeEach
    void setUp() {
        controller = new TaskController();
        controller.addTask("Task A");
        controller.addTask("Task B");
        controller.addTask("Task C");
    }

    @Test
    void addTask() {
        assertEquals(3, controller.allTasks().size());
        controller.addTask("Task D");
        assertEquals(4, controller.allTasks().size());
        assertEquals("Task D", controller.allTasks().get(3).title());
    }

    @Test
    void deleteSelectedTask() {
        controller.setSelectedIndex(1);
        controller.deleteSelected();
        assertEquals(2, controller.allTasks().size());
        assertEquals("Task A", controller.allTasks().get(0).title());
        assertEquals("Task C", controller.allTasks().get(1).title());
    }

    @Test
    void cycleStatusOfSelected() {
        controller.setSelectedIndex(0);
        assertEquals(TaskStatus.PENDING, controller.selectedTask().status());
        controller.cycleSelectedStatus();
        assertEquals(TaskStatus.IN_PROGRESS, controller.selectedTask().status());
    }

    @Test
    void filterByStatus() {
        controller.allTasks().get(0).setStatus(TaskStatus.IN_PROGRESS);
        controller.allTasks().get(1).setStatus(TaskStatus.DONE);
        assertEquals(1, controller.tasksByStatus(TaskStatus.IN_PROGRESS).size());
        assertEquals(1, controller.tasksByStatus(TaskStatus.DONE).size());
        assertEquals(1, controller.tasksByStatus(TaskStatus.PENDING).size());
    }

    @Test
    void searchTasks() {
        controller.setSearchQuery("task b");
        List<Task> results = controller.filteredTasks();
        assertEquals(1, results.size());
        assertEquals("Task B", results.get(0).title());
    }

    @Test
    void clearSearch() {
        controller.setSearchQuery("task b");
        assertEquals(1, controller.filteredTasks().size());
        controller.clearSearch();
        assertEquals(3, controller.filteredTasks().size());
    }

    @Test
    void selectedIndexClampsToValidRange() {
        controller.setSelectedIndex(99);
        assertEquals(2, controller.selectedIndex());
        controller.setSelectedIndex(-5);
        assertEquals(0, controller.selectedIndex());
    }

    @Test
    void selectedTaskReturnsNullWhenEmpty() {
        TaskController empty = new TaskController();
        assertNull(empty.selectedTask());
    }

    @Test
    void sortByCreationDate() {
        List<Task> sorted = controller.allTasks();
        assertTrue(sorted.get(0).createdAt().isBefore(sorted.get(2).createdAt())
                || sorted.get(0).createdAt().isEqual(sorted.get(2).createdAt()));
    }
}
