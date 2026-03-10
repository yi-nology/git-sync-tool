# Git Manage Service - Project Structure

This document provides a comprehensive overview of the project's directory structure and architectural layers. It is intended to help developers and AI assistants quickly understand the codebase organization.

## 1. Directory Overview

| Directory | Description |
| :--- | :--- |
| `biz/` | **Core Business Logic**. Contains Handlers, Services, DAL, and Models. |
| `cmd/` | **Entry Points**. Main applications for API and RPC servers. |
| `pkg/` | **Shared Packages**. Common utilities, configs, and helper functions. |
| `idl/` | **Interface Definitions**. Protocol Buffer files (.proto) for RPC. |
| `biz/kitex_gen/` | **Generated Code**. Auto-generated Go code from IDL (Kitex). |
| `conf/` | **Configuration**. YAML config files for different environments. |
| `public/` | **Frontend Assets**. Static HTML, CSS, and JS files for the web UI. |
| `deploy/` | **Deployment**. Kubernetes manifests and Docker configs. |
| `docs/` | **Documentation**. API Swagger docs and product manuals. |

---

## 2. Business Architecture (`biz/`)

The `biz/` directory follows a layered architecture: **Handler -> Service -> DAL**.

### 2.1 Handler Layer (`biz/handler/`)
Responsible for handling HTTP requests, input validation, and response formatting.
*   **`repo.go`**: Repository registration, scanning, and CRUD operations.
*   **`sync_task.go`**: Management of synchronization tasks (Create, List, Run).
*   **`branch.go`**: Branch management (List, Create, Delete).
*   **`audit.go`**: Retrieval of system audit logs.
*   **`system.go`**: System utilities (File browser, SSH keys).

### 2.2 Service Layer (`biz/service/`)
Contains the core business logic and orchestration. Services are reusable components.
*   **`git/`**: Encapsulates raw Git operations.
    *   `git_service.go`: Core Git commands (Fetch, Push, Clone).
    *   `git_branch.go`: Branch operations logic.
*   **`sync/`**: Logic for repository synchronization.
    *   `sync_service.go`: Execution logic for sync tasks.
    *   `cron_service.go`: Scheduled task management.
*   **`audit/`**: Service for recording user actions.
*   **`stats/`**: Statistics calculation (Commit history, Code lines).

### 2.3 Data Access Layer (`biz/dal/`)
Responsible for direct database interactions using GORM.
*   **`db/`**:
    *   `init.go`: Database connection and schema migration.
    *   `repo_dao.go`: Database operations for `Repo` entity.
    *   `sync_task_dao.go`: Database operations for `SyncTask` entity.
    *   `audit_log_dao.go`: Database operations for `AuditLog` entity.

### 2.4 Model Layer (`biz/model/`)
Defines data structures used across layers.
*   **`api/`**: **Web Layer Models**. Request/Response structs and DTOs (Data Transfer Objects).
    *   `repo_req.go`, `task_req.go`: Input request validation structs.
    *   `repo_dto.go`, `task_dto.go`: Output DTOs for API responses.
*   **`domain/`**: **Domain Models**. Internal business objects not tied to DB or API.
    *   `git.go`: `GitBranch`, `GitRemote` structures.
    *   `common.go`: Shared value objects like `AuthInfo`.
*   **`po/`**: **Persistent Objects**. Structs mapped directly to database tables (GORM).
    *   `repo.go`: `repos` table.
    *   `sync_task.go`: `sync_tasks` table.
    *   `sync_run.go`: `sync_runs` table.

### 2.5 RPC Handler (`biz/rpc_handler/`)
Implements the gRPC/Kitex interfaces defined in `biz/kitex_gen`.
*   **`git_handler.go`**: Handles RPC requests for Git services.

---

## 3. Entry Points (`cmd/`)

*   **`api/main.go`**: Starts the HTTP API server (Hertz framework).
*   **`rpc/main.go`**: Starts the RPC server (Kitex framework).
*   **`all/main.go`**: (Optional) Combined entry point.

---

## 4. Shared Packages (`pkg/`)

*   **`configs/`**: Global configuration loader (`config.yaml`).
*   **`response/`**: Standard API response wrappers (Success, Error).
*   **`errno/`**: Application-specific error codes.

---

## 5. Deployment (`deploy/`)

*   **`docker-compose.yml`**: For local development or simple deployment.
*   **`k8s/`**: Kubernetes resources (`deployment.yaml`, `service.yaml`).
*   **`bootstrap.sh`**: Startup script for initialization.

## 6. Key Workflows

### Repository Synchronization Flow
1.  **User** creates a task via API (`POST /api/sync/tasks`).
2.  **Handler** (`biz/handler/sync_task.go`) validates input and calls DAO.
3.  **CronService** (`biz/service/sync/cron_service.go`) schedules the task.
4.  **SyncService** (`biz/service/sync/sync_service.go`) executes the logic:
    *   Calls `GitService` to Fetch source.
    *   Calls `GitService` to Push to target.
    *   Records result in `SyncRun` table via DAO.

### Branch Management Flow
1.  **User** requests branch list (`GET /api/repos/:key/branches`).
2.  **Handler** (`biz/handler/branch.go`) gets Repo path from DB.
3.  **GitService** (`biz/service/git/git_branch.go`) reads local Git refs.
4.  **Handler** formats the output as JSON response.
