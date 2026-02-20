# Todo App TamboUI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a modern, beautiful TUI task manager in Java using TamboUI with tabs, subtasks, search, persistence, and GraalVM native compilation.

**Architecture:** MVC pattern following TamboUI's recommended Toolkit DSL (Level 3 API). Controller holds all state and mutations, views are pure render functions, events dispatch through `@OnAction` and `onKeyEvent`. JSON persistence via Gson.

**Tech Stack:** Java 25, TamboUI 0.2.0-SNAPSHOT (Toolkit DSL), Maven, Gson, JUnit 5, GraalVM native image

---

## Task 1: Maven Project Setup

**Files:**
- Create: `pom.xml`
- Create: `src/main/java/dev/todoapp/TodoApp.java`

**Step 1: Create pom.xml**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>dev.todoapp</groupId>
    <artifactId>todo-app</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>jar</packaging>

    <name>Todo App TamboUI</name>
    <description>TUI task manager built with TamboUI</description>

    <properties>
        <maven.compiler.source>25</maven.compiler.source>
        <maven.compiler.target>25</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <tamboui.version>0.2.0-SNAPSHOT</tamboui.version>
        <gson.version>2.11.0</gson.version>
        <junit.version>5.11.4</junit.version>
    </properties>

    <repositories>
        <repository>
            <id>ossrh-snapshots</id>
            <url>https://central.sonatype.com/repository/maven-snapshots/</url>
            <releases><enabled>false</enabled></releases>
            <snapshots><enabled>true</enabled></snapshots>
        </repository>
    </repositories>

    <dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>dev.tamboui</groupId>
                <artifactId>tamboui-bom</artifactId>
                <version>${tamboui.version}</version>
                <type>pom</type>
                <scope>import</scope>
            </dependency>
        </dependencies>
    </dependencyManagement>

    <dependencies>
        <dependency>
            <groupId>dev.tamboui</groupId>
            <artifactId>tamboui-toolkit</artifactId>
        </dependency>
        <dependency>
            <groupId>dev.tamboui</groupId>
            <artifactId>tamboui-jline</artifactId>
        </dependency>
        <dependency>
            <groupId>dev.tamboui</groupId>
            <artifactId>tamboui-css</artifactId>
        </dependency>
        <dependency>
            <groupId>com.google.code.gson</groupId>
            <artifactId>gson</artifactId>
            <version>${gson.version}</version>
        </dependency>
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>${junit.version}</version>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-compiler-plugin</artifactId>
                <version>3.13.0</version>
                <configuration>
                    <source>25</source>
                    <target>25</target>
                    <annotationProcessorPaths>
                        <path>
                            <groupId>dev.tamboui</groupId>
                            <artifactId>tamboui-processor</artifactId>
                            <version>${tamboui.version}</version>
                        </path>
                    </annotationProcessorPaths>
                </configuration>
            </plugin>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-jar-plugin</artifactId>
                <version>3.4.2</version>
                <configuration>
                    <archive>
                        <manifest>
                            <mainClass>dev.todoapp.TodoApp</mainClass>
                        </manifest>
                    </archive>
                </configuration>
            </plugin>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-shade-plugin</artifactId>
                <version>3.6.0</version>
                <executions>
                    <execution>
                        <phase>package</phase>
                        <goals><goal>shade</goal></goals>
                        <configuration>
                            <transformers>
                                <transformer implementation="org.apache.maven.plugins.shade.resource.ManifestResourceTransformer">
                                    <mainClass>dev.todoapp.TodoApp</mainClass>
                                </transformer>
                                <transformer implementation="org.apache.maven.plugins.shade.resource.ServicesResourceTransformer"/>
                            </transformers>
                        </configuration>
                    </execution>
                </executions>
            </plugin>
            <plugin>
                <groupId>org.codehaus.mojo</groupId>
                <artifactId>exec-maven-plugin</artifactId>
                <version>3.5.0</version>
                <configuration>
                    <mainClass>dev.todoapp.TodoApp</mainClass>
                </configuration>
            </plugin>
        </plugins>
    </build>

    <profiles>
        <profile>
            <id>native</id>
            <build>
                <plugins>
                    <plugin>
                        <groupId>org.graalvm.buildtools</groupId>
                        <artifactId>native-maven-plugin</artifactId>
                        <version>0.10.4</version>
                        <extensions>true</extensions>
                        <executions>
                            <execution>
                                <id>build-native</id>
                                <goals><goal>compile-no-fork</goal></goals>
                                <phase>package</phase>
                            </execution>
                        </executions>
                        <configuration>
                            <mainClass>dev.todoapp.TodoApp</mainClass>
                            <imageName>todo-app</imageName>
                            <buildArgs>
                                <buildArg>--no-fallback</buildArg>
                            </buildArgs>
                        </configuration>
                    </plugin>
                </plugins>
            </build>
        </profile>
    </profiles>
