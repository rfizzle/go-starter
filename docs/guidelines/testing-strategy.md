# Testing Strategy

This document outlines the testing strategy, practices, and standards for Go development in this project.

---

## Testing Philosophy

### Test-Driven Development (TDD)
We follow a strict TDD approach:

1. **Write failing tests first** - Define the expected behavior before implementation
2. **Write minimal code** to pass tests - Implement just enough to make tests green
3. **Refactor** while keeping tests green - Improve code quality without changing behavior
4. **Maintain 80%+ coverage** for new code - Ensure comprehensive test coverage

### Testing Pyramid
- **Unit Tests (70%)** - Fast, isolated tests for individual functions and methods
- **Integration Tests (20%)** - Test component interactions and database operations
- **End-to-End Tests (10%)** - Full system tests through HTTP APIs

---

## Test Organization

### File Structure
```
internal/
├── auth/
│   ├── auth.go
│   ├── auth_test.go          # Unit tests
│   ├── integration_test.go   # Integration tests
│   └── v1/
│       ├── service.go
│       └── service_test.go
```

### Test Naming Conventions
- **File names**: `*_test.go` in the same package
- **Function names**: `TestFunctionName` for unit tests
- **Specific scenarios**: `TestFunctionName_Condition_ExpectedResult`
- **Benchmark tests**: `BenchmarkFunctionName`

### Test Categories
- **Unit tests**: `go test ./...`
- **Integration tests**: `go test ./... -tags=integration`
- **Race condition tests**: `go test ./... -race`

---

## Unit Testing Patterns

### Table-Driven Tests
Preferred pattern for testing multiple scenarios:

```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name    string
        input   CalculateInput
        want    Result
        wantErr bool
    }{
        {
            name: "valid calculation",
            input: CalculateInput{
                Value: 10,
                Operation: "add",
                Operand: 5,
            },
            want: Result{Value: 15},
            wantErr: false,
        },
        {
            name: "invalid operation",
            input: CalculateInput{
                Value: 10,
                Operation: "invalid",
                Operand: 5,
            },
            want: Result{},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Calculate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Calculate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Interface Testing
Test against interfaces to ensure implementation compliance:

```go
func TestRepository(t *testing.T) {
    // Test any implementation of Repository interface
    testRepository := func(t *testing.T, repo Repository) {
        ctx := context.Background()
        
        // Test Create
        user := User{ID: "123", Name: "Test User"}
        err := repo.Create(ctx, user)
        require.NoError(t, err)
        
        // Test GetByID
        retrieved, err := repo.GetByID(ctx, "123")
        require.NoError(t, err)
        assert.Equal(t, user.Name, retrieved.Name)
    }
    
    t.Run("memory implementation", func(t *testing.T) {
        repo := NewMemoryRepository()
        testRepository(t, repo)
    })
    
    t.Run("postgres implementation", func(t *testing.T) {
        repo := NewPostgresRepository(testDB)
        testRepository(t, repo)
    })
}
```

### Test Helpers
Create reusable test utilities:

```go
// testutil/helpers.go
func CreateTestUser(t *testing.T) User {
    t.Helper()
    return User{
        ID:    generateTestID(),
        Name:  "Test User",
        Email: "test@example.com",
    }
}

func AssertUserEqual(t *testing.T, want, got User) {
    t.Helper()
    assert.Equal(t, want.ID, got.ID)
    assert.Equal(t, want.Name, got.Name)
    assert.Equal(t, want.Email, got.Email)
}
```

---

## Integration Testing

### Database Testing
```go
//go:build integration

func TestUserRepository_Integration(t *testing.T) {
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    repo := NewUserRepository(db)
    
    t.Run("create and retrieve user", func(t *testing.T) {
        ctx := context.Background()
        user := User{
            ID:    "test-id",
            Name:  "Test User",
            Email: "test@example.com",
        }
        
        err := repo.Create(ctx, user)
        require.NoError(t, err)
        
        retrieved, err := repo.GetByID(ctx, "test-id")
        require.NoError(t, err)
        assert.Equal(t, user.Name, retrieved.Name)
    })
}

func setupTestDatabase(t *testing.T) *sql.DB {
    t.Helper()
    // Setup test database connection
    // Run migrations
    // Return connection
}

func cleanupTestDatabase(t *testing.T, db *sql.DB) {
    t.Helper()
    // Clean up test data
    // Close connections
}
```

### HTTP API Testing
```go
func TestUserHandler_Integration(t *testing.T) {
    server := setupTestServer(t)
    defer server.Close()
    
    client := server.Client()
    
    t.Run("create user", func(t *testing.T) {
        user := User{Name: "Test User", Email: "test@example.com"}
        body, _ := json.Marshal(user)
        
        resp, err := client.Post(server.URL+"/users", "application/json", bytes.NewReader(body))
        require.NoError(t, err)
        defer resp.Body.Close()
        
        assert.Equal(t, http.StatusCreated, resp.StatusCode)
        
        var created User
        err = json.NewDecoder(resp.Body).Decode(&created)
        require.NoError(t, err)
        assert.Equal(t, user.Name, created.Name)
    })
}
```

---

## Mocking Strategy

### Minimal Mocking
- **Prefer real implementations** when they're fast and reliable
- **Mock external dependencies** (databases, APIs, file systems)
- **Test against interfaces** to enable easy mocking

### Using Testify Mock
```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(User), args.Error(1)
}

