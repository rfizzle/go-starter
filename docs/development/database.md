# Database Patterns Guide

Comprehensive guide for database development, migration strategies, and repository patterns in the Go Starter project.

## Overview

This project uses PostgreSQL with a **generic storage abstraction layer**. The storage layer remains business-agnostic, while domain-specific structs and repository logic are implemented within their respective domain packages (e.g., `internal/user/`, `internal/order/`).

## Architecture

### Layered Architecture

```
┌─────────────────────────────────┐
│     Controllers (HTTP)          │
├─────────────────────────────────┤
│     Services (Business Logic)   │
├─────────────────────────────────┤
│     Domain Repositories         │  ← internal/user/repository.go
├─────────────────────────────────┤
│     Generic Storage Driver      │  ← internal/storage
├─────────────────────────────────┤
│     PostgreSQL Implementation   │  ← internal/storage/postgres
├─────────────────────────────────┤
│     Database (PostgreSQL)       │
└─────────────────────────────────┘
```

### Key Components

- **Generic Storage Driver** (`internal/storage/`) - Database abstraction without business logic
- **PostgreSQL Implementation** (`internal/storage/postgres/`) - Concrete database implementation
- **Domain Repositories** (`internal/<domain>/repository.go`) - Business-specific data operations
- **Domain Models** (`internal/<domain>/types.go`) - Business entities and structs
- **Migrations** (`migrations/`) - Schema version control
- **Connection Management** - Connection pooling and lifecycle

## Generic Storage Driver Pattern

### Interface Definition

```go
// internal/storage/storage.go
package storage

import "context"

// Scanner interface for query results
type Scanner interface {
    ScanOne(dst interface{}) error
    ScanAll(dst interface{}) error
}

// Client interface for database operations
type Client interface {
    Query(sql string, args ...interface{}) (Scanner, error)
    Execute(sql string, args ...interface{}) error
    
    // Transaction support
    BeginTx(ctx context.Context) (Transaction, error)
    
    // Health check
    Ping(ctx context.Context) error
    
    // Cleanup
    Close() error
}

// Transaction interface
type Transaction interface {
    Query(sql string, args ...interface{}) (Scanner, error)
    Execute(sql string, args ...interface{}) error
    Commit() error
    Rollback() error
}
```

### No Domain Models in Storage Layer

The storage layer **does not contain business domain models**. Instead, it provides generic interfaces for:
- Executing SQL queries
- Scanning results into any struct
- Transaction management
- Connection lifecycle

**Domain-specific structs and queries belong in domain packages.**

### PostgreSQL Implementation

```go
// internal/storage/postgres/postgres.go
package postgres

import (
    "context"
    "fmt"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rfizzle/go-starter/internal/storage"
)

type Client struct {
    pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (storage.Client, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }
    
    // Configure connection pool
    config.MaxConns = 25
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = time.Minute * 30
    
    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }
    
    return &Client{pool: pool}, nil
}

// Generic query execution - no business logic
func (c *Client) Query(sql string, args ...interface{}) (storage.Scanner, error) {
    rows, err := c.pool.Query(context.Background(), sql, args...)
    if err != nil {
        return nil, fmt.Errorf("query failed: %w", err)
    }
    
    return &scanner{rows: rows}, nil
}

func (c *Client) Execute(sql string, args ...interface{}) error {
    _, err := c.pool.Exec(context.Background(), sql, args...)
    return err
}
```

## Domain Repository Pattern

### User Domain Example

```go
// internal/user/types.go
package user

import "time"

// User represents the business entity
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Name      string    `json:"name" db:"name"`
    Role      Role      `json:"role" db:"role"`
    Active    bool      `json:"active" db:"active"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Role string

const (
    RoleUser  Role = "user"
    RoleAdmin Role = "admin"
)

type Filter struct {
    Email  *string
    Role   *Role
    Active *bool
    Limit  int
    Offset int
}
```

### Repository Implementation

```go
// internal/user/repository.go
package user

import (
    "context"
    "fmt"
    
    "github.com/rfizzle/go-starter/internal/storage"
)

type Repository struct {
    db storage.Client
}