</project>
```

**Step 2: Create minimal TodoApp.java**

```java
package dev.todoapp;

import dev.tamboui.toolkit.ToolkitApp;
import dev.tamboui.toolkit.Element;
import static dev.tamboui.toolkit.Toolkit.*;

public class TodoApp extends ToolkitApp {

    @Override
    protected Element render() {
        return panel("Todo App",
            text("Welcome to Todo App!").bold().cyan(),
            spacer(),
            text("Press 'q' to quit").dim()
        ).rounded();
    }

    public static void main(String[] args) throws Exception {
        new TodoApp().run();
    }
}
```

**Step 3: Verify build compiles**

Run: `mvn clean compile`
Expected: BUILD SUCCESS

**Step 4: Verify app launches**

Run: `mvn exec:java -Dexec.mainClass="dev.todoapp.TodoApp"`
Expected: TUI renders with "Welcome to Todo App!" panel, press q to quit

**Step 5: Init git and commit**

```bash
git init
echo "target/" > .gitignore
echo "*.class" >> .gitignore
echo ".idea/" >> .gitignore
echo "*.iml" >> .gitignore
git add pom.xml src/ .gitignore CLAUDE.md docs/ tasks/
git commit -m "feat: initial project setup with TamboUI skeleton"
```

---

## Task 2: Data Model - Enums

**Files:**
- Create: `src/main/java/dev/todoapp/model/TaskStatus.java`
- Create: `src/main/java/dev/todoapp/model/Priority.java`
- Test: `src/test/java/dev/todoapp/model/TaskStatusTest.java`
- Test: `src/test/java/dev/todoapp/model/PriorityTest.java`

**Step 1: Write failing tests for TaskStatus**

```java
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
```

**Step 2: Write failing tests for Priority**

```java
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
```

**Step 3: Run tests to verify they fail**

Run: `mvn test`
Expected: FAIL - classes don't exist

**Step 4: Implement TaskStatus**

```java
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
```

**Step 5: Implement Priority**

```java
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
```

**Step 6: Run tests to verify they pass**

Run: `mvn test`
Expected: All 8 tests PASS

**Step 7: Commit**

```bash
git add src/
git commit -m "feat: add TaskStatus and Priority enums with labels and symbols"
```

---

## Task 3: Data Model - Task and SubTask

**Files:**
- Create: `src/main/java/dev/todoapp/model/SubTask.java`
- Create: `src/main/java/dev/todoapp/model/Task.java`
- Test: `src/test/java/dev/todoapp/model/TaskTest.java`

**Step 1: Write failing tests**

```java
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
        assertFalse(task.isOverdue()); // done tasks are never overdue
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
```

**Step 2: Run tests to verify they fail**

Run: `mvn test`
Expected: FAIL

**Step 3: Implement SubTask**

```java
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
```

**Step 4: Implement Task**

```java
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
```

**Step 5: Run tests to verify they pass**

Run: `mvn test`
Expected: All tests PASS

**Step 6: Commit**

```bash
git add src/
git commit -m "feat: add Task and SubTask models with search, status cycling, progress"
```

---

## Task 4: JSON Persistence

**Files:**
- Create: `src/main/java/dev/todoapp/storage/JsonTaskStore.java`
- Test: `src/test/java/dev/todoapp/storage/JsonTaskStoreTest.java`

**Step 1: Write failing tests**

```java
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
```

**Step 2: Run tests to verify they fail**

Run: `mvn test -pl . -Dtest="dev.todoapp.storage.JsonTaskStoreTest"`
Expected: FAIL

**Step 3: Implement JsonTaskStore**

```java
package dev.todoapp.storage;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.reflect.TypeToken;
import dev.todoapp.model.*;

