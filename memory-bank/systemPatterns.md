# System Patterns

## Data Access Layer
- **sqlc Code Generation**:
  - SQL-first approach with type-safe Go bindings
  - Clear separation between SQL and application code
  - Generated models match database schema exactly

## Type Safety
- **Compile-time Validation**:
  - All queries validated at generation time
  - Parameter and return types strictly enforced
  - Database schema changes break builds if queries are invalid

## Database Operations
- **Interface Pattern**:
  - Querier interface provides all database operations
  - Easy mocking for testing
  - Clear contract between application and database layers

## Code Organization
- **Generated Code**:
  - models.go: Pure data structures
  - querier.go: Operation interfaces
  - [table].sql.go: Implementation details
- **Separation of Concerns**:
  - SQL files contain only queries
  - Go code handles business logic
  - Clear boundary between layers

## CRUD Patterns
- Standard operations for each table:
  - Create: Returns created record
  - Read: By ID or other unique fields
  - Update: Partial updates with COALESCE
  - Delete: Simple removal
  - List: Paginated results
