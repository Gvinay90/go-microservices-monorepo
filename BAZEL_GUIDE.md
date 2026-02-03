# Bazel Guide & Cheat Sheet

This document serves as the primary reference for the **active** Bazel build system in this project.

## ✅ Current Status: **Active & Supported**

The project is fully configured to use Bazel for building and running all services.

---

## 🚀 Quick Start: Run Everything

### Option 1: Run All Services Together

To build and run all services (User, Expense, and Client) in a single command:

```bash
# Build everything first
bazel build //...

# Run services in background and client in foreground
(bazel run //services/user:user_service & \
 bazel run //services/expense:expense_service & \
 sleep 5 && \
 bazel run //client:client --action_env=USER_SERVICE_ADDR=localhost:50051 --action_env=EXPENSE_SERVICE_ADDR=localhost:50052)
```

To clean up background processes:
```bash
killall user_service expense_service
```

### Option 2: Run Services Individually

**Terminal 1 - User Service:**
```bash
bazel run //services/user:user_service
```

**Terminal 2 - Expense Service:**
```bash
# Note: Configure port in config/base.yaml or use env var
APP_SERVER_PORT=:50052 bazel run //services/expense:expense_service
```

**Terminal 3 - Client:**
```bash
bazel run //client:client \
  --action_env=USER_SERVICE_ADDR=localhost:50051 \
  --action_env=EXPENSE_SERVICE_ADDR=localhost:50052
```

### Option 3: Run Pre-Built Binaries

After building with `bazel build //...`, you can run binaries directly:

```bash
# Run User Service
./bazel-bin/services/user/user_service_/user_service

# Run Expense Service (in another terminal)
APP_SERVER_PORT=:50052 ./bazel-bin/services/expense/expense_service_/expense_service

# Run Client (in another terminal)
USER_SERVICE_ADDR=localhost:50051 \
EXPENSE_SERVICE_ADDR=localhost:50052 \
./bazel-bin/client/client_/client
```

---

## 🛠️ Bazel Cheat Sheet

### Building
| Command | Description |
| :--- | :--- |
| `bazel build //...` | Build **everything** in the workspace. |
| `bazel build //services/user:all` | Build all targets in the User service package. |
| `bazel build //services/user:user_service` | Build specifically the User Service binary. |
| `bazel build //client:client` | Build the Client binary. |

### Running
| Command | Description |
| :--- | :--- |
| `bazel run //services/user:user_service` | Run the User Service. |
| `bazel run //client:client` | Run the Client (defaults to expected ports). |
| `bazel run //:gazelle` | Regenerate `BUILD` files after adding new source code. |
| `bazel run //:gazelle-update-repos` | Update `deps.bzl` after changing `go.mod`. |

### Maintenance
| Command | Description |
| :--- | :--- |
| `bazel clean` | Delete the `bazel-bin` and `bazel-out` directories. |
| `bazel clean --expunge` | **Deep clean**. Removes the entire external cache (forces re-download). Use this if build is stuck. |
| `bazel query "deps(//services/user:user_service)"` | Graph query: Show all dependencies of the User Service. |

---

## 🔧 Build Configuration & Troubleshooting History

We encountered and resolved several critical issues to enable Bazel on this project (specifically on macOS).

### 1. The macOS `zlib` / `fdopen` Conflict
**Issue**: Compilation of `zlib` (a dependency of Protobuf) failed on macOS with `error: expected identifier` near `fdopen`. This was due to a conflict between the system headers and the internal configuration of older zlib versions bundled with `rules_proto_grpc`.
**Resolution**:
*   We explicitly overrode the `zlib` repository in `WORKSPACE` to use **version 1.3.1**.
*   This newer version correctly handles the `_POSIX_C_SOURCE` macros on macOS.

### 2. Bazel 7 & Bzlmod Compatibility
**Issue**: Bazel 7 enables **Bzlmod** (the new dependency management system) by default. However, our configuration relies on the legacy `WORKSPACE` file and `gazelle` macros. Mixing these caused errors like `Label '@@rules_proto...' is invalid`.
**Resolution**:
*   Created `.bazelrc` with `common --noenable_bzlmod`.
*   Downgraded `rules_go` to **v0.42.0** and `rules_proto` to **v5.3.0** to ensure strict compatibility with the legacy Workspace system.

### 3. Checksum Mismatches
**Issue**: `bazel sync` failed because the SHA256 checksums in `deps.bzl` did not match the downloaded artifacts for `googleapis_rpc`.
**Resolution**:
*   Manually corrected the checksums in `deps.bzl`.
*   Ran `bazel run //:gazelle-update-repos` to allow Gazelle to manage the rest.

### 4. Redundant Proto Arguments
**Issue**: `bazel build` failed with `Either proto or protos... argument must be specified, but not both`. This was caused by Gazelle generating `BUILD` files that contained both the legacy `proto=` and new `protos=` attributes.
**Resolution**:
*   Fixed `proto/user/BUILD.bazel` and `proto/expense/BUILD.bazel` by removing the redundant `proto` attribute.

### 5. Runtime Configuration
**Issue**: Services crashed when running via `bazel run` because they couldn't find `config/base.yaml`.
**Resolution**:
*   Added a `filegroup` for config files in the root `BUILD` file.
*   Added `data = ["//:config_files"]` to the `go_binary` rules, ensuring these files are available in the sandbox at runtime.

---


## ⚙️ Configuration

Services read configuration from `config/base.yaml` and environment variables.

**Key Configuration:**
- `server.port`: Service port (default `:50051`)
- `database.driver`: `sqlite` or `postgres`
- `database.dsn`: Database connection string
- `log.level`: `debug`, `info`, `warn`, `error`
- `log.format`: `text` or `json`

**Environment Variable Override:**
```bash
APP_SERVER_PORT=:50052 \
APP_DATABASE_DRIVER=postgres \
APP_DATABASE_DSN="host=localhost user=postgres password=secret dbname=expenses" \
bazel run //services/expense:expense_service
```

## 🐛 Troubleshooting

### Services Can't Connect
**Problem**: Client shows `connection refused`
**Solution**: Ensure services are running and ports match:
```bash
# Check if services are listening
lsof -i :50051  # User service
lsof -i :50052  # Expense service
```

### Database Errors
**Problem**: `failed to connect to database`
**Solution**: Check database configuration in `config/base.yaml`. For SQLite (default), ensure write permissions in the project directory.

### Bazel Build Fails with "proto" Error
**Problem**: `Either proto or protos argument must be specified, but not both`
**Solution**: This is a known issue with Gazelle regenerating BUILD files. Manually remove the `proto = ":xxx_proto"` line from `proto/*/BUILD.bazel` files, keeping only `protos = [":xxx_proto"]`.

### Port Already in Use
**Problem**: `bind: address already in use`
**Solution**: 
```bash
# Find and kill process using the port
lsof -ti:50051 | xargs kill -9
```

---

## Architecture Note
The current Bazel setup uses **standard Go toolchains**. It downloads a hermetic Go SDK (currently configured to **Go 1.23.0** via `WORKSPACE`), ensuring that all developers build with the exact same Go version regardless of what is installed on their local machine.
