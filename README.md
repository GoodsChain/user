# User Management API

A RESTful API service for user management built with Go, Gin framework, and PostgreSQL.

![Go](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Latest-336791?style=flat&logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat&logo=docker)

## Prerequisites

Make sure you have the following installed:

- [Go](https://golang.org/dl/) 1.19 or higher
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) CLI tool
- Git

### Install golang-migrate

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Windows
scoop install migrate
```

## Quick Setup

Get the API running locally in 4 simple steps:

### 1. Clone and Configure

```bash
git clone <your-repo-url>
cd user
cp .env.example .env
```

### 2. Start PostgreSQL Database

```bash
make db.up
```

### 3. Run Database Migrations

```bash
make migrate.up
```

### 4. Start the Server

```bash
make run
```

🎉 **API is now running at `http://localhost:3000`**

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/users` | Get all users |
| `POST` | `/api/v1/users` | Create a new user |
| `GET` | `/api/v1/users/:id` | Get user by ID |
| `PATCH` | `/api/v1/users/:id` | Update user |
| `DELETE` | `/api/v1/users/:id` | Delete user |

### Example Usage

```bash
# Create a user
curl -X POST http://localhost:3000/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# Get all users
curl http://localhost:3000/api/v1/users

# Get user by ID
curl http://localhost:3000/api/v1/users/1
```

## Testing

Run the test suite:

```bash
make test
```

## Development Commands

### Database Management

```bash
# Stop database
make db.stop

# Stop and remove database container
make db.down

# Create new migration
make migrate.create

# Rollback last migration
make migrate.down
```

### Environment Configuration

The `.env` file contains all configuration variables:

```env
PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_NAME=user-database
DB_USER=postgre
DB_PASSWORD=postgre
DB_SSLMODE=disable
MIGRATIONS_DIR=db/migrations
CONTAINER_NAME=user-container
```

## Project Structure

```
.
├── main.go                 # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── db/                # Database connection
│   ├── handler/           # HTTP handlers
│   ├── models/            # Data models
│   ├── repository/        # Data access layer
│   └── router/            # Route definitions
├── db/migrations/         # Database migrations
└── Makefile              # Development commands
```

## Cleanup

When you're done developing:

```bash
make db.down  # Stops and removes the database container
```

---

**Happy coding!** 🚀
