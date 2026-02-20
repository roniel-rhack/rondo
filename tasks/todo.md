# Todo App TamboUI - Implementation Plan

## Phase 1: Project Setup
- [ ] Initialize Maven project with `pom.xml`
- [ ] Configure TamboUI dependencies (toolkit, jline, css, processor) and snapshot repository
- [ ] Set up Java 25 compiler + annotation processor in maven-compiler-plugin
- [ ] Create base package structure (`dev.todoapp`)
- [ ] Verify build compiles with a minimal ToolkitApp

## Phase 2: Data Model
- [ ] Create `Task` model (id, title, description, status, priority, dates, subtasks, tags)
- [ ] Create `SubTask` model (id, title, completed)
- [ ] Create `TaskStatus` enum (PENDING, IN_PROGRESS, DONE)
- [ ] Create `Priority` enum (LOW, MEDIUM, HIGH, URGENT)
- [ ] Write unit tests for models

## Phase 3: Persistence
- [ ] Implement `JsonTaskStore` for JSON file persistence
- [ ] Support save/load of task list
- [ ] Auto-save on changes
- [ ] Handle first-run (create file with sample tasks)
- [ ] Write unit tests for persistence

## Phase 4: Controllers
- [ ] Implement `TaskController` (CRUD, status changes, sorting, filtering)
- [ ] Implement `SearchController` (search by title, filter by status/priority)
- [ ] Write unit tests for controllers

## Phase 5: Views & UI
- [ ] Create dark theme TCSS stylesheet
- [ ] Build `HeaderView` (tabs + search bar)
- [ ] Build `TaskListView` (left panel with task titles)
- [ ] Build `TaskDetailView` (right panel with description, subtasks, metadata)
- [ ] Build `TaskFormView` (add/edit dialog)
- [ ] Build status bar footer
- [ ] Wire up Dock layout in main `TodoApp`

## Phase 6: Event Handling & Bindings
- [ ] Configure vim-style + standard key bindings
- [ ] Implement tab switching
- [ ] Implement task selection and navigation
- [ ] Implement add/edit/delete workflows
- [ ] Implement search focus and filtering
- [ ] Implement status toggle
- [ ] Implement subtask toggle

## Phase 7: Polish
- [ ] Confirmation dialog for delete
- [ ] Empty state messages
- [ ] Priority color coding
- [ ] Due date formatting and overdue highlighting
- [ ] Help overlay (key bindings reference)
- [ ] GraalVM native image configuration (native-maven-plugin profile)
- [ ] Add GraalVM reflection/resource config if needed (`META-INF/native-image/`)
- [ ] Verify native binary runs correctly (`mvn -Pnative package`)

## Review
_To be filled after implementation_