func NewRepository(db storage.Client) *Repository {
    return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *User) error {
    query := `
        INSERT INTO users (email, name, role, active)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at
    `
    
    scanner, err := r.db.Query(query, user.Email, user.Name, user.Role, user.Active)
    if err != nil {
        return fmt.Errorf("create user: %w", err)
    }
    
    return scanner.ScanOne(&struct {
        ID        *string    `db:"id"`
        CreatedAt *time.Time `db:"created_at"`
        UpdatedAt *time.Time `db:"updated_at"`
    }{
        ID:        &user.ID,
        CreatedAt: &user.CreatedAt,
        UpdatedAt: &user.UpdatedAt,
    })
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
    query := `
        SELECT id, email, name, role, active, created_at, updated_at
        FROM users
        WHERE id = $1
    `
    
    scanner, err := r.db.Query(query, id)
    if err != nil {
        return nil, fmt.Errorf("get user: %w", err)
    }
    
    var user User
    if err := scanner.ScanOne(&user); err != nil {
        return nil, fmt.Errorf("scan user: %w", err)
    }
    
    return &user, nil
}

func (r *Repository) List(ctx context.Context, filter Filter) ([]*User, error) {
    query := r.buildListQuery(filter)
    args := r.buildListArgs(filter)
    
    scanner, err := r.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("list users: %w", err)
    }
    
    var users []*User
    if err := scanner.ScanAll(&users); err != nil {
        return nil, fmt.Errorf("scan users: %w", err)
    }
    
    return users, nil
}

func (r *Repository) buildListQuery(filter Filter) string {
    query := "SELECT id, email, name, role, active, created_at, updated_at FROM users WHERE 1=1"
    
    if filter.Email != nil {
        query += " AND email = $?"
    }
    if filter.Role != nil {
        query += " AND role = $?"
    }
    if filter.Active != nil {
        query += " AND active = $?"
    }
    
    query += " ORDER BY created_at DESC"
    
    if filter.Limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", filter.Limit)
    }
    if filter.Offset > 0 {
        query += fmt.Sprintf(" OFFSET %d", filter.Offset)
    }
    
    return query
}
```

### Service Integration

```go
// internal/user/service.go
package user

import (
    "context"
    "fmt"
)

type Service struct {
    repo *Repository
}

func NewService(repo *Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, email, name string) (*User, error) {
    // Business logic validation
    if email == "" {
        return nil, fmt.Errorf("email is required")
    }
    
    user := &User{
        Email:  email,
        Name:   name,
        Role:   RoleUser,
        Active: true,
    }
    
    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }
    
    return user, nil
}
```

## Migration Management

### Migration Files

Migrations are stored in `migrations/` with numbered prefixes:

```
migrations/
├── 001_initial_schema.up.sql
├── 001_initial_schema.down.sql
├── 002_add_users_table.up.sql
├── 002_add_users_table.down.sql
├── 003_add_indexes.up.sql
└── 003_add_indexes.down.sql
```

### Migration Example

```sql
-- migrations/002_add_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role) WHERE active = true;
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
```

```sql
-- migrations/002_add_users_table.down.sql
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users;
```

### Running Migrations

```bash
# Using migrate CLI
migrate -path migrations -database "postgresql://user:pass@localhost/dbname?sslmode=disable" up

# Or using make command
make migrate-up

# Rollback
make migrate-down

# Create new migration
make migrate-create name=add_products_table
```

## Domain-Specific Query Patterns

### Query Builder in Domain Layer

```go
// internal/user/query_builder.go
package user

import (
    "fmt"
    "strings"
)

// UserQueryBuilder builds user-specific queries
type QueryBuilder struct {
    conditions []string
    args       []interface{}
    orderBy    string
    limit      int
    offset     int
}

func NewQueryBuilder() *QueryBuilder {
    return &QueryBuilder{}
}

func (qb *QueryBuilder) WhereEmail(email string) *QueryBuilder {
    qb.conditions = append(qb.conditions, "email = $?")
    qb.args = append(qb.args, email)
    return qb
}

func (qb *QueryBuilder) WhereRole(role Role) *QueryBuilder {
    qb.conditions = append(qb.conditions, "role = $?")
    qb.args = append(qb.args, role)
    return qb
}

func (qb *QueryBuilder) WhereActive(active bool) *QueryBuilder {
    qb.conditions = append(qb.conditions, "active = $?")
    qb.args = append(qb.args, active)
    return qb
}

func (qb *QueryBuilder) OrderByCreated(desc bool) *QueryBuilder {
    if desc {
        qb.orderBy = "created_at DESC"
    } else {
        qb.orderBy = "created_at ASC"
    }
    return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
    qb.limit = limit
    return qb
}

