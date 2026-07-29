# Chat API — AI Agent Guide

## 📋 Project Overview

A modern, scalable **Chat Application API** built with Go. It provides RESTful endpoints for user management, group messaging (with role-based access), direct messaging, message read receipts, file uploads, and Docker deployment.

**Purpose**: Backend service for a chat application.

---

## 🏗️ Architecture

### Layered Pattern

```
Routes → Handlers → Services → Repositories → Database (PostgreSQL/GORM)
```

Tests use **SQLite in-memory** for isolation (see Testing Notes).

| Layer | Directory | Responsibility |
|---|---|---|
| **Routes** | `routes/routes.go` | HTTP endpoint definitions, route grouping |
| **Handlers** | `handler/*.go` | Request parsing, response formatting, input validation |
| **Services** | `service/*.go` | Business logic, validation rules, authorization checks |
| **Repositories** | `repository/*.go` | Data access, CRUD operations via GORM |
| **Models** | `models/*.go` | GORM entity definitions, domain types |
| **DB** | `db/db.go` | Database connection, auto-migration setup |

### Dependency Injection

All layers are wired together in **`main.go`** using constructor injection:

```go
dbConn := db.InitDB()
userRepo := repository.NewUserRepository(dbConn)
userSvc := service.NewUserService(userRepo)
userHandler := handler.NewUserHandler(userSvc)
r := routes.SetupRouter(userHandler, groupHandler, messageHandler, fileHandler)
r.Run(":8080")
```

---

## 🛠️ Tech Stack

| Technology | Version | Purpose |
|---|---|---|
| **Go** | 1.22 | Language |
| **Gin** | v1.10.0 | HTTP web framework |
| **GORM** | v1.25.7 | ORM |
| **PostgreSQL** | (via gorm.io/driver/postgres v1.5.7) | Production database |
| **SQLite** | (via gorm.io/driver/sqlite v1.5.7) | Test database (in-memory) |
| **Docker** | — | Containerization |
| **testify** | v1.9.0 | Test assertions + mocks |

---

## 📁 Project Structure

```
chat/
├── main.go                      # Entry point — wires dependencies, starts server
├── go.mod                       # Module: "chat", Go 1.22, PostgreSQL + SQLite drivers
├── go.sum                       # Dependency checksums
├── README.md                    # Full documentation
├── postman_collection.json      # Postman collection for testing
├── .golangci.yml                # Lint configuration
├── .gitignore                   # Ignores .idea/, chat.db, uploads/
│
├── Dockerfile                   # Multi-stage build (golang:1.22-alpine → alpine:3.19)
├── docker-compose.yml           # PostgreSQL 16 + app services with healthcheck
├── .dockerignore                # Build context exclusions
│
├── models/                      # GORM entity models
│   ├── user.go                  # User struct
│   ├── group.go                 # Group + GroupMember structs (roles: owner/admin/member)
│   └── message.go               # Message + Seen + File structs (types: text/image/file)
│
├── db/
│   └── db.go                    # PostgreSQL with env-based config, AutoMigrate all models
│
├── repository/                  # Data access layer (interfaces + implementations)
│   ├── user_repository.go       # UserRepository
│   ├── group_repository.go      # GroupRepository (includes IsMember, GetMemberRole)
│   ├── message_repository.go    # MessageRepository
│   ├── file_repository.go       # FileRepository (CreateFile, GetFileByID)
│   ├── setup_test.go            # SQLite in-memory test DB helper
│   ├── user_repository_test.go
│   ├── group_repository_test.go
│   └── message_repository_test.go
│
├── service/                     # Business logic layer (interfaces + implementations)
│   ├── user_service.go          # UserService
│   ├── group_service.go         # GroupService (role-based authorization checks)
│   ├── message_service.go       # MessageService (group membership validation)
│   ├── file_service.go          # FileService (disk storage + metadata)
│   ├── mock_repository_test.go  # Mock repos for service tests
│   ├── user_service_test.go
│   ├── group_service_test.go
│   └── message_service_test.go
│
├── handler/                     # HTTP handlers (Gin context)
│   ├── user_handler.go          # UserHandler — POST/GET /users, GET /users/:id
│   ├── group_handler.go         # GroupHandler — CRUD + member mgmt (requester_id checks)
│   ├── message_handler.go       # MessageHandler — send/get/delete/seen/unseen + file upload
│   ├── file_handler.go          # FileHandler — POST /upload (multipart)
│   ├── mock_service_test.go     # Mock services for handler tests
│   ├── user_handler_test.go
│   ├── group_handler_test.go
│   ├── message_handler_test.go
│   └── file_handler_test.go
│
└── routes/
    └── routes.go                # SetupRouter() — defines all routes + static file serving
```

