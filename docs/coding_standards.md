# Golang Coding Standards

This document contains the coding standards that the project should follow.

---

## 1. Guiding Principles

The goal of these standards is to produce Go code that is **readable, simple, and maintainable**. Code is read far more often than it is written. These guidelines ensure our codebase is uniform, idiomatic, and easy for any contributor to understand.

*   **Clarity over Brevity:** Code should be easy to understand.
*   **Simplicity:** Avoid unnecessary complexity. A straightforward implementation is better than a clever one.
*   **Consistency:** Adhering to the defined project structure and style ensures the codebase feels familiar and predictable.

---

## 2. Project Directory Structure

We will follow a standardized project layout to ensure scalability and a clear separation of concerns. This structure is based on the [Golang Standard Project Layout](https://github.com/golang-standards/project-layout).

### 2.1. Core Layout

```
/
├── cmd/
│   └── [app-name]/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── [domain]/
│   │   └── [service]/
│   ├── platform/
│   │   └── [postgres]/
│   └── api/
├── pkg/
├── api/
├── configs/
└── scripts/
```

### 2.2. Key Directory Descriptions

*   **/cmd**: The entry points to your application. Each subdirectory here corresponds to a single executable binary (e.g., `/cmd/server`, `/cmd/worker`). The `main.go` files in this directory should be minimal, responsible only for initializing and starting the application.

*   **/internal**: This is where all the private application code lives. **The majority of your project's code will be here.** The Go compiler enforces that code within `/internal` cannot be imported by other projects, making it the perfect place for your core business logic.

*   **/pkg**: This directory is for code that is safe to be imported and used by *external* applications. **Use this directory with caution.** For most backend services, which are self-contained applications, you may not need a `/pkg` directory at all. If you are not building a reusable public library, prefer `/internal`.

*   **/api**: Contains API specification files, such as OpenAPI/Swagger specs, JSON schemas, or protocol definition files.

*   **/configs**: Configuration file templates and default configs.

*   **/scripts**: Supporting scripts for the project (e.g., `build.sh`, `test.sh`).

---

## 3. Code Organization within `/internal`

While the directory structure provides the skeleton, the principles of good package design apply *within* the `/internal` directory.

### 3.1. Package Naming

*   **Single Purpose:** A package should have a single, clear purpose. The name should describe what it provides (e.g., `http`, `user`, `postgres`).
*   **Short and Descriptive:** Use short, lowercase, single-word names.
*   **No `util` or `helpers`:** Avoid generic package names. These names do not describe the package's purpose and often become a dumping ground for unrelated code, leading to circular dependencies. Group functions by their domain instead.

### 3.2. Dependency Injection and Application Design

To maintain testability and a clean architecture, we apply dependency injection principles within `/internal`:

1.  **Domain Logic (`/internal/app` or `/internal/[domain]`):**
    *   Define your core business logic, types, and interfaces here. For example, `internal/user/user.go` would contain `type User` and `type Store interface { ... }`.
    *   This part of the code defines *what* the application does, not *how*. It should have minimal external dependencies.

2.  **Implementations (`/internal/platform` or `/internal/[dependency]`):**
    *   The concrete *implementations* of your domain interfaces should be in separate packages, named after their core dependency.
    *   Example: The implementation of the `user.Store` interface for a PostgreSQL database should be in `/internal/postgres/user_store.go`.
    *   This "one package per dependency" rule isolates external services, prevents circular dependencies, and makes testing straightforward by allowing mock implementations.

---

## 4. Naming Conventions

*   **Use `camelCase`:** All names (variables, functions, etc.) should use `camelCase`. For exported names, use `PascalCase`. Do not use `snake_case`.
*   **Variable Names:** The length of a variable name should be proportional to its scope. Short names (`i`, `k`) are fine for short scopes. Longer-lived variables need more descriptive names.
*   **Avoid Repeating Package Name:** If a function is in the `log` package, name it `log.Info()`, not `log.LogInfo()`.
*   **No Getters/Setters:** A function to retrieve a user should be `user()` or `User()`, not `getUser()`.
*   **Interface Naming:** For single-method interfaces, use the method name with an "er" suffix (e.g., `Reader`). For multi-method interfaces, use a descriptive noun (e.g., `UserStore`).

---

## 5. Code Formatting and Style

*   **Use `goimports`:** Always use `goimports` to format code and organize imports. This is non-negotiable.
*   **Avoid `else`:** Structure `if` statements to handle the error case first and return early. This reduces nesting and improves readability.
*   **Avoid long lines:** Lines should be 120 characters or less.

---

## 6. Comments

*   **Full Sentences:** Write comments as complete sentences.
*   **Document Exports:** Every exported name must have a doc comment.
*   **Explain the "Why," Not the "What":** Use comments to explain non-obvious logic.

---

## 7. Error Handling

*   **Handle All Errors:** Never ignore an error. If you intentionally ignore one, make it explicit with the blank identifier: `_ = f.Close()`.
*   **Provide Context:** Wrap errors with `fmt.Errorf("...: %w", err)` to provide a clear error chain.

---

## 8. Testing

*   **Table-Driven Tests:** Use table-driven tests for testing multiple scenarios.
*   **Test Packages:** Place test files in the same package as the code they are testing (`_test.go` suffix).
*   **Mocks and Interfaces:** Use the dependency injection structure to easily mock external services for reliable unit tests.