func (qb *QueryBuilder) Build() (string, []interface{}) {
    query := "SELECT id, email, name, role, active, created_at, updated_at FROM users"
    
    if len(qb.conditions) > 0 {
        query += " WHERE " + strings.Join(qb.conditions, " AND ")
    }
    
    if qb.orderBy != "" {
        query += " ORDER BY " + qb.orderBy
    }
    
    if qb.limit > 0 {
        query += fmt.Sprintf(" LIMIT %d", qb.limit)
    }
    
    if qb.offset > 0 {
        query += fmt.Sprintf(" OFFSET %d", qb.offset)
    }
    
    return query, qb.args
}
```

### Batch Operations in Repository

```go
// internal/user/repository.go - batch operations
func (r *Repository) CreateBatch(ctx context.Context, users []*User) error {
    if len(users) == 0 {
        return nil
    }
    
    // Build batch insert query
    query := "INSERT INTO users (email, name, role, active) VALUES "
    var values []string
    var args []interface{}
    
    for i, user := range users {
        values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d)", 
            i*4+1, i*4+2, i*4+3, i*4+4))
        args = append(args, user.Email, user.Name, user.Role, user.Active)
    }
    
    query += strings.Join(values, ", ")
    query += " RETURNING id, created_at, updated_at"
    
    scanner, err := r.db.Query(query, args...)
    if err != nil {
        return fmt.Errorf("batch create users: %w", err)
    }
    
    // Scan results back into user structs
    var results []struct {
        ID        string    `db:"id"`
        CreatedAt time.Time `db:"created_at"`
        UpdatedAt time.Time `db:"updated_at"`
    }
    
    if err := scanner.ScanAll(&results); err != nil {
        return fmt.Errorf("scan batch results: %w", err)
    }
    
    // Update original user structs with generated IDs and timestamps
    for i, result := range results {
        users[i].ID = result.ID
        users[i].CreatedAt = result.CreatedAt
        users[i].UpdatedAt = result.UpdatedAt
    }
    
    return nil
}
```

## Transaction Management

### Repository Transaction Support

```go
// internal/user/repository.go - transaction support
func (r *Repository) WithTransaction(ctx context.Context, fn func(*Repository) error) error {
    tx, err := r.db.BeginTx(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
    }()
    
    // Create repository with transaction client
    txRepo := &Repository{db: tx}
    
    if err := fn(txRepo); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
        }
        return err
    }
    
    return tx.Commit()
}
```

### Cross-Domain Transaction Example

```go
// internal/order/service.go - using multiple repositories in transaction
func (s *OrderService) CreateOrderWithInventoryCheck(ctx context.Context, userID string, items []OrderItem) (*Order, error) {
    var order *Order
    
    // Use user repository transaction to coordinate across domains
    err := s.userRepo.WithTransaction(ctx, func(txUserRepo *user.Repository) error {
        // Verify user exists
        user, err := txUserRepo.GetByID(ctx, userID)
        if err != nil {
            return fmt.Errorf("get user: %w", err)
        }
        
        if !user.Active {
            return fmt.Errorf("user is not active")
        }
        
        // Create order (this would use the same transaction)
        order = &Order{
            UserID: userID,
            Items:  items,
            Status: OrderStatusPending,
        }
        
        if err := s.orderRepo.Create(ctx, order); err != nil {
            return fmt.Errorf("create order: %w", err)
        }
        
        // Update inventory (this would also use the same transaction)
        for _, item := range items {
            if err := s.inventoryRepo.ReserveItem(ctx, item.ProductID, item.Quantity); err != nil {
                return fmt.Errorf("reserve inventory: %w", err)
            }
        }
        
        return nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("create order transaction: %w", err)
    }
    
    return order, nil
}
```

### Alternative: Service-Level Transaction

```go
// internal/app/transaction.go - application-level transaction management
type TransactionManager struct {
    db storage.Client
}

func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(storage.Transaction) error) error {
    tx, err := tm.db.BeginTx(ctx)
    if err != nil {
        return err
    }
    
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p)
        }
    }()
    
    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback failed: %v, original error: %w", rbErr, err)
        }
        return err
    }
    
    return tx.Commit()
}
```

## Testing Database Code

### Mock Storage Client

```go
// internal/storage/mock/mock.go
package mock

import (
    "github.com/stretchr/testify/mock"
    "github.com/rfizzle/go-starter/internal/storage"
)