---

## 🗄️ Database Schema (PostgreSQL via GORM AutoMigrate)

### Users
| Field | Type | Constraints |
|---|---|---|
| id | uint | PK, auto-increment |
| username | string | size:255, NOT NULL, UNIQUE |
| email | string | size:255, NOT NULL, UNIQUE |
| created_at | time.Time | auto |
| updated_at | time.Time | auto |
| deleted_at | gorm.DeletedAt | soft delete |

### Groups
| Field | Type | Constraints |
|---|---|---|
| id | uint | PK, auto-increment |
| name | string | size:255, NOT NULL, UNIQUE |
| created_at/updated_at/deleted_at | — | standard GORM timestamps |

**Relations**: HasMany Messages (GroupID FK), HasMany GroupMembers

### GroupMembers (composite PK)
| Field | Type | Constraints |
|---|---|---|
| group_id | uint | PK (composite) |
| user_id | uint | PK (composite) |
| role | string | size:50, NOT NULL, default:'member' (owner|admin|member) |
| created_at/updated_at | time.Time | auto |

### Messages
| Field | Type | Constraints |
|---|---|---|
| id | uint | PK |
| sender_id | uint | NOT NULL, FK → User |
| receiver_id | *uint | nullable — direct message target |
| group_id | *uint | nullable — group message target |
| type | MessageType (string) | size:50, NOT NULL (text|image|file) |
| content | string | type:text |
| file_id | *uint | nullable, FK → File |
| created_at/updated_at/deleted_at | — | standard GORM timestamps |

**Constraint**: Exactly one of `receiver_id` or `group_id` must be set (validated in service layer).

### Seen (read receipts — composite PK)
| Field | Type | Constraints |
|---|---|---|
| message_id | uint | PK (composite), FK → Message |
| user_id | uint | PK (composite), FK → User |
| seen_at | time.Time | autoCreateTime |

### Files
| Field | Type | Constraints |
|---|---|---|
| id | uint | PK |
| url | string | size:512, NOT NULL |
| size | int64 | nullable |
| type | string | size:100, nullable |
| created_at/updated_at/deleted_at | — | standard GORM timestamps |

---

## 🌐 API Endpoints (17 total)

Base URL: `http://localhost:8080`

### Health
| Method | Path | Description |
|---|---|---|
| GET | `/health` | Returns `{"status": "ok"}` |

### Users
| Method | Path | Description |
|---|---|---|
| POST | `/users` | Create user (body: `username`, `email` — email validated by Gin binding) |
| GET | `/users` | List all users |
| GET | `/users/:id` | Get user by ID |

### Groups
| Method | Path | Description |
|---|---|---|
| POST | `/groups` | Create group (body: `name`) |
| DELETE | `/groups/:id` | Soft-delete group + its members (body: `requester_id`; owner only) |
| POST | `/groups/:id/members` | Add member (body: `requester_id`, `user_id`, `role` optional; admin/owner only) |
| DELETE | `/groups/:id/members/:user_id` | Remove member (body: `requester_id`; admin/owner only, cannot remove owner) |
| GET | `/groups/:id/members` | List group members |

### Messages
| Method | Path | Description |
|---|---|---|
| POST | `/messages` | Send message (body: `sender_id`, `receiver_id` or `group_id`, `type`, `content`, `file_id`) |
| POST | `/messages/upload` | Upload file + send message in one step (multipart: `file`, `sender_id`, `receiver_id`/`group_id`, `content`) |
| GET | `/messages` | Get conversation (query: `receiver_id` or `group_id`, `limit` default 50, `offset` default 0) |
| POST | `/messages/:id/seen` | Mark message as seen (body: `user_id`) |
| DELETE | `/messages/:id` | Soft-delete message |
| GET | `/messages/unseen/:user_id` | Get unseen counts grouped by sender + group |

