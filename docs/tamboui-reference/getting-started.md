# TamboUI Getting Started Guide

## Version Information
Current version: **0.2.0-SNAPSHOT**

## Requirements

The framework requires "Java 8 or later (Java 17+ highly recommended)" for compatibility.

## Installation Methods

### Maven Setup

Add the snapshot repository and dependencies to `pom.xml`:

```xml
<repository>
    <id>ossrh-snapshots</id>
    <url>https://central.sonatype.com/repository/maven-snapshots/</url>
    <releases>
        <enabled>false</enabled>
    </releases>
    <snapshots>
        <enabled>true</enabled>
    </snapshots>
</repository>

<dependencies>
    <groupId>dev.tamboui</groupId>
    <artifactId>tamboui-bom</artifactId>
    <version>0.2.0-SNAPSHOT</version>
    <type>pom</type>
    <scope>import</scope>

    <groupId>dev.tamboui</groupId>
    <artifactId>tamboui-toolkit</artifactId>

    <groupId>dev.tamboui</groupId>
    <artifactId>tamboui-jline</artifactId>
</dependencies>
```

### Gradle (Kotlin DSL)

```kotlin
repositories {
    maven {
        url = uri("https://central.sonatype.com/repository/maven-snapshots/")
        mavenContent {
            snapshotsOnly()
        }
    }
}

dependencies {
    implementation(platform("dev.tamboui:tamboui-bom:0.2.0-SNAPSHOT"))
    implementation("dev.tamboui:tamboui-toolkit")
    implementation("dev.tamboui:tamboui-jline")
}
```

### JBang Configuration

```bash
///usr/bin/env jbang "$0" "$@" ; exit $?
//DEPS dev.tamboui:tamboui-toolkit:LATEST
//DEPS dev.tamboui:tamboui-jline:LATEST
//REPOS https://central.sonatype.com/repository/maven-snapshots/
```

## API Levels Overview

Four distinct API levels serve different development needs:

| Level | Use Case | Complexity |
|-------|----------|-----------|
| Toolkit DSL | Component-based UI, most applications | Low |
| TuiRunner | Custom event handling, animations | Medium |
| Immediate Mode | Maximum control, custom rendering | High |
| Inline Display | CLI progress bars, status lines | Low |

## Hello World Examples

### Toolkit DSL (Recommended)

```java
public class HelloDsl extends ToolkitApp {

    @Override
    protected Element render() {
        return panel("Hello",
            text("Welcome to TamboUI DSL!").bold().cyan(),
            spacer(),
            text("Press 'q' to quit").dim()
        ).rounded();
    }

    public void runApp() throws Exception {
        new HelloDsl().run();
    }
}
```

### TuiRunner with Event Handling

```java
try (var tui = TuiRunner.create()) {
    tui.run(
        (event, runner) ->
            switch (event) {
                case KeyEvent k when k.isQuit() -> {
                    runner.quit();
                    yield false;
                }
                default -> false;
            },
        frame -> {
            var paragraph = Paragraph.builder()
                .text(Text.from("Hello, TamboUI! Press 'q' to quit."))
                .build();
            frame.renderWidget(paragraph, frame.area());
        }
    );
}
```

### Immediate Mode

```java
try (var backend = BackendFactory.create()) {
    backend.enableRawMode();
    backend.enterAlternateScreen();

    try (var terminal = new Terminal<>(backend)) {
        terminal.draw(frame -> {
            var paragraph = Paragraph.builder()
                .text(Text.from("Hello, Immediate Mode!"))
                .build();
            frame.renderWidget(paragraph, frame.area());
        });
    }
    Thread.sleep(2000);
}
```

### Inline Display Mode

```java
try (var display = InlineDisplay.create(2)) {
    for (int i = 0; i <= 100; i += 5) {
        int progress = i;
        display.render((area, buffer) -> {
            var gauge = Gauge.builder()
                .ratio(progress / 100.0)
                .label("Progress: " + progress + "%")
                .build();
            gauge.render(area, buffer);
        });
        Thread.sleep(50);
    }
    display.println("Done!");
}
```

## Running and Debugging

### Maven

Maven execution works seamlessly with TamboUI. Standard `mvn exec:java` commands function correctly.

### Gradle

Important note: "Gradle runs Java via a daemon that has no terminal." The workaround involves building the application first, then executing with `java -jar` or `jbang run` from a real terminal:

```bash
./gradlew build
java -jar build/libs/my-app-1.0.0-SNAPSHOT-all.jar
```

For debugging: `jbang run --debug build/libs/my-app-1.0.0-SNAPSHOT-all.jar`

### IDEs and Editors

**Visual Studio Code**: Works well with Maven, Gradle, and JBang using standard Run & Debug features.

**IntelliJ IDEA**: Run & Debug does not reliably obtain a real terminal. Workaround: build the application, then execute from the command line or IDE terminal.

**Apache NetBeans**: Similar terminal limitations as IntelliJ. Use external terminal with `java -jar` or `jbang run`.

**Eclipse**: Console and built-in terminal are not full TTYs. External terminal with `java -jar` or `jbang run` is recommended.

## Running Demos

### Without Repository Cloning

```bash
jbang demos@tamboui
```

This launches an interactive demo selector.

### From TamboUI Sources

**List available demos:**
```bash
jbang alias list .
```

**Run specific demo:**
```bash
jbang sparkline-demo
```

**Using run-demo.sh:**
```bash
./run-demo.sh
./run-demo.sh sparkline-demo
./run-demo.sh sparkline-demo --native
```

**Manual build and run:**
```bash
./gradlew :demos:sparkline-demo:installDist
./demos/sparkline-demo/build/install/sparkline-demo/bin/sparkline-demo
```

**Native compilation (requires GraalVM):**
```bash
./gradlew :demos:sparkline-demo:nativeCompile
./demos/sparkline-demo/build/native/nativeCompile/sparkline-demo
```

## Next Steps

- Review Core Concepts documentation for architectural understanding
- Explore Widgets Reference for available UI components
- Study API Levels documentation for detailed information
- Follow Application Structure guide for maintainable app design
