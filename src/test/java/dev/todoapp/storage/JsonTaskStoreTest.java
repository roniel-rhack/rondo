package dev.todoapp.storage;

import dev.todoapp.model.*;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import java.nio.file.Path;
import java.time.LocalDate;
import java.util.List;
import static org.junit.jupiter.api.Assertions.*;

class JsonTaskStoreTest {

    @TempDir
    Path tempDir;

    @Test
    void saveAndLoad() {
        Path file = tempDir.resolve("tasks.json");
        JsonTaskStore store = new JsonTaskStore(file);

        Task task = Task.create("Test task");
        task.setDescription("A description");
        task.setPriority(Priority.HIGH);
        task.setDueDate(LocalDate.of(2026, 3, 15));
        task.addSubTask("Sub 1");
        task.addTag("work");

        store.save(List.of(task));
        List<Task> loaded = store.load();

        assertEquals(1, loaded.size());
        Task t = loaded.get(0);
        assertEquals("Test task", t.title());
        assertEquals("A description", t.description());
        assertEquals(Priority.HIGH, t.priority());
        assertEquals(TaskStatus.PENDING, t.status());
        assertEquals(LocalDate.of(2026, 3, 15), t.dueDate());
        assertEquals(1, t.subtasks().size());
        assertEquals("Sub 1", t.subtasks().get(0).title());
        assertEquals(1, t.tags().size());
        assertEquals("work", t.tags().get(0));
    }

    @Test
    void loadReturnsEmptyListWhenFileDoesNotExist() {
        Path file = tempDir.resolve("nonexistent.json");
        JsonTaskStore store = new JsonTaskStore(file);
        List<Task> loaded = store.load();
        assertTrue(loaded.isEmpty());
    }

    @Test
    void createsSampleTasksOnFirstRun() {
        Path file = tempDir.resolve("tasks.json");
        JsonTaskStore store = new JsonTaskStore(file);
        List<Task> samples = store.loadOrCreateSamples();
        assertFalse(samples.isEmpty());
        assertTrue(samples.size() >= 3);
    }
}
