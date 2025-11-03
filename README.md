# Chat API

A modern, scalable chat application API built with Go, featuring user management, group messaging, and real-time message tracking.

## 🚀 Features

- **User Management**: Create, retrieve, and list users with email validation
- **Group Messaging**: Create groups, manage members with role-based access (owner/admin/member)
- **Direct Messaging**: Send private messages between users
- **Message Types**: Support for text, image, and file messages
- **Read Receipts**: Track message read status with seen receipts
- **File Support**: Handle file uploads with metadata tracking
- **RESTful API**: Clean, intuitive endpoints following REST conventions
- **Database Persistence**: SQLite database with automatic migrations using GORM

## 🏗️ Architecture

This project follows a **layered architecture pattern**:

```
Routes → Handlers → Services → Repositories → Database
```

### Layers

- **Routes**: HTTP endpoint definitions and routing configuration
- **Handlers**: Request/response handling and input validation
- **Services**: Business logic and core application functionality
- **Repositories**: Data access layer with CRUD operations
- **Models**: Data structures and domain entities

## 🛠️ Tech Stack

- **Language**: Go 1.22
- **Web Framework**: [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- **ORM**: [GORM](https://gorm.io/) - Object-relational mapping for Go
- **Database**: SQLite - Embedded SQL database engine

### Dependencies

```
github.com/gin-gonic/gin v1.10.0     - Web framework
gorm.io/gorm v1.25.7                  - ORM framework
gorm.io/driver/sqlite v1.5.7          - SQLite driver for GORM
```

## 📋 Project Structure

```
chat/
├── main.go                           # Application entry point
├── go.mod                            # Go module definition
├── go.sum                            # Dependency checksums
├── chat.db                           # SQLite database (auto-generated)
│
├── models/                           # Data models
│   ├── user.go                       # User model
│   ├── group.go                      # Group and GroupMember models
│   └── message.go                    # Message, Seen, and File models
│
├── db/                               # Database initialization
│   └── db.go                         # GORM setup and migrations
│
├── repository/                       # Data access layer
│   ├── user_repository.go            # User CRUD operations
│   ├── group_repository.go           # Group CRUD operations
│   └── message_repository.go         # Message CRUD operations
│
├── service/                          # Business logic layer
│   ├── user_service.go               # User business logic
│   ├── group_service.go              # Group business logic
│   └── message_service.go            # Message business logic
│
├── handler/                          # HTTP request handlers
│   ├── user_handler.go               # User endpoints
│   ├── group_handler.go              # Group endpoints
│   └── message_handler.go            # Message endpoints
│
├── routes/                           # Route configuration
│   └── routes.go                     # Route setup
│
├── postman_collection.json           # Postman API collection
└── README.md                         # This file
```

## 📦 Installation

### Prerequisites

- Go 1.22 or higher
- Git

### Setup

1. **Clone the repository** (if applicable):
   ```bash
   git clone <repository-url>
   cd chat
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build the application**:
   ```bash
   go build
   ```

## 🚀 Running the Application

### Start the Server

```bash
go run main.go
```

The API server will start on `http://localhost:8080`

You should see output like:
```
Database initialized and migrations completed successfully
[GIN-debug] Loaded HTML Templates (0): 
[GIN-debug] Registered 14 routes
```

### Health Check

Verify the server is running:
```bash
curl http://localhost:8080/health
```

Response:
```json
{"status": "ok"}
```

## 📡 API Endpoints

Base URL: `http://localhost:8080`

For detailed request/response examples and to test endpoints interactively, use the included **Postman collection** (`postman_collection.json`).

### User Endpoints
- `POST /users` - Create a new user
- `GET /users` - Get all users
- `GET /users/:id` - Get user by ID

### Group Endpoints
- `POST /groups` - Create a new group
- `DELETE /groups/:id` - Delete a group
- `POST /groups/:id/members` - Add member to group
- `DELETE /groups/:id/members/:user_id` - Remove member from group
- `GET /groups/:id/members` - Get group members

### Message Endpoints
- `POST /messages` - Send a message (direct or group)
- `GET /messages` - Get conversation history
- `POST /messages/:id/seen` - Mark message as seen
- `DELETE /messages/:id` - Delete a message
- `GET /messages/unseen/:user_id` - Get unseen message count

## 🗄️ Database Schema

The application uses SQLite with the following tables:

### Users
```sql
users (
  id INTEGER PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

### Groups
```sql
groups (
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

### GroupMembers
```sql
group_members (
  group_id INTEGER PRIMARY KEY,
  user_id INTEGER PRIMARY KEY,
  role TEXT DEFAULT 'member',
  created_at TIMESTAMP,
  updated_at TIMESTAMP
)
```

### Messages
```sql
messages (
  id INTEGER PRIMARY KEY,
  sender_id INTEGER NOT NULL,
  receiver_id INTEGER,
  group_id INTEGER,
  type TEXT NOT NULL,
  content TEXT,
  file_id INTEGER,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

### Seen (Read Receipts)
```sql
seens (
  message_id INTEGER PRIMARY KEY,
  user_id INTEGER PRIMARY KEY,
  seen_at TIMESTAMP
)
```

### Files
```sql
files (
  id INTEGER PRIMARY KEY,
  url TEXT NOT NULL,
  size INTEGER,
  type TEXT,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
)
```

## 🧪 Testing

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test ./... -v
```

Run tests with coverage:
```bash
go test ./... -cover
```

Generate coverage report:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🔍 Error Handling

The API returns standard HTTP status codes:

- `200 OK` - Request succeeded
- `201 Created` - Resource created successfully
- `204 No Content` - Resource deleted successfully
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server-side error

## 🚀 Using Postman

A Postman collection is included (`postman_collection.json`). To use it:

1. Open Postman
2. Click **Import** → **Upload Files**
3. Select `postman_collection.json`
4. Select the collection and start testing endpoints

## 📝 Development Guidelines

### Adding New Features

1. **Define Models** in `models/`
2. **Create Repository** in `repository/` for database operations
3. **Implement Service** in `service/` for business logic
4. **Build Handler** in `handler/` for HTTP endpoints
5. **Register Routes** in `routes/routes.go`
6. **Write Tests** alongside implementation

### Code Style

- Follow Go conventions ([Effective Go](https://golang.org/doc/effective_go))
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused

## 🐛 Troubleshooting

### Database Lock Error
If you encounter "database is locked":
- Ensure only one instance of the app is running
- Delete `chat.db` and restart the application

### Port Already in Use
If port 8080 is already in use:
- Edit `main.go` line 34 to use a different port: `r.Run(":8081")`
- Restart the application

### Dependency Issues
Resolve dependency issues with:
```bash
go mod tidy
go mod download
```

## 📄 License

This project is provided as-is for educational and development purposes.

## 👥 Contributing

Contributions are welcome! Please ensure:
- Code follows Go conventions
- Tests are included for new features
- Commit messages are clear and descriptive

## 📞 Support

For issues or questions, please check:
1. The API endpoints documentation above
2. The Postman collection for example requests
3. Application logs for error details
