package dev.todoapp;

import dev.tamboui.toolkit.app.ToolkitApp;
import dev.tamboui.toolkit.element.Element;
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
