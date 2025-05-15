# Technical Context

## Core Technologies
- **Go**: Backend language
- **PostgreSQL**: Primary database
- **SQL Migrations**: Schema management

## Development Setup
### Dependencies
- Go modules (go.mod)
- Database driver (to be added)

### Build System
- Makefile present (contents to be reviewed)
- Potential targets:
  - build
  - test
  - migrate
  - run

### Database
- Migration system in place
- Initial users table migration created
  - `000001_create_users_table.up.sql`
  - `000001_create_users_table.down.sql`

## Environment
- `.env.example` present for configuration
- Likely needs:
  - Database connection string
  - Server port
  - JWT secret
