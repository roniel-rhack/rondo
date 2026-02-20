package dev.todoapp.storage;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonDeserializer;
import com.google.gson.JsonPrimitive;
import com.google.gson.JsonSerializer;
import com.google.gson.reflect.TypeToken;
import dev.todoapp.model.*;

import java.io.IOException;
import java.lang.reflect.Type;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

public class JsonTaskStore {
    private static final Type TASK_LIST_TYPE = new TypeToken<List<Task>>() {}.getType();
    private final Path filePath;
    private final Gson gson;

    public JsonTaskStore(Path filePath) {
        this.filePath = filePath;
        this.gson = new GsonBuilder()
                .setPrettyPrinting()
                .registerTypeAdapter(LocalDate.class,
                        (JsonSerializer<LocalDate>) (src, type, ctx) -> new JsonPrimitive(src.toString()))
                .registerTypeAdapter(LocalDate.class,
                        (JsonDeserializer<LocalDate>) (json, type, ctx) -> LocalDate.parse(json.getAsString()))
                .registerTypeAdapter(LocalDateTime.class,
                        (JsonSerializer<LocalDateTime>) (src, type, ctx) -> new JsonPrimitive(src.toString()))
                .registerTypeAdapter(LocalDateTime.class,
                        (JsonDeserializer<LocalDateTime>) (json, type, ctx) -> LocalDateTime.parse(json.getAsString()))
                .create();
    }

    public void save(List<Task> tasks) {
        try {
            Files.createDirectories(filePath.getParent());
            Files.writeString(filePath, gson.toJson(tasks));
        } catch (IOException e) {
            throw new RuntimeException("Failed to save tasks", e);
        }
    }

    public List<Task> load() {
        if (!Files.exists(filePath)) return new ArrayList<>();
        try {
            String json = Files.readString(filePath);
            List<Task> tasks = gson.fromJson(json, TASK_LIST_TYPE);
            return tasks != null ? new ArrayList<>(tasks) : new ArrayList<>();
        } catch (IOException e) {
            throw new RuntimeException("Failed to load tasks", e);
        }
    }

    public List<Task> loadOrCreateSamples() {
        List<Task> tasks = load();
        if (!tasks.isEmpty()) return tasks;

        tasks = createSampleTasks();
        save(tasks);
        return tasks;
    }

    private List<Task> createSampleTasks() {
        List<Task> samples = new ArrayList<>();

        Task t1 = Task.create("Learn TamboUI framework");
        t1.setDescription("Read the official docs and build a sample app");
        t1.setPriority(Priority.HIGH);
        t1.setStatus(TaskStatus.IN_PROGRESS);
        t1.addSubTask("Read getting started guide");
        t1.addSubTask("Try the Toolkit DSL");
        t1.addSubTask("Understand layout system");
        t1.addTag("learning");
        samples.add(t1);

        Task t2 = Task.create("Set up development environment");
        t2.setDescription("Install Java 25, Maven, and GraalVM");
        t2.setPriority(Priority.MEDIUM);
        t2.setStatus(TaskStatus.DONE);
        t2.addTag("setup");
        samples.add(t2);

        Task t3 = Task.create("Build todo app features");
        t3.setDescription("Implement CRUD, search, tabs, and persistence");
        t3.setPriority(Priority.HIGH);
        t3.setDueDate(LocalDate.now().plusDays(7));
        t3.addSubTask("Data model");
        t3.addSubTask("Persistence layer");
        t3.addSubTask("TUI views");
        t3.addSubTask("Key bindings");
        t3.addTag("feature");
        samples.add(t3);

        Task t4 = Task.create("Write project documentation");
        t4.setPriority(Priority.LOW);
        t4.addTag("docs");
        samples.add(t4);

        return samples;
    }
}