type MockClient struct {
    mock.Mock
}

func (m *MockClient) Query(sql string, args ...interface{}) (storage.Scanner, error) {
    mockArgs := m.Called(sql, args)
    if mockArgs.Get(0) == nil {
        return nil, mockArgs.Error(1)
    }
    return mockArgs.Get(0).(storage.Scanner), mockArgs.Error(1)
}

func (m *MockClient) Execute(sql string, args ...interface{}) error {
    mockArgs := m.Called(sql, args)
    return mockArgs.Error(0)
}

type MockScanner struct {
    mock.Mock
}

func (m *MockScanner) ScanOne(dst interface{}) error {
    args := m.Called(dst)
    return args.Error(0)
}

func (m *MockScanner) ScanAll(dst interface{}) error {
    args := m.Called(dst)
    return args.Error(0)
}
```

### Repository Testing

```go
// internal/user/repository_test.go
package user_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/rfizzle/go-starter/internal/storage/mock"
    "github.com/rfizzle/go-starter/internal/user"
)

func TestRepository_Create(t *testing.T) {
    mockClient := &mock.MockClient{}
    mockScanner := &mock.MockScanner{}
    repo := user.NewRepository(mockClient)
    
    testUser := &user.User{
        Email:  "test@example.com",
        Name:   "Test User",
        Role:   user.RoleUser,
        Active: true,
    }
    
    expectedSQL := `INSERT INTO users (email, name, role, active) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
    
    // Mock the query call
    mockClient.On("Query", expectedSQL, testUser.Email, testUser.Name, testUser.Role, testUser.Active).Return(mockScanner, nil)
    
    // Mock the scan call
    mockScanner.On("ScanOne", mock.AnythingOfType("*struct { ID *string; CreatedAt *time.Time; UpdatedAt *time.Time }")).Return(nil).Run(func(args mock.Arguments) {
        // Simulate database returning values
        result := args.Get(0).(*struct {
            ID        *string
            CreatedAt *time.Time
            UpdatedAt *time.Time
        })
        id := "123e4567-e89b-12d3-a456-426614174000"
        *result.ID = id
    })
    
    err := repo.Create(context.Background(), testUser)
    
    assert.NoError(t, err)
    assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", testUser.ID)
    mockClient.AssertExpectations(t)
    mockScanner.AssertExpectations(t)
}
```

### Integration Testing

```go
// internal/user/repository_integration_test.go
package user_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/rfizzle/go-starter/internal/storage/postgres"
    "github.com/rfizzle/go-starter/internal/user"
)

func setupTestDB(t *testing.T) *user.Repository {
    ctx := context.Background()
    
    // Start PostgreSQL container
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    require.NoError(t, err)
    
    t.Cleanup(func() {
        container.Terminate(ctx)
    })
    
    // Get connection string
    dsn, err := container.ConnectionString(ctx)
    require.NoError(t, err)
    
    // Create storage client
    client, err := postgres.New(ctx, dsn)
    require.NoError(t, err)
    
    // Run migrations
    runMigrations(t, dsn)
    
    return user.NewRepository(client)
}

func TestUserRepository_Integration(t *testing.T) {
    repo := setupTestDB(t)
    ctx := context.Background()
    
    // Create user
    user := &user.User{
        Email:  "test@example.com",
        Name:   "Test User",
        Role:   user.RoleUser,
        Active: true,
    }
    
    err := repo.Create(ctx, user)
    require.NoError(t, err)
    assert.NotEmpty(t, user.ID)
    assert.False(t, user.CreatedAt.IsZero())
    
    // Get user by ID
    fetched, err := repo.GetByID(ctx, user.ID)
    require.NoError(t, err)
    assert.Equal(t, user.Email, fetched.Email)
    assert.Equal(t, user.Name, fetched.Name)
    assert.Equal(t, user.Role, fetched.Role)
    
    // List users
    users, err := repo.List(ctx, user.Filter{
        Active: &[]bool{true}[0],
        Limit:  10,
    })
    require.NoError(t, err)
    assert.Len(t, users, 1)
    assert.Equal(t, user.ID, users[0].ID)
}
```

## Performance Optimization

### Connection Pooling

```go
// Configure pool for optimal performance
config := &pgxpool.Config{
    MaxConns:        25,              // Maximum connections
    MinConns:        5,               // Minimum idle connections
    MaxConnLifetime: time.Hour,      // Connection lifetime
    MaxConnIdleTime: 30 * time.Minute, // Idle timeout
    HealthCheckPeriod: time.Minute,  // Health check interval
}
```

### Query Optimization

```go
// Use prepared statements for repeated queries
func (db *DB) PrepareGetUser(ctx context.Context) error {
    _, err := db.pool.Prepare(ctx, "get_user", `
        SELECT id, email, name, role, created_at
        FROM users
        WHERE id = $1
    `)
    return err
}

func (db *DB) GetUserPrepared(ctx context.Context, id string) (*User, error) {
    var user User
    err := db.pool.QueryRow(ctx, "get_user", id).Scan(
        &user.ID, &user.Email, &user.Name, &user.Role, &user.CreatedAt,
    )
    return &user, err
}
```

### Indexing Strategy

```sql
-- Indexes for common queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active_role ON users(role) WHERE active = true;
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at DESC);

-- Composite indexes for complex queries
CREATE INDEX idx_users_search ON users(email, name) 
    WHERE active = true;

-- Partial indexes for filtered queries
CREATE INDEX idx_orders_pending ON orders(status) 
    WHERE status = 'pending';
```

## Best Practices

### 1. Use Context Everywhere
```go
func (db *DB) GetUser(ctx context.Context, id string) (*User, error) {
    // Context enables cancellation and timeout
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return db.queryUser(ctx, id)
}
```

### 2. Handle Errors Properly
```go
if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, storage.ErrNotFound
    }
    if pgErr, ok := err.(*pgconn.PgError); ok {
        if pgErr.Code == "23505" { // Unique violation
            return nil, storage.ErrDuplicate
        }
    }
    return nil, fmt.Errorf("query failed: %w", err)
}
```

### 3. Use Domain-Specific Queries
```go
// Keep queries in domain repositories where they belong
func (r *Repository) ListActiveUsers(ctx context.Context, limit int) ([]*User, error) {
    query := `
        SELECT id, email, name, role, active, created_at, updated_at
        FROM users
        WHERE active = true
        ORDER BY created_at DESC
        LIMIT $1
    `
    
    scanner, err := r.db.Query(query, limit)
    if err != nil {
        return nil, fmt.Errorf("query active users: %w", err)
    }
    
    var users []*User
    if err := scanner.ScanAll(&users); err != nil {
        return nil, fmt.Errorf("scan users: %w", err)
    }
    
    return users, nil
}
```

### 4. Implement Retries in Repositories
```go
// internal/user/repository.go
func (r *Repository) CreateWithRetry(ctx context.Context, user *User) error {
    var err error
    for i := 0; i < 3; i++ {
        if err = r.Create(ctx, user); err == nil {
            return nil
        }
        
        if !r.isRetryableError(err) {
            return err
        }
        
        time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
    }
    return fmt.Errorf("failed after 3 retries: %w", err)
}

func (r *Repository) isRetryableError(err error) bool {
    // Check for connection errors, deadlocks, etc.
    // Domain-specific retry logic
    return strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "deadlock")
}
```

## Monitoring and Debugging

### Query Logging

```go
// Enable query logging in development
config.ConnConfig.Tracer = &pgx.QueryTracer{
    Log: func(ctx context.Context, conn *pgx.Conn, data pgx.QueryTracerLogData) {
        log.Printf("SQL: %s, Args: %v, Time: %s",
            data.SQL, data.Args, data.Time)
    },
}
```

### Health Checks

```go
// internal/storage/postgres/health.go
func (c *Client) HealthCheck(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
    defer cancel()
    
    scanner, err := c.Query("SELECT 1")
    if err != nil {
        return fmt.Errorf("health check query failed: %w", err)
    }
    
    var result int
    if err := scanner.ScanOne(&result); err != nil {
        return fmt.Errorf("health check scan failed: %w", err)
    }
    
    if result != 1 {
        return fmt.Errorf("unexpected health check result: %d", result)
    }
    
    return nil
}
```

## Related Documentation

- [project-context.md](project-context.md) - Project overview
- [interface-pattern.md](../guidelines/interface-pattern.md) - Interface patterns
- [testing-strategy.md](../guidelines/testing-strategy.md) - Testing approaches
- [pgx Documentation](https://github.com/jackc/pgx)
- [PostgreSQL Best Practices](https://wiki.postgresql.org/wiki/Main_Page)

---

*This guide provides comprehensive patterns for database development, ensuring consistent, performant, and testable data layer implementation.*