import java.io.IOException;
import java.lang.reflect.Type;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.LocalDate;
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
```

Note: Gson may need custom adapters for `LocalDate`/`LocalDateTime`. If the save/load test fails on date serialization, add type adapters:

```java
// Add to GsonBuilder if needed:
.registerTypeAdapter(LocalDate.class, (com.google.gson.JsonSerializer<LocalDate>)
    (src, type, ctx) -> new com.google.gson.JsonPrimitive(src.toString()))
.registerTypeAdapter(LocalDate.class, (com.google.gson.JsonDeserializer<LocalDate>)
    (json, type, ctx) -> LocalDate.parse(json.getAsString()))
.registerTypeAdapter(LocalDateTime.class, (com.google.gson.JsonSerializer<LocalDateTime>)
    (src, type, ctx) -> new com.google.gson.JsonPrimitive(src.toString()))
.registerTypeAdapter(LocalDateTime.class, (com.google.gson.JsonDeserializer<LocalDateTime>)
    (json, type, ctx) -> LocalDateTime.parse(json.getAsString()))
```

**Step 4: Run tests to verify they pass**

Run: `mvn test`
Expected: All tests PASS. If date tests fail, add the type adapters above.

**Step 5: Commit**

```bash
git add src/
git commit -m "feat: add JSON persistence with Gson and sample tasks"
```

---

## Task 5: TaskController

**Files:**
- Create: `src/main/java/dev/todoapp/controller/TaskController.java`
- Test: `src/test/java/dev/todoapp/controller/TaskControllerTest.java`

**Step 1: Write failing tests**

```java
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
        // Tasks are already sorted by creation date (order added)
        List<Task> sorted = controller.allTasks();
        assertTrue(sorted.get(0).createdAt().isBefore(sorted.get(2).createdAt())
                || sorted.get(0).createdAt().isEqual(sorted.get(2).createdAt()));
    }
}
```

**Step 2: Run tests to verify they fail**

Run: `mvn test`
Expected: FAIL

**Step 3: Implement TaskController**

```java
package dev.todoapp.controller;

import dev.todoapp.model.*;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;

public class TaskController {
    private final List<Task> tasks = new ArrayList<>();
    private int selectedIndex = 0;
    private String searchQuery = "";
    private TaskStatus activeTab = null; // null = All

    public List<Task> allTasks() { return tasks; }

    public void setTasks(List<Task> newTasks) {
        tasks.clear();
        tasks.addAll(newTasks);
        clampSelectedIndex();
    }

    public Task addTask(String title) {
        Task task = Task.create(title);
        tasks.add(task);
        return task;
    }

    public void deleteSelected() {
        List<Task> visible = filteredTasks();
        if (visible.isEmpty()) return;
        int idx = Math.min(selectedIndex, visible.size() - 1);
        Task toRemove = visible.get(idx);
        tasks.remove(toRemove);
        clampSelectedIndex();
    }

    public void cycleSelectedStatus() {
        Task task = selectedTask();
        if (task != null) task.cycleStatus();
    }

    public Task selectedTask() {
        List<Task> visible = filteredTasks();
        if (visible.isEmpty()) return null;
        int idx = Math.min(selectedIndex, visible.size() - 1);
        return visible.get(idx);
    }

    public int selectedIndex() { return selectedIndex; }

    public void setSelectedIndex(int index) {
        this.selectedIndex = index;
        clampSelectedIndex();
    }

    public void moveUp() { setSelectedIndex(selectedIndex - 1); }
    public void moveDown() { setSelectedIndex(selectedIndex + 1); }

    public List<Task> tasksByStatus(TaskStatus status) {
        return tasks.stream().filter(t -> t.status() == status).toList();
    }

    public List<Task> filteredTasks() {
        return tasks.stream()
                .filter(t -> activeTab == null || t.status() == activeTab)
                .filter(t -> searchQuery.isEmpty() || t.matchesSearch(searchQuery))
                .toList();
    }

    public String searchQuery() { return searchQuery; }

    public void setSearchQuery(String query) {
        this.searchQuery = query != null ? query : "";
        clampSelectedIndex();
    }

    public void clearSearch() { setSearchQuery(""); }

    public TaskStatus activeTab() { return activeTab; }

    public void setActiveTab(TaskStatus tab) {
        this.activeTab = tab;
        this.selectedIndex = 0;
    }

    public void cycleTab() {
        if (activeTab == null) activeTab = TaskStatus.IN_PROGRESS;
        else if (activeTab == TaskStatus.IN_PROGRESS) activeTab = TaskStatus.DONE;
        else activeTab = null;
        this.selectedIndex = 0;
    }

