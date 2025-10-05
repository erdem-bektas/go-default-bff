# Database Integration Tests

This directory contains comprehensive database integration tests for the Zitadel authentication system.

## Test Coverage

### 1. User Repository Tests (`database_integration_test.go`)

**TestDatabaseConnection**
- Tests basic database connectivity
- Verifies database health checks
- Tests connection pool configuration
- Validates simple query execution

**TestUserRepository_BasicOperations**
- Tests user creation with Zitadel ID
- Tests user retrieval by Zitadel ID
- Tests user updates (email, verification status)
- Validates proper handling of nullable email fields for social logins

**TestUserRepository_RoleOperations**
- Tests role assignment to users
- Tests role replacement (updating existing roles)
- Validates normalized user-role relationship
- Tests multi-tenant role context (org_id, project_id)

**TestUserRepository_ListUsers**
- Tests user listing with pagination
- Tests filtering by organization ID
- Tests filtering by project ID
- Validates query performance with proper indexing

**TestUserRepository_DeleteUser**
- Tests user deletion
- Validates cascade deletion of associated roles
- Tests foreign key constraint handling

### 2. Migration Runner Tests (`migration_test.go`)

**TestMigrationRunner_GetMigrationStatus**
- Tests migration status retrieval
- Validates pending migration detection
- Tests version tracking

**TestMigrationRunner_RunMigrations**
- Tests migration execution
- Validates idempotent migration runs
- Tests migration completion status

**TestMigrationRunner_ValidateSchema**
- Tests schema validation functionality
- Validates database structure compliance

**TestMigrationRunner_MigrationVersionParsing**
- Tests migration filename parsing
- Validates version number extraction
- Tests edge cases with invalid filenames

### 3. Database Health Tests (`health_test.go`)

**TestDatabaseHealthCheck**
- Tests database health check functionality
- Validates connection pool statistics
- Tests ping functionality

**TestDatabaseHealthWithClosedConnection**
- Tests health check behavior with closed connections
- Validates error handling and reporting

**TestMigrationRunnerHealthIntegration**
- Tests integration between migration runner and health checks
- Validates schema validation in health context

## Test Features

### Database Schema Management
- **Auto-migration**: Uses GORM AutoMigrate for test table creation
- **Clean State**: Drops and recreates tables for each test run
- **Constraint Handling**: Ensures proper foreign key constraints with CASCADE DELETE

### Test Data Isolation
- **Prefixed IDs**: Uses `test_` prefix for all test data
- **Cleanup**: Automatic cleanup of test data after each test
- **No Interference**: Tests don't interfere with production data

### Error Handling
- **Graceful Skipping**: Skips tests if database is unavailable
- **Meaningful Errors**: Provides clear error messages for debugging
- **Edge Cases**: Tests various error conditions and edge cases

## Requirements Covered

This test suite addresses the requirements from task 2.3:

### ✅ Test user creation and role assignment
- User creation with various field combinations
- Role assignment and management
- Multi-tenant role context validation
- Cascade deletion testing

### ✅ Test migration runner functionality
- Migration status tracking
- Idempotent migration execution
- Schema validation
- Version parsing and management

### ✅ Test database health checks
- Connection health validation
- Connection pool statistics
- Error condition handling
- Integration with migration status

## Running the Tests

```bash
# Run all database integration tests
go test ./test -v

# Run specific test categories
go test ./test -v -run TestUserRepository
go test ./test -v -run TestMigrationRunner
go test ./test -v -run TestDatabaseHealth

# Run migration and health tests
go test ./pkg/database -v
```

## Test Environment

The tests use the following environment variables (with defaults):

- `DB_HOST`: localhost
- `DB_PORT`: 5432
- `DB_USER`: postgres
- `DB_PASSWORD`: postgres
- `DB_NAME`: fiber_app
- `DB_SSLMODE`: disable

## Notes

- Tests require a running PostgreSQL instance
- Tests will drop and recreate `users` and `user_roles` tables
- Migration tests work with existing migration records (idempotent)
- All test data uses `test_` prefixes for easy identification and cleanup