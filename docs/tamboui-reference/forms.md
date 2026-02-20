# TamboUI Forms Documentation

## Overview

TamboUI provides high-level form abstractions to simplify interactive form construction. Rather than managing individual state declarations, labels, and styling separately, developers can use centralized form state management with built-in validation support.

## Form Fields

### Basic Usage

The `formField()` factory creates a labeled input field pairing a label with an input widget:

```java
import static dev.tamboui.toolkit.Toolkit.*;

TextInputState nameState = new TextInputState("John");
formField("Name", nameState);

// Creates internal state automatically
formField("Email")
    .placeholder("you@example.com");
```

### Field Types

| Type | Description | State Class |
|------|-------------|------------|
| `TEXT` (default) | Single-line text input | `TextInputState` |
| `CHECKBOX` | Boolean checkbox with [x] rendering | `BooleanFieldState` |
| `TOGGLE` | On/off toggle switch | `BooleanFieldState` |
| `SELECT` | Dropdown selection | `SelectFieldState` |

### Text Fields

```java
TextInputState emailState = new TextInputState("");
formField("Email", emailState)
    .placeholder("you@example.com")
    .rounded();
```

### Password Fields

Use `maskedField()` in the FormState builder for password fields:

```java
FormState form = FormState.builder()
    .textField("username", "")
    .maskedField("password", "")  // automatically masked with '*'
    .build();
```

For `formField()`, use `.masked()` explicitly:

```java
formField("Password", passwordState)
    .placeholder("Enter password")
    .masked()          // displays '*' for each character
    .rounded();

// Or with a custom mask character
formField("PIN", pinState)
    .masked('\u25CF');       // displays a filled circle for each character
```

### Boolean Fields (Checkbox/Toggle)

```java
BooleanFieldState subscribeState = new BooleanFieldState(false);

// Checkbox style: [x] or [ ]
formField("Subscribe", subscribeState);

// Toggle style: [ON] or [OFF]
formField("Dark Mode", subscribeState, FieldType.TOGGLE);
```

### Select Fields

```java
SelectFieldState countryState = new SelectFieldState("USA", "UK", "Germany");
formField("Country", countryState);
```

### Styling

FormFieldElement supports comprehensive styling:

```java
formField("Email", emailState)
    .labelWidth(14)           // Fixed label width for alignment
    .spacing(2)               // Gap between label and input
    .rounded()                // Rounded border
    .borderColor(Color.DARK_GRAY)
    .focusedBorderColor(Color.CYAN)
    .errorBorderColor(Color.RED);
```

### CSS Selectors

FormFieldElement exposes these CSS child selectors:

| Selector | Description |
|----------|------------|
| `FormFieldElement-label` | The label text style |
| `FormFieldElement-input` | The input wrapper |
| `FormFieldElement-error` | The inline error message style |

Example TCSS:

```css
FormFieldElement-label {
    color: gray;
}

FormFieldElement-error {
    color: red;
}
```

## Form State

`FormState` provides centralized state management for all form fields, eliminating individual state declarations.

### Creating Form State

```java
FormState form = FormState.builder()
    // Text fields
    .textField("username", "")
    .textField("email", "user@example.com")

    // Boolean fields
    .booleanField("newsletter", true)
    .booleanField("darkMode", false)

    // Select fields
    .selectField("country", Arrays.asList("USA", "UK", "Germany"), 0)
    .selectField("role", Arrays.asList("Admin", "User", "Guest"))  // defaults to index 0

    .build();
```

### Accessing Values

```java
// Get state objects for UI binding
TextInputState usernameState = form.textField("username");
BooleanFieldState newsletterState = form.booleanField("newsletter");
SelectFieldState countryState = form.selectField("country");

// Get/set text values directly
String email = form.textValue("email");
form.setTextValue("email", "new@example.com");

// Get/set boolean values
boolean subscribed = form.booleanValue("newsletter");
form.setBooleanValue("newsletter", false);

// Get/set select values
String country = form.selectValue("country");
int countryIndex = form.selectIndex("country");
form.selectIndex("country", 2);  // Select "Germany"

// Get all text values as map
Map<String, String> allText = form.textValues();
```

## Validation

The validation framework provides built-in validators and supports custom validation logic.

### Built-in Validators