    public void sortByCreatedAt() {
        tasks.sort(Comparator.comparing(Task::createdAt));
    }

    public void sortByDueDate() {
        tasks.sort(Comparator.comparing(Task::dueDate, Comparator.nullsLast(Comparator.naturalOrder())));
    }

    public void sortByPriority() {
        tasks.sort(Comparator.comparing(Task::priority).reversed());
    }

    public int totalCount() { return tasks.size(); }
    public long doneCount() { return tasks.stream().filter(t -> t.status() == TaskStatus.DONE).count(); }
    public long inProgressCount() { return tasks.stream().filter(t -> t.status() == TaskStatus.IN_PROGRESS).count(); }

    private void clampSelectedIndex() {
        List<Task> visible = filteredTasks();
        if (visible.isEmpty()) {
            selectedIndex = 0;
        } else {
            selectedIndex = Math.max(0, Math.min(selectedIndex, visible.size() - 1));
        }
    }
}
```

**Step 4: Run tests to verify they pass**

Run: `mvn test`
Expected: All tests PASS

**Step 5: Commit**

```bash
git add src/
git commit -m "feat: add TaskController with CRUD, filtering, search, tab support"
```

---

## Task 6: TCSS Theme + Key Bindings

**Files:**
- Create: `src/main/resources/themes/dark.tcss`
- Create: `src/main/resources/bindings.properties`

**Step 1: Create dark.tcss theme**

```tcss
/* Dark Theme for Todo App */

$accent: cyan;
$bg: black;
$fg: white;
$border: gray;
$border-focus: cyan;
$done: green;
$overdue: red;
$priority-low: green;
$priority-medium: yellow;
$priority-high: red;
$priority-urgent: bright-magenta;

/* Global */
Panel {
    border-type: rounded;
    border-color: $border;
    &:focus {
        border-color: $border-focus;
    }
}

/* Header */
#header {
    border-type: rounded;
    border-color: $accent;
}

TabsElement-tab {
    color: gray;
}

TabsElement-tab:selected {
    color: $accent;
    text-style: bold reversed;
}

/* Task List */
#task-list {
    border-color: $border;
    &:focus {
        border-color: $border-focus;
    }
}

ListElement-item {
    color: $fg;
}

ListElement-item:selected {
    background: $accent;
    color: black;
    text-style: bold;
}

/* Task Detail */
#task-detail {
    border-color: $border;
}

/* Status Bar */
#status-bar {
    color: gray;
    text-style: dim;
}

/* Search */
#search-input {
    border-type: rounded;
    border-color: $border;
    &:focus {
        border-color: $accent;
    }
}

TextInputElement-placeholder {
    color: gray;
    text-style: italic;
}

/* Priority Colors */
.priority-low { color: $priority-low; }
.priority-medium { color: $priority-medium; }
.priority-high { color: $priority-high; }
.priority-urgent { color: $priority-urgent; text-style: bold; }

/* Status Colors */
.status-done { color: $done; }
.status-overdue { color: $overdue; text-style: bold; }

/* Form Dialog */
#dialog {
    border-type: double;
    border-color: $accent;
    background: black;
}
```

**Step 2: Create bindings.properties**

```properties
# Navigation
moveUp = Up, k
moveDown = Down, j
select = Enter, Space
cancel = Escape

# Tab navigation
nextTab = Tab

# Task actions
addTask = a
editTask = e
deleteTask = d
toggleStatus = s

# Search
search = /

# Sorting
sortByDate = F1
sortByPriority = F2
sortByDueDate = F3

# App
quit = q, Ctrl+C
help = ?
```

**Step 3: Verify build still compiles**

Run: `mvn clean compile`
Expected: BUILD SUCCESS

**Step 4: Commit**

```bash
git add src/main/resources/
git commit -m "feat: add dark TCSS theme and key bindings config"
```

---

## Task 7: Main TodoApp with Full TUI Layout

**Files:**
- Modify: `src/main/java/dev/todoapp/TodoApp.java`

This is the core task - wire up the full TUI with dock layout, tabs, list, detail panel, search, status bar, and all event handling. This is a single large file since TamboUI Toolkit DSL works best with the render method in the app class, keeping component state as fields.

**Step 1: Implement the full TodoApp**

```java
package dev.todoapp;

