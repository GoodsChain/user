# Technical Context

## Core Technologies
- **Go**: Backend language
- **PostgreSQL**: Primary database
- **SQL Migrations**: Schema management
- **sqlc**: Type-safe SQL to Go code generation

## Development Setup
### Dependencies
- Go modules (go.mod)
- Database driver (pgx/v5 via sqlc)

### Build System
- Makefile targets:
  - sqlc.generate: Generate type-safe Go code
  - sqlc.vet: Validate SQL queries
  - db.up/down: Database management
  - migrate.up/down: Schema migrations

### Database
- Migration system in place
- Initial users table migration created
  - `000001_create_users_table.up.sql`
  - `000001_create_users_table.down.sql`
- sqlc configuration:
  - sqlc.yaml defines code generation settings
  - Generated code in db/ directory
  - CRUD operations for users table

## Environment
- `.env.example` present for configuration
- Likely needs:
  - Database connection string
  - Server port
  - JWT secret

## sqlc Implementation
- Generated code includes:
  - models.go: User model struct
  - querier.go: Database operations interface
  - users.sql.go: CRUD implementations
- Type-safe queries for:
  - Create/Read/Update/Delete operations
  - User lookup by ID/email
  - Paginated user listing