func TestService_WithMock(t *testing.T) {
    mockRepo := new(MockRepository)
    service := NewUserService(mockRepo)
    
    user := User{ID: "123", Name: "Test User"}
    mockRepo.On("Create", mock.Anything, user).Return(nil)
    
    err := service.CreateUser(context.Background(), user)
    assert.NoError(t, err)
    
    mockRepo.AssertExpectations(t)
}
```

---

## Test Data Management

### Test Fixtures
```go
// testdata/users.json
[
    {
        "id": "user1",
        "name": "John Doe",
        "email": "john@example.com"
    },
    {
        "id": "user2", 
        "name": "Jane Smith",
        "email": "jane@example.com"
    }
]

// In tests
func loadTestUsers(t *testing.T) []User {
    t.Helper()
    data, err := os.ReadFile("testdata/users.json")
    require.NoError(t, err)
    
    var users []User
    err = json.Unmarshal(data, &users)
    require.NoError(t, err)
    
    return users
}
```

### Factory Functions
```go
func NewTestUser(overrides ...func(*User)) User {
    user := User{
        ID:    "test-" + uuid.New().String(),
        Name:  "Test User",
        Email: "test@example.com",
        CreatedAt: time.Now(),
    }
    
    for _, override := range overrides {
        override(&user)
    }
    
    return user
}

// Usage
user := NewTestUser(func(u *User) {
    u.Name = "Custom Name"
    u.Email = "custom@example.com"
})
```

---

## Testing Tools and Libraries

### Core Testing
- **testing** - Go standard library
- **testify/assert** - Rich assertions
- **testify/require** - Assertions that stop test execution
- **testify/suite** - Test suites with setup/teardown

### Test Database
- **dockertest** - Spin up test databases in Docker
- **testcontainers-go** - Integration testing with containers
- **migrate** - Database migrations for tests

### HTTP Testing
- **httptest** - HTTP testing utilities
- **net/http/httptest** - Test HTTP servers and clients

---

## Performance Testing

### Benchmarks
```go
func BenchmarkCalculate(b *testing.B) {
    input := CalculateInput{Value: 10, Operation: "add", Operand: 5}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, _ = Calculate(input)
    }
}

func BenchmarkCalculateParallel(b *testing.B) {
    input := CalculateInput{Value: 10, Operation: "add", Operand: 5}
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, _ = Calculate(input)
        }
    })
}
```

### Load Testing
```go
func TestService_Load(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }
    
    service := setupService(t)
    
    const numRequests = 1000
    const concurrency = 10
    
    sem := make(chan struct{}, concurrency)
    var wg sync.WaitGroup
    
    start := time.Now()
    
    for i := 0; i < numRequests; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            
            _, err := service.Process(context.Background(), testInput)
            assert.NoError(t, err)
        }()
    }
    
    wg.Wait()
    duration := time.Since(start)
    
    t.Logf("Processed %d requests in %v (%.2f req/sec)", 
        numRequests, duration, float64(numRequests)/duration.Seconds())
}
```

---

## Continuous Integration

### Test Commands
```bash
# Unit tests
go test ./... -v -race -count=1

# Integration tests
go test ./... -v -race -count=1 -tags=integration

# Coverage
go test ./... -race -coverprofile=coverage.out
go tool cover -html=coverage.out

# Benchmarks
go test ./... -bench=. -benchmem
```

### Quality Gates
Tests must pass these criteria:
- [ ] All tests pass
- [ ] Code coverage ≥ 80% for new code
- [ ] No race conditions detected
- [ ] Benchmarks don't regress significantly
- [ ] Integration tests pass

---

## Best Practices

### Test Quality
- **One assertion per test** when possible
- **Descriptive test names** that explain the scenario
- **Arrange-Act-Assert** pattern for clarity
- **Clean up resources** in defer statements or teardown functions

### Test Reliability
- **Deterministic tests** - no random data or timing dependencies
- **Isolated tests** - tests don't depend on each other
- **Fast tests** - unit tests should run in milliseconds
- **Stable tests** - minimize flaky tests through proper synchronization

### Test Maintenance
- **Keep tests simple** - easier to understand and maintain
- **DRY principle** - extract common test setup into helpers
- **Update tests with code** - keep tests in sync with implementation changes
- **Delete obsolete tests** - remove tests for deleted functionality

---

*For specific Go testing patterns and examples, see `docs/guidelines/go-standards.md`. For development workflow, see `docs/development/project-context.md`.*