import dev.tamboui.style.Color;
import dev.tamboui.toolkit.*;
import dev.tamboui.toolkit.events.EventResult;
import dev.tamboui.toolkit.events.KeyEvent;
import dev.todoapp.controller.TaskController;
import dev.todoapp.model.*;
import dev.todoapp.storage.JsonTaskStore;

import java.nio.file.Path;
import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.List;

import static dev.tamboui.toolkit.Toolkit.*;

public class TodoApp extends ToolkitApp {

    private static final DateTimeFormatter DATE_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd");
    private static final DateTimeFormatter DATETIME_FMT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm");

    private final TaskController controller = new TaskController();
    private final JsonTaskStore store;
    private final ListElement<?> taskList = list().highlightColor(Color.CYAN).autoScroll();

    // UI State
    private boolean searchMode = false;
    private boolean showHelp = false;
    private boolean confirmDelete = false;
    private boolean addMode = false;
    private boolean editMode = false;
    private String inputBuffer = "";
    private int editingSubtaskIndex = -1;

    public TodoApp() {
        Path dataDir = Path.of(System.getProperty("user.home"), ".todo-app");
        this.store = new JsonTaskStore(dataDir.resolve("tasks.json"));
        controller.setTasks(store.loadOrCreateSamples());
    }

    @Override
    protected Element render() {
        Element mainContent = dock()
                .top(renderHeader())
                .center(row(renderTaskList().fill(1), renderTaskDetail().fill(2)).spacing(0))
                .bottom(renderStatusBar())
                .topHeight(3)
                .bottomHeight(1);

        // Overlay dialogs using stack
        if (showHelp) {
            return stack(mainContent, renderHelpDialog());
        }
        if (confirmDelete) {
            return stack(mainContent, renderConfirmDialog());
        }
        if (addMode || editMode) {
            return stack(mainContent, renderInputDialog());
        }

        return mainContent;
    }

    // ---- Header with Tabs ----

    private Element renderHeader() {
        String allLabel = "All (" + controller.totalCount() + ")";
        String progressLabel = "In Progress (" + controller.inProgressCount() + ")";
        String doneLabel = "Done (" + controller.doneCount() + ")";

        int activeIndex;
        TaskStatus tab = controller.activeTab();
        if (tab == null) activeIndex = 0;
        else if (tab == TaskStatus.IN_PROGRESS) activeIndex = 1;
        else activeIndex = 2;

        Element tabsRow = tabs(allLabel, progressLabel, doneLabel)
                .selected(activeIndex)
                .highlightColor(Color.CYAN);

        if (searchMode) {
            return column(
                    tabsRow,
                    panel(text("[/] " + inputBuffer + "█").cyan()).id("search-input")
            );
        }

        return panel(tabsRow).id("header");
    }

    // ---- Task List (Left Panel) ----

    private Element renderTaskList() {
        List<Task> visible = controller.filteredTasks();

        if (visible.isEmpty()) {
            return panel("Tasks",
                    spacer(),
                    text("No tasks found").dim(),
                    text(controller.searchQuery().isEmpty()
                            ? "Press 'a' to add a task"
                            : "Try a different search").dim(),
                    spacer()
            ).id("task-list").focusable();
        }

        String[] items = visible.stream()
                .map(this::formatTaskTitle)
                .toArray(String[]::new);

        return panel("Tasks",
                taskList.items(items).selected(controller.selectedIndex())
        ).id("task-list")
                .focusable()
                .onKeyEvent(this::handleKeyEvent);
    }

    private String formatTaskTitle(Task task) {
        String status = task.status().symbol();
        String priority = task.priority().symbol();
        String overdue = task.isOverdue() ? " !" : "";
        String subtaskInfo = task.subtasks().isEmpty() ? ""
                : " [" + task.subtasks().stream().filter(SubTask::completed).count()
                + "/" + task.subtasks().size() + "]";
        return status + " " + priority + " " + task.title() + subtaskInfo + overdue;
    }

    // ---- Task Detail (Right Panel) ----