| Validator | Description | Example |
|-----------|------------|---------|
| `required()` | Field must not be empty | `Validators.required()` |
| `email()` | Valid email format | `Validators.email()` |
| `minLength(n)` | Minimum n characters | `Validators.minLength(3)` |
| `maxLength(n)` | Maximum n characters | `Validators.maxLength(100)` |
| `pattern(regex)` | Matches regex pattern | `Validators.pattern("\\d{5}")` |
| `range(min, max)` | Numeric value in range | `Validators.range(1, 100)` |

### Using Validators

```java
formField("Email", emailState)
    .validate(Validators.required(), Validators.email())
    .errorBorderColor(Color.RED)
    .showInlineErrors(true);
```

### Custom Error Messages

All built-in validators accept optional custom messages:

```java
Validators.required("Please enter your name");
Validators.email("Invalid email format");
Validators.minLength(3, "Username must be at least 3 characters");
Validators.range(1, 100, "Age must be between 1 and 100");
```

### Triggering Validation

```java
FormFieldElement field = formField("Email", emailState)
    .validate(Validators.required(), Validators.email());

// Validate and get result
ValidationResult result = field.validateField();

if (!result.isValid()) {
    String error = result.errorMessage();  // "Field is required" or "Invalid email"
}

// Get last validation result
ValidationResult lastResult = field.lastValidation();
```

### Custom Validators

Create custom validators using the `Validator` functional interface:

```java
Validator usernameAvailable = value -> {
    if (userService.exists(value)) {
        return ValidationResult.invalid("Username already taken");
    }
    return ValidationResult.valid();
};

formField("Username", usernameState)
    .validate(Validators.required(), usernameAvailable);
```

### Composing Validators

Validators can be composed using `and()`:

```java
Validator emailValidator = Validators.required()
    .and(Validators.email())
    .and(Validators.maxLength(100));

formField("Email", emailState)
    .validate(emailValidator);
```

## Form Container

`FormElement` provides a container for managing multiple form fields with optional grouping.

### Basic Usage

```java
form(formState)
    .field("fullName", "Full Name")
    .field("email", "Email")
    .field("role", "Role")
    .labelWidth(14)
    .rounded();
```

### Grouping Fields

```java
form(formState)
    .group("Personal Info")
        .field("fullName", "Full Name")
        .field("email", "Email")
        .field("phone", "Phone")
    .group("Preferences")
        .field("newsletter", "Newsletter", FieldType.CHECKBOX)
        .field("darkMode", "Dark Mode", FieldType.TOGGLE)
    .labelWidth(14)
    .spacing(1);
```

### Form Submission

#### Submit on Enter

Pressing Enter in any text field triggers form submission:

```java
form(formState)
    .field("username", "Username")
    .field("password", "Password")
    .submitOnEnter(true)
    .onSubmit(state -> {
        String user = state.textValue("username");
        String pass = state.textValue("password");
        authenticate(user, pass);
    });
```

#### Programmatic Submit (Button)

Call `submit()` from a button or other trigger:

```java
FormElement loginForm = form(formState)
    .field("username", "Username", Validators.required())
    .field("password", "Password", Validators.required())
    .onSubmit(state -> authenticate(state));

// Render form with submit button
column(
    loginForm,
    text(" Login ").bold()
);
```

#### Validation on Submit

By default, `submit()` validates all fields before calling the `onSubmit` callback. If validation fails, the callback is not called and `submit()` returns `false`.

```java
FormElement form = form(formState)
    .field("email", "Email", Validators.required(), Validators.email())
    .validateOnSubmit(true)  // default behavior
    .onSubmit(state -> {
        // Only called if ALL validations pass
        save(state);
    });

boolean success = form.submit();
// success == true  -> validation passed, onSubmit was called
// success == false -> validation failed, onSubmit was NOT called
```

To always call `onSubmit` regardless of validation:

```java
form(formState)
    .validateOnSubmit(false)  // skip validation
    .onSubmit(state -> save(state));  // always called
```

### Arrow Key Navigation

Enable arrow key navigation to move between fields:

```java
form(formState)
    .field("username", "Username")
    .field("email", "Email")
    .field("role", "Role", FieldType.SELECT)
    .arrowNavigation(true);  // Up/Down navigate between fields
```