### File Upload
| Method | Path | Description |
|---|---|---|
| POST | `/upload` | Upload file (multipart: `file`). Returns file metadata with `id`, `url`, `size`, `type` |
| GET | `/uploads/*path` | Static file serving (serves uploaded files) |

### Unseen Count Response Format
```json
{
  "direct_messages": [
    { "user_id": 1, "username": "john_doe", "unseen_count": 5 }
  ],
  "group_messages": [
    { "group_id": 1, "group_name": "Project Team", "unseen_count": 4 }
  ],
  "total": 14
}
```

---

## 🔐 Authorization Rules

Since there is **no JWT authentication** yet, authorization uses a `requester_id` field passed in request bodies. In the future this would be extracted from the auth token.

| Action | Required Role | Endpoint |
|---|---|---|
| Add member | Owner or Admin | `POST /groups/:id/members` |
| Assign owner/admin role | Owner only | `POST /groups/:id/members` (with `role: "owner"` or `"admin"`) |
| Remove member | Owner or Admin | `DELETE /groups/:id/members/:user_id` |
| Remove the owner | Nobody (forbidden) | `DELETE /groups/:id/members/:user_id` (target is owner) |
| Delete group | Owner only | `DELETE /groups/:id` |
| Send group message | Member | `POST /messages` (with `group_id`) |

**Auth errors return HTTP 403 Forbidden.**

---

## 🧩 Key Conventions & Patterns

### 1. Interface-Based Design
Every layer defines an **interface** and a **private struct implementation**:
- `repository/` — `UserRepository`, `GroupRepository`, `MessageRepository`, `FileRepository`
- `service/` — `UserService`, `GroupService`, `MessageService`, `FileService`
- `handler/` — `UserHandler`, `GroupHandler`, `MessageHandler`, `FileHandler`
- Constructors: `New*Repository(db)`, `New*Service(repo)`, `New*Handler(service)`

### 2. Upsert via `clause.OnConflict`
Both `AddMember` (GroupMember) and `MarkMessageSeen` (Seen) use `clause.OnConflict` to upsert:
```go
r.db.Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "group_id"}, {Name: "user_id"}},
    DoUpdates: clause.AssignmentColumns([]string{"role", "updated_at"}),
}).Create(&member)
```

### 3. Soft Deletes
Models with `gorm.DeletedAt` support soft deletes. The `DeleteGroup` method also cascades to `GroupMember` records.

### 4. Message Types
Constants in `models/message.go`: `MessageTypeText = "text"`, `MessageTypeImage = "image"`, `MessageTypeFile = "file"`
File messages are auto-detected: images (Content-Type starts with `"image/"`) get type `"image"`, everything else gets type `"file"`.

### 5. Role Constants (GroupMember)
Defined in `models/group.go`: `RoleOwner = "owner"`, `RoleAdmin = "admin"`, `RoleMember = "member"`

### 6. Pagination
All conversation queries use `limit` (default 50) and `offset` (default 0) query parameters.

### 7. Error Handling
- HTTP 400: Bad request (validation errors, invalid conversation target)
- HTTP 403: Forbidden (authorization failures)
- HTTP 404: Resource not found
- HTTP 500: Internal server errors
- All errors returned as `{"error": "<message>"}`

### 8. Thread Safety
The package-level `DB` variable in `db/db.go` is read after initialization only.

---

## 🔧 Common Development Tasks

### Adding a New Feature

1. **Model** — Add struct in `models/` with GORM tags
2. **Repository** — Add interface + implementation in `repository/`
3. **Service** — Add interface + implementation in `service/`
4. **Handler** — Add Gin handler in `handler/`
5. **Routes** — Register endpoint in `routes/routes.go`
6. **Migrations** — Add model to `db.AutoMigrate()` in `db/db.go`
7. **Tests** — Add test files for each layer using established patterns

### Running Locally (requires PostgreSQL)
```bash
# Set env vars (or use defaults):
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=chat

go run main.go   # Starts on :8080
```

### Running with Docker
```bash
docker compose up --build   # Starts PostgreSQL + app
```

### Running Tests
```bash
go test ./...          # All tests (uses SQLite in-memory, no DB required)
go test ./... -v       # Verbose
go test ./... -cover   # Coverage
```

### Database (PostgreSQL)
- Connection configured via env vars: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- Auto-migration runs on every start
- Tests use isolated SQLite in-memory databases (no PostgreSQL needed)