    private Element renderTaskDetail() {
        Task task = controller.selectedTask();

        if (task == null) {
            return panel("Details",
                    spacer(),
                    text("Select a task to view details").dim(),
                    spacer()
            ).id("task-detail");
        }

        List<Element> content = new java.util.ArrayList<>();

        // Title
        content.add(text(task.title()).bold().cyan());
        content.add(text(""));

        // Status and Priority
        content.add(text("Status:   " + task.status().symbol() + " " + task.status().label()));
        content.add(text("Priority: " + task.priority().symbol() + " " + task.priority().label()));

        // Dates
        content.add(text("Created:  " + task.createdAt().format(DATETIME_FMT)).dim());
        if (task.dueDate() != null) {
            String dueText = "Due:      " + task.dueDate().format(DATE_FMT);
            if (task.isOverdue()) {
                content.add(text(dueText + " (OVERDUE)").fg(Color.RED).bold());
            } else {
                content.add(text(dueText));
            }
        }

        // Tags
        if (!task.tags().isEmpty()) {
            content.add(text("Tags:     " + String.join(", ", task.tags())).dim());
        }

        // Description
        if (task.description() != null && !task.description().isEmpty()) {
            content.add(text(""));
            content.add(text("Description").bold());
            content.add(text(task.description()));
        }

        // Subtasks
        if (!task.subtasks().isEmpty()) {
            content.add(text(""));
            long done = task.subtasks().stream().filter(SubTask::completed).count();
            content.add(text("Subtasks (" + done + "/" + task.subtasks().size() + ")").bold());
            for (SubTask sub : task.subtasks()) {
                String check = sub.completed() ? "✓" : "○";
                Color color = sub.completed() ? Color.GREEN : Color.WHITE;
                content.add(text("  " + check + " " + sub.title()).fg(color));
            }
        }

        content.add(spacer());

        return panel("Details", content.toArray(Element[]::new)).id("task-detail");
    }

    // ---- Status Bar ----

    private Element renderStatusBar() {
        long total = controller.totalCount();
        long done = controller.doneCount();
        long progress = controller.inProgressCount();
        String today = LocalDate.now().format(DATE_FMT);

        String left = total + " tasks | " + progress + " active | " + done + " done | " + today;
        String right = "[a]dd [e]dit [d]el [s]tatus [/]search [Tab]tab [?]help [q]uit";

        return row(
                text(" " + left).dim(),
                spacer(),
                text(right + " ").dim()
        ).id("status-bar");
    }

    // ---- Dialogs ----

    private Element renderHelpDialog() {
        return panel("Keyboard Shortcuts",
                text(""),
                text("  j / ↓       Move down").cyan(),
                text("  k / ↑       Move up").cyan(),
                text("  Enter       Select / toggle subtask"),
                text("  a           Add new task"),
                text("  e           Edit selected task"),
                text("  d           Delete selected task"),
                text("  s           Cycle task status"),
                text("  Tab         Switch tab"),
                text("  /           Search tasks"),
                text("  F1          Sort by date"),
                text("  F2          Sort by priority"),
                text("  F3          Sort by due date"),
                text("  Escape      Close dialog / cancel"),
                text("  q           Quit"),
                text(""),
                text("  Press any key to close").dim()
        ).rounded().id("dialog")
                .focusable()
                .onKeyEvent(e -> { showHelp = false; return EventResult.HANDLED; });
    }

    private Element renderConfirmDialog() {
        Task task = controller.selectedTask();
        String title = task != null ? task.title() : "this task";
        return panel("Confirm Delete",
                text(""),
                text("  Delete \"" + title + "\"?").bold(),
                text(""),
                text("  [y] Yes   [n/Esc] No").dim(),
                text("")
        ).rounded().id("dialog")
                .focusable()
                .onKeyEvent(e -> {
                    if (e.isChar('y') || e.isChar('Y')) {
                        controller.deleteSelected();
                        saveAndRefresh();
                    }
                    confirmDelete = false;
                    return EventResult.HANDLED;
                });
    }

    private Element renderInputDialog() {
        String title = addMode ? "New Task" : "Edit Task";
        return panel(title,
                text(""),
                text("  Title:").bold(),
                text("  " + inputBuffer + "█").cyan(),
                text(""),
                text("  [Enter] Confirm   [Esc] Cancel").dim(),
                text("")
        ).rounded().id("dialog")
                .focusable()
                .onKeyEvent(this::handleInputEvent);
    }

    // ---- Event Handling ----