Behavior:
- **Text fields**: Up/Down navigate to previous/next field (like Tab/Shift+Tab)
- **Boolean fields**: Up/Down navigate to previous/next field
- **Select fields**: Up/Down still change selection

## Complete Example

```java
public static class SettingsForm {

    private static final FormState FORM = FormState.builder()
        // Profile
        .textField("fullName", "Ada Lovelace")
        .textField("email", "ada@analytical.io")
        .textField("role", "Research")
        .textField("timezone", "UTC+1")
        // Preferences
        .textField("theme", "Nord")
        .booleanField("notifications", true)
        // Security
        .textField("twoFa", "Enabled")
        .build();

    public static Element render() {
        return column(
            panel("Profile", column(
                formField("Full name", FORM.textField("fullName"))
                    .labelWidth(14).rounded()
                    .borderColor(Color.DARK_GRAY)
                    .focusedBorderColor(Color.CYAN),
                formField("Email", FORM.textField("email"))
                    .labelWidth(14).rounded()
                    .borderColor(Color.DARK_GRAY)
                    .focusedBorderColor(Color.CYAN)
                    .validate(Validators.required(), Validators.email()),
                formField("Role", FORM.textField("role"))
                    .labelWidth(14).rounded()
                    .borderColor(Color.DARK_GRAY)
                    .focusedBorderColor(Color.CYAN)
            ).spacing(1)).rounded().borderColor(Color.CYAN),

            panel("Preferences", column(
                formField("Theme", FORM.textField("theme"))
                    .labelWidth(14).rounded()
                    .borderColor(Color.DARK_GRAY),
                formField("Notifications", FORM.booleanField("notifications"))
                    .labelWidth(14)
            ).spacing(1)).rounded().borderColor(Color.GREEN),

            row(
                text(" Save ").bold().black().onGreen(),
                text(" Cancel ").bold().white().bg(Color.DARK_GRAY)
            ).spacing(2)
        ).spacing(1).fill();
    }
}
```

## State Classes Reference

### TextInputState

```java
TextInputState state = new TextInputState("initial");

// Get/set text
String text = state.text();
state.setText("new value");

// Cursor operations
state.insert('c');
state.deleteBackward();
state.deleteForward();
state.moveCursorLeft();
state.moveCursorRight();
state.clear();
```

### BooleanFieldState

```java
BooleanFieldState state = new BooleanFieldState(false);

// Get/set value
boolean value = state.value();
state.setValue(true);

// Toggle
state.toggle();  // Flips the value
```

### SelectFieldState

```java
SelectFieldState state = new SelectFieldState("A", "B", "C");
// Or with List
SelectFieldState stateFromList = new SelectFieldState(options, 1);  // Select index 1

// Get values
String selected = state.selectedValue();  // "A"
int index = state.selectedIndex();        // 0
List<String> opts = state.options();   // ["A", "B", "C"]

// Change selection
state.selectIndex(2);    // Select "C"
state.selectNext();      // Move to next option (wraps)
state.selectPrevious();  // Move to previous option (wraps)
```

## Migration Guide

### Before

```java
// Individual state declarations
private static final TextInputState FULL_NAME = new TextInputState("Ada");
private static final TextInputState EMAIL = new TextInputState("ada@example.com");
private static final TextInputState ROLE = new TextInputState("Research");

// Custom helper
private static Element formRow(String label, TextInputState state) {
    return row(
        text(label).dim().length(14),
        textInput(state).rounded().borderColor(Color.DARK_GRAY).fill()
    ).spacing(1).length(3);
}

// Usage
Element beforeExample1 = formRow("Full name", FULL_NAME);
Element beforeExample2 = formRow("Email", EMAIL);
```

### After

```java
// Centralized state
private static final FormState FORM_STATE = FormState.builder()
    .textField("fullName", "Ada")
    .textField("email", "ada@example.com")
    .textField("role", "Research")
    .build();

// Usage - no helper needed
Element afterExample1 = formField("Full name", FORM_STATE.textField("fullName"))
    .labelWidth(14).rounded().borderColor(Color.DARK_GRAY);
Element afterExample2 = formField("Email", FORM_STATE.textField("email"))
    .labelWidth(14).rounded().borderColor(Color.DARK_GRAY);
```

---

**Version:** 0.2.0-SNAPSHOT