### Linting
```bash
golangci-lint run   # Uses .golangci.yml config
```

### Adding Dependencies
```bash
go get <package>
go mod tidy
```

---

## 🔐 Validation Rules

| Field | Validation |
|---|---|
| username | Required, unique |
| email | Required, valid email format (Gin `binding:"email"`), unique |
| group name | Required, unique |
| sender_id | Required for messages |
| message type | Required (`text`, `image`, or `file`) |
| receiver_id / group_id | Exactly one must be set (validated in `SendMessage`) |
| requester_id | Required for group management endpoints |
| file | Required for `/upload` and `/messages/upload` |

---

## 🧪 Testing Approach

Tests use **SQLite in-memory** (`:memory:` with `MaxOpenConns(1)`) — no PostgreSQL needed.

### Test Coverage by Layer (100+ total subtests)

| Package | Test Files | Approach |
|---|---|---|
| `repository/` | 4 test files, 30 subtests | Real SQLite in-memory, full CRUD + edge cases |
| `service/` | 3 test files, 29 subtests | Mocked repositories via function-field mocks |
| `handler/` | 4 test files, ~50 subtests | Mocked services via function-field mocks, `httptest.NewRecorder()` |

### Service Test Mock Pattern
```go
type mockGroupRepo struct {
    addMemberFn    func(groupID, userID uint, role string) error
    getMemberRoleFn func(groupID, userID uint) (string, error)
    // ... each interface method has a function field
}
```
This pattern avoids import-time mocking frameworks and gives fine-grained control per test.

### Handler Test Pattern
```go
func setupMessageRoutes(svc *mockMsgSvc, fileSvc *mockFileSvc) *gin.Engine {
    gin.SetMode(gin.TestMode)
    h := NewMessageHandler(svc, fileSvc)
    r := gin.New()
    r.POST("/messages/upload", h.SendFileMessage)
    // ... register all routes
    return r
}
```
Tests send HTTP requests through Gin's test router and assert on response code + body.

---

## 📦 Key Dependencies (from go.mod)

| Module | Purpose |
|---|---|
| `github.com/gin-gonic/gin` v1.10.0 | HTTP router + middleware |
| `gorm.io/gorm` v1.25.7 | ORM |
| `gorm.io/driver/postgres` v1.5.7 | PostgreSQL driver for GORM (production) |
| `gorm.io/driver/sqlite` v1.5.7 | SQLite driver for GORM (tests only) |
| `github.com/stretchr/testify` v1.9.0 | Test assertions (`assert`, `require`) + mocks |

---

## 🐳 Docker Setup

### Dockerfile (multi-stage)
- **Build**: `golang:1.22-alpine` → `CGO_ENABLED=0` + `-ldflags="-s -w"` → small static binary
- **Runtime**: `alpine:3.19` with `ca-certificates` and `tzdata`
- Exposes port 8080, creates `/app/uploads` directory

### Docker Compose
```yaml
services:
  db:   # postgres:16-alpine with healthcheck
  app:  # builds from Dockerfile, depends on healthy db
```
- Persistent volumes: `pgdata` (database), `uploads` (files)
- `GIN_MODE=release` for production

---

## 🧠 AI Agent Notes

- **This is a Go module** named `chat` — all imports use `chat/...` prefix.
- **Gin context** (`*gin.Context`) is used for all HTTP handlers — use `c.ShouldBindJSON` for JSON, `c.PostForm`/`c.Request.FormFile` for multipart.
- **GORM** is used for all database operations — `Preload` for eager loading, `Clauses(clause.OnConflict{})` for upsert.
- **No authentication middleware** exists yet — `requester_id` is passed in request bodies as a placeholder.
- **No WebSocket** support — this is a REST-only API.
- **The `GetUnseenCountDetailed`** repository method returns function-local types inside `map[string]interface{}`. Tests handle this via JSON round-trip (`jsonRoundTrip` helper).
- **When modifying existing code**, look at **both the interface and the implementation** in the same file (repository, service layers).
- **Mock files** are in the same package as tests (suffixed `_test.go`) and use function-field mocks, not `testify/mock`.
- **Tests use SQLite in-memory**, not PostgreSQL — `repository/setup_test.go` provides the `setupTestDB()` helper.
- **File uploads** go to `./uploads/` directory (served statically at `/uploads/*`).

---