    private EventResult handleKeyEvent(KeyEvent event) {
        // Navigation
        if (event.isUp() || event.isChar('k')) {
            controller.moveUp();
            return EventResult.HANDLED;
        }
        if (event.isDown() || event.isChar('j')) {
            controller.moveDown();
            taskList.selectNext(controller.filteredTasks().size());
            return EventResult.HANDLED;
        }

        // Tab switching
        if (event.isChar('\t')) {
            controller.cycleTab();
            return EventResult.HANDLED;
        }

        // Task actions
        if (event.isChar('a')) {
            addMode = true;
            inputBuffer = "";
            return EventResult.HANDLED;
        }
        if (event.isChar('e')) {
            Task selected = controller.selectedTask();
            if (selected != null) {
                editMode = true;
                inputBuffer = selected.title();
            }
            return EventResult.HANDLED;
        }
        if (event.isChar('d')) {
            if (controller.selectedTask() != null) {
                confirmDelete = true;
            }
            return EventResult.HANDLED;
        }
        if (event.isChar('s')) {
            controller.cycleSelectedStatus();
            saveAndRefresh();
            return EventResult.HANDLED;
        }

        // Search
        if (event.isChar('/')) {
            searchMode = true;
            inputBuffer = controller.searchQuery();
            return EventResult.HANDLED;
        }

        // Sorting
        if (event.isKey(dev.tamboui.tui.events.KeyCode.F1)) {
            controller.sortByCreatedAt();
            return EventResult.HANDLED;
        }
        if (event.isKey(dev.tamboui.tui.events.KeyCode.F2)) {
            controller.sortByPriority();
            return EventResult.HANDLED;
        }
        if (event.isKey(dev.tamboui.tui.events.KeyCode.F3)) {
            controller.sortByDueDate();
            return EventResult.HANDLED;
        }

        // Help
        if (event.isChar('?')) {
            showHelp = true;
            return EventResult.HANDLED;
        }

        return EventResult.UNHANDLED;
    }

    private EventResult handleInputEvent(KeyEvent event) {
        if (event.isCancel()) {
            addMode = false;
            editMode = false;
            inputBuffer = "";
            return EventResult.HANDLED;
        }
        if (event.isSelect() && !inputBuffer.isBlank()) {
            if (addMode) {
                controller.addTask(inputBuffer.trim());
            } else if (editMode) {
                Task task = controller.selectedTask();
                if (task != null) task.setTitle(inputBuffer.trim());
            }
            addMode = false;
            editMode = false;
            inputBuffer = "";
            saveAndRefresh();
            return EventResult.HANDLED;
        }
        // Backspace
        if (event.isChar('\b') || event.matches("deleteBackward")) {
            if (!inputBuffer.isEmpty()) {
                inputBuffer = inputBuffer.substring(0, inputBuffer.length() - 1);
            }
            if (searchMode) controller.setSearchQuery(inputBuffer);
            return EventResult.HANDLED;
        }
        // Regular character input
        char ch = event.character();
        if (ch >= 32 && ch < 127) {
            inputBuffer += ch;
            if (searchMode) controller.setSearchQuery(inputBuffer);
            return EventResult.HANDLED;
        }
        return EventResult.UNHANDLED;
    }

    private void saveAndRefresh() {
        store.save(controller.allTasks());
    }

    // ---- Main ----

