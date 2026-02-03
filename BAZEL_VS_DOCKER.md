# Bazel vs Docker vs Make

This document explains the roles, benefits, and differences between the three tools used in this project: **Bazel**, **Docker**, and **Make**.

## At a Glance

| Feature | **Bazel** | **Docker** | **Make** |
| :--- | :--- | :--- | :--- |
| **Primary Role** | **Build System** | **Runtime Environment** | **Task Runner** |
| **Focus** | How strictly and efficiently to build code. | Where and how the code runs. | How to orchestrate commands. |
| **Hermeticity** | High (Builds are isolated from host). | High (Runtime is isolated). | Low (Dependent on host tools). |
| **Caching** | Advanced (Artifact-level caching). | Layer Caching (Image layers). | Basic (File timestamps). |
| **Best For** | Large, polyglot monorepos. | Deployment, reproducible environments. | Local dev shortcuts, simple application workflows. |

---

## 1. Bazel (The Builder)
**"Build logic that is correct, reproducible, and fast."**

Bazel is a build system developed by Google. It treats your build like a pure function: `Input + Config = Output`.

### Benefits
*   **Hermetic Builds**: Bazel builds code in isolated sandboxes. It doesn't care what libraries you have installed on your Mac vs Linux machine; it downloads its own toolchains (like the Go SDK we configured). This eliminates "works on my machine" build errors.
*   **Incremental Builds**: It tracks the exact hash of every file. If you change one comment in a proto file, Bazel knows *exactly* which Go files need recompilation and which do not.
*   **Parallelism**: It builds independent targets in parallel automatically.
*   **Polyglot**: It can build Go, Java, C++, and Protobufs in a single graph, handling cross-language dependencies (like generating Go code from `.proto` files) seamlessly.

### In This Project
We use Bazel to:
1.  Download the exact Go version (`1.23.0`).
2.  Compile Protocol Buffers into Go code (`rules_go` + `rules_proto`).
3.  Compile the User and Expense services.
4.  Link C libraries (like `zlib` and `sqlite3`) correctly across platforms.

---

## 2. Docker (The Container)
**"Run anywhere, exactly as built."**

Docker provides a standardized *runtime* environment. While Bazel ensures the *binary* is built correctly, Docker ensures the *OS* it runs on is consistent.

### Benefits
*   **Isolation**: The application runs in a Linux container with its own file system. It doesn't clutter your host OS.
*   **Portability**: A Docker image runs the same on your MacBook, a Windows PC, or a Kubernetes cluster in the cloud.
*   **Dependency Management**: It packages system libraries (like `glibc`, `sqlite` CLI, certificates) along with your application binary.

### In This Project
We use Docker to:
1.  Package the binaries produced (by Bazel or Go Build) into lightweight images (`alpine`).
2.  Orchestrate the full stack (User Service + Expense Service + Client) using `docker-compose`.

---

## 3. Make (The Orchestrator)
**"Shortcuts for complex commands."**

Make is a task runner. It doesn't know *how* to compile Go code explicitly (unless you tell it); it simply runs shell commands.

### Benefits
*   **Simplicity**: Provides a simple interface (`make build`, `make run`) for developers.
*   **Glue**: It allows you to chain tools together. For example, a `make deploy` command might run `bazel build`, then `docker build`, then `kubectl apply`.

### In This Project
We use Make means to:
1.  Provide easy aliases for long Bazel commands (e.g., `make clean` -> `bazel clean --expunge`).
2.  Bridge the gap between developers and the build system.

---

## Summary: How They Work Together

1.  **Make** kicks off the process (`make release`).
2.  **Bazel** compiles the source code into a binary, ensuring it uses the correct compiler versions and dependencies.
3.  **Docker** takes that binary and wraps it into a container image.
4.  **Docker Compose** runs those containers to start your application.
