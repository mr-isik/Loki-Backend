# Loki Backend API

A clean, scalable REST API built with Go, following SOLID principles and clean architecture patterns. This project provides a robust foundation for building workflow management systems with PostgreSQL integration.

## 🏗️ Architecture

This project follows **Clean Architecture** principles with clear separation of concerns.

### Architecture Layers

1. **Domain Layer** (`internal/domain/`)

   - Defines core business entities and interfaces
   - Contains repository and service interfaces
   - Domain-specific errors and types
   - **Zero dependencies** on other layers

2. **Repository Layer** (`internal/repository/`)

   - Implements data access logic
   - Handles database operations (CRUD)
   - Converts PostgreSQL errors to domain errors
   - Uses pgx/v5 for efficient database access

3. **Service Layer** (`internal/service/`)

   - Implements business logic
   - Orchestrates between repositories
   - Handles data validation and transformations
   - Password hashing and authentication logic

4. **Handler Layer** (`internal/handler/`)

   - HTTP request/response handling
   - Input validation
   - Error mapping to HTTP status codes
   - JSON serialization/deserialization

5. **Router Layer** (`internal/router/`)
   - Route configuration
   - Middleware setup (CORS, logging, recovery)
   - Dependency injection

## 🚀 Features

- ✅ **Clean Architecture** - SOLID principles, dependency injection
- ✅ **PostgreSQL** - Powerful relational database with pgx driver
- ✅ **Docker Support** - Containerized application with Docker Compose
- ✅ **Graceful Shutdown** - Proper cleanup on termination
- ✅ **Error Handling** - Custom domain errors, PostgreSQL error parsing
- ✅ **Password Security** - Bcrypt hashing
- ✅ **Soft Deletes** - User data preservation
- ✅ **Pagination** - Efficient data retrieval
- ✅ **Logging** - Request/response logging middleware

## 📋 Prerequisites

- Go 1.25 or higher
- Docker & Docker Compose
- PostgreSQL 16 (if running locally)

## 🛠️ Installation & Setup

### Using Docker (Recommended)

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd loki-backend
   ```

2. **Start the services**

   ```bash
   docker-compose up -d
   ```

   This will start:

   - PostgreSQL database on port 5432
   - Backend API on port 3000

3. **Run migrations**

   ```bash
   # For Windows PowerShell
   Get-Content migrations/001_create_users_table.sql | docker exec -i loki-postgres psql -U loki -d loki_db
   Get-Content migrations/002_create_workspaces_workflows.sql | docker exec -i loki-postgres psql -U loki -d loki_db

   # For Linux/Mac
   docker exec -i loki-postgres psql -U loki -d loki_db < migrations/001_create_users_table.sql
   docker exec -i loki-postgres psql -U loki -d loki_db < migrations/002_create_workspaces_workflows.sql
   ```

4. **Verify the setup**
   ```bash
   curl http://localhost:3000/health
   ```

### Local Development

1. **Install dependencies**

   ```bash
   go mod download
   ```

2. **Set up environment variables**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start PostgreSQL** (using Docker)

   ```bash
   docker-compose up -d postgres
   ```

4. **Run migrations** (see step 3 above)

5. **Run the application**
   ```bash
   go run cmd/main.go
   ```

## 🔧 Configuration

Environment variables can be configured in `.env`:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=loki
DB_PASSWORD=loki_password
DB_NAME=loki_db

# Application Configuration
PORT=:3000
```

## 📚 API Documentation

### Health Check

```http
GET /health
```

Response:

```json
{
  "status": "ok",
  "service": "loki-backend"
}
```

### Users

#### Create User

```http
POST /api/users
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "securePassword123"
}
```

#### Get User

```http
GET /api/users/:id
```

#### List Users (with pagination)

```http
GET /api/users?page=1&page_size=10
```

#### Update User

```http
PUT /api/users/:id
Content-Type: application/json

{
  "email": "newemail@example.com",
  "name": "Jane Doe"
}
```

#### Delete User (Soft Delete)

```http
DELETE /api/users/:id
```

### Workspaces

#### Create Workspace

```http
POST /api/workspaces
Content-Type: application/json

{
  "owner_user_id": "uuid",
  "name": "My Workspace"
}
```

#### Get Workspace

```http
GET /api/workspaces/:id
```

#### List User Workspaces

```http
GET /api/workspaces/user/:userId?page=1&page_size=10
```

#### Update Workspace

```http
PUT /api/workspaces/:id
Content-Type: application/json

{
  "name": "Updated Workspace Name"
}
```

#### Delete Workspace

```http
DELETE /api/workspaces/:id
```

### Workflows

#### Create Workflow

```http
POST /api/workflows
Content-Type: application/json

{
  "workspace_id": "uuid",
  "title": "My Workflow",
  "status": "draft"
}
```

Status options: `draft`, `published`, `archived`

#### Get Workflow

```http
GET /api/workflows/:id
```

#### List Workspace Workflows

```http
GET /api/workflows/workspace/:workspaceId?page=1&page_size=10
```

#### Update Workflow

```http
PUT /api/workflows/:id
Content-Type: application/json

{
  "title": "Updated Workflow",
  "status": "published"
}
```

#### Update Workflow Status

```http
PATCH /api/workflows/:id/status
Content-Type: application/json

{
  "status": "published"
}
```

#### Delete Workflow

```http
DELETE /api/workflows/:id
```

## 🗄️ Database Schema

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

### Workspaces Table

```sql
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Workflows Table

```sql
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    title VARCHAR NOT NULL DEFAULT 'Untitled Workflow',
    status workflow_status NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## 🐳 Docker Commands

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# View backend logs only
docker-compose logs -f backend

# Stop services
docker-compose down

# Stop and remove volumes (⚠️ deletes data)
docker-compose down -v

# Rebuild containers
docker-compose up -d --build

# Access PostgreSQL CLI
docker exec -it loki-postgres psql -U loki -d loki_db
```

## 📝 Error Handling

The application uses custom domain errors that are parsed from PostgreSQL errors:

- `ErrNotFound` - Resource not found
- `ErrAlreadyExists` - Resource already exists
- `ErrUniqueViolation` - Unique constraint violation
- `ErrForeignKeyViolation` - Foreign key constraint violation
- `ErrInvalidInput` - Invalid input data

PostgreSQL errors are automatically converted to clean domain errors using `ParseDBError()`.

## 🔐 Security

- Passwords are hashed using bcrypt with default cost (10)
- Sensitive data (passwords) are never exposed in API responses
- CORS middleware configured for cross-origin requests
- Input validation on all endpoints
- Prepared statements prevent SQL injection

## 🚦 Production Considerations

- [ ] Add authentication/authorization (JWT, OAuth)
- [ ] Implement rate limiting
- [ ] Add request ID tracking
- [ ] Set up structured logging (JSON format)
- [ ] Add metrics and monitoring (Prometheus)
- [ ] Implement API versioning
- [ ] Add input validation library (e.g., go-playground/validator)
- [ ] Set up CI/CD pipeline
- [ ] Add integration tests
- [ ] Configure TLS/HTTPS
- [ ] Implement database migration tool (golang-migrate)

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 👨‍💻 Author

Built with ❤️ by Ömer Faruk

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Express-inspired web framework
- [pgx](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit
- [godotenv](https://github.com/joho/godotenv) - Environment variable loader
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Password hashing
- [uuid](https://github.com/google/uuid) - UUID generation