    public static void main(String[] args) throws Exception {
        new TodoApp().run();
    }
}
```

**Step 2: Verify it compiles**

Run: `mvn clean compile`
Expected: BUILD SUCCESS

**Step 3: Run the app and test manually**

Run: `mvn exec:java`
Expected: Full TUI with tabs, task list, detail panel, status bar. Navigate with j/k, press a to add, s to toggle status, Tab to switch tabs, / to search, q to quit.

**Step 4: Fix any compilation or runtime issues**

Note: TamboUI is 0.2.0-SNAPSHOT so some API names may differ. Key areas to adjust:
- `list()` method might be `listWidget()` or similar
- `tabs()` method might need different construction
- `KeyEvent` methods may have slightly different signatures
- `EventResult` import path may differ
- `stack()` for overlays might need `ContentAlignment`

If the Toolkit DSL static imports differ, check actual class names in the tamboui-toolkit jar: `jar tf ~/.m2/repository/dev/tamboui/tamboui-toolkit/0.2.0-SNAPSHOT/*.jar | grep -i toolkit`

**Step 5: Commit**

```bash
git add src/
git commit -m "feat: implement full TUI with dock layout, tabs, search, dialogs"
```

---

## Task 8: Handle Search Mode Key Events in Main App

**Files:**
- Modify: `src/main/java/dev/todoapp/TodoApp.java`

The search mode needs special handling at the app level since key events flow to the focused element. We need the search input to capture all typing when active.

**Step 1: Override the global event handler**

Add to `TodoApp.java` - the render method should use `.onKeyEvent` on the outermost element when in search mode. The search panel should be focusable and receive key events:

When `searchMode` is true, the key handler should:
- Escape: exit search mode, clear query if empty
- Enter: exit search mode, keep query
- Backspace: delete last char
- Printable chars: append to query
- All other keys: ignore

**Step 2: Verify search works**

Run: `mvn exec:java`
Test: Press `/`, type a query, see list filter. Press Esc to cancel, Enter to confirm search.

**Step 3: Commit**

```bash
git add src/
git commit -m "fix: search mode key event handling"
```

---

## Task 9: GraalVM Native Image

**Files:**
- Create: `src/main/resources/META-INF/native-image/dev.todoapp/todo-app/reflect-config.json`
- Create: `src/main/resources/META-INF/native-image/dev.todoapp/todo-app/resource-config.json`

**Step 1: Create reflect-config.json**

```json
[
  {
    "name": "dev.todoapp.model.Task",
    "allDeclaredFields": true,
    "allDeclaredConstructors": true,
    "allDeclaredMethods": true
  },
  {
    "name": "dev.todoapp.model.SubTask",
    "allDeclaredFields": true,
    "allDeclaredConstructors": true,
    "allDeclaredMethods": true
  },
  {
    "name": "dev.todoapp.model.TaskStatus",
    "allDeclaredFields": true,
    "allDeclaredConstructors": true,
    "allDeclaredMethods": true
  },
  {
    "name": "dev.todoapp.model.Priority",
    "allDeclaredFields": true,
    "allDeclaredConstructors": true,
    "allDeclaredMethods": true
  }
]
```

**Step 2: Create resource-config.json**

```json
{
  "resources": {
    "includes": [
      { "pattern": "themes/.*\\.tcss$" },
      { "pattern": "bindings\\.properties$" }
    ]
  }
}
```

**Step 3: Build native image**

Run: `mvn -Pnative package`
Expected: BUILD SUCCESS, native binary at `target/todo-app`

**Step 4: Test native binary**

Run: `./target/todo-app`
Expected: App launches instantly, all features work

**Step 5: Commit**

```bash
git add src/main/resources/META-INF/ pom.xml
git commit -m "feat: add GraalVM native image config and build profile"
```

---

## Task 10: Final Integration Test & Polish

**Step 1: Run all tests**

Run: `mvn test`
Expected: All tests PASS

**Step 2: Manual QA checklist**

Run the app (`mvn exec:java`) and verify:
- [ ] App launches with sample tasks
- [ ] j/k navigate the task list
- [ ] Tab switches between All / In Progress / Done tabs
- [ ] 'a' opens add dialog, Enter confirms, Esc cancels
- [ ] 'e' opens edit dialog with current title pre-filled
- [ ] 'd' shows confirmation, 'y' deletes, 'n' cancels
- [ ] 's' cycles status (Pending -> In Progress -> Done -> Pending)
- [ ] '/' enters search mode, typing filters list
- [ ] '?' shows help overlay
- [ ] Detail panel shows description, subtasks, dates, priority
- [ ] Status bar shows counts and shortcuts
- [ ] Data persists to `~/.todo-app/tasks.json`
- [ ] q quits the application

**Step 3: Fix any issues found during QA**

**Step 4: Final commit**

```bash
git add -A
git commit -m "feat: todo app v1.0 - TUI task manager with TamboUI"
```

---

## Execution Notes

- TamboUI is version 0.2.0-SNAPSHOT - APIs may have changed. If compilation fails, check the actual jar contents and adjust imports.
- The Toolkit DSL static import is `import static dev.tamboui.toolkit.Toolkit.*` - this provides `panel()`, `text()`, `row()`, `column()`, `dock()`, `list()`, `tabs()`, `stack()`, `spacer()`, etc.
- Gson needs custom type adapters for `LocalDate` and `LocalDateTime` - add them in Task 4 if serialization fails.
- The Panama backend (`tamboui-panama`) could replace JLine for better performance but requires Java 22+. Since we use Java 25, it's an option if JLine has issues.
- Key event handling may need adjustment based on actual TamboUI API - method names like `isChar()`, `isUp()`, `isKey()`, `character()` are based on the docs but may vary.
