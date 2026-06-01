# Test Structure Documentation

## Overview
This document describes the improved test structure for the blog application, implementing best practices for Go integration testing.

## Directory Structure

```
tests/
├── test_utils/              # Reusable test utilities
│   ├── db_helper.go         # Database setup and utilities
│   ├── app_helper.go        # Application creation helpers
│   └── http_helper.go       # HTTP test client and assertions
├── api_tests/               # API integration tests
│   ├── main_test.go         # TestMain setup (runs once)
│   ├── setup_test.go        # Setup verification tests
│   └── users_create_test.go # User creation endpoint tests
```

## Key Improvements

### 1. HTTP Test Client (`test_utils/http_helper.go`)
Eliminates boilerplate by providing a convenient test client:

```go
client := test_utils.NewTestClient(t)

// Simple POST request
w := client.Post("/api/users", payload)

// Authenticated requests
w := client.GetWithAuth("/api/user", token)
```

**Benefits:**
- Reduces code duplication
- Consistent request handling
- Built-in JSON serialization
- Support for authenticated requests

### 2. Test Assertions
Helper functions for common assertions:

```go
// Assert HTTP status
test_utils.AssertStatus(t, w, http.StatusOK)

// Parse JSON response
var response map[string]interface{}
test_utils.ParseJSON(t, w, &response)

// Assert JSON field value
test_utils.AssertJSONField(t, w, "status", "success")
```

**Benefits:**
- Cleaner test code
- Better error messages
- Automatic failure handling

### 3. Table-Driven Tests
Validation tests use table-driven approach:

```go
tests := []struct {
    name       string
    payload    map[string]interface{}
    wantStatus int
}{
    {
        name: "invalid email",
        payload: /* ... */,
        wantStatus: http.StatusBadRequest,
    },
    // More test cases...
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        w := client.Post("/api/users", tt.payload)
        test_utils.AssertStatus(t, w, tt.wantStatus)
    })
}
```

**Benefits:**
- Easy to add new test cases
- Clear test structure
- Reduced code duplication

### 4. Separated Test Concerns
Tests are organized by purpose:

- **TestCreateUser_Success**: Happy path
- **TestCreateUser_ValidationErrors**: All validation scenarios (table-driven)
- **TestCreateUser_DuplicateEmail/Username**: Business logic
- **TestCreateUser_DatabaseVerification**: Database integration

**Benefits:**
- Easy to navigate
- Clear test intent
- Easier to maintain

### 5. Test Helpers
Helper functions for common operations:

```go
// Create a test user
response := createTestUser(t, client, "user@example.com", "username", "password")

// Verify user in database
verifyUserInDB(t, db, "user@example.com")
```

## Running Tests

```bash
# Run all tests
cd tests/api_tests
go test -v

# Run specific test
go test -v -run TestCreateUser_Success

# Run with race detector
go test -v -race

# Run with coverage
go test -v -cover
```

## Writing New Tests

### Example: Adding a Login Test

```go
func TestUserLogin_Success(t *testing.T) {
    defer test_utils.ResetTestDB()
    
    client := test_utils.NewTestClient(t)
    
    // First create a user
    createTestUser(t, client, "user@example.com", "testuser", "password123")
    
    // Then login
    payload := map[string]interface{}{
        "user": map[string]string{
            "email":    "user@example.com",
            "password": "password123",
        },
    }
    
    w := client.Post("/api/users/login", payload)
    test_utils.AssertStatus(t, w, http.StatusOK)
    
    var response map[string]interface{}
    test_utils.ParseJSON(t, w, &response)
    
    user := response["user"].(map[string]interface{})
    if user["token"] == nil {
        t.Error("Expected token in response")
    }
}
```

## Best Practices

### 1. **Always Clean Up**
```go
func TestSomething(t *testing.T) {
    defer test_utils.ResetTestDB() // Clean up after test
    // ... test code
}
```

### 2. **Use Test Helpers**
```go
// Good
client := test_utils.NewTestClient(t)
w := client.Post("/api/users", payload)

// Avoid
ginEngine := server.NewHttpServer(...)
req := httptest.NewRequest(...)
// ... lots of boilerplate
```

### 3. **Use Table-Driven Tests for Variations**
```go
// Good - for testing multiple similar scenarios
tests := []struct{ name, input string, want int }{...}
for _, tt := range tests { t.Run(tt.name, ...) }

// Avoid - separate test functions for each variation
func TestValidation1(t *testing.T) {...}
func TestValidation2(t *testing.T) {...}
func TestValidation3(t *testing.T) {...}
```

### 4. **Meaningful Test Names**
```go
// Good
func TestCreateUser_DuplicateEmail(t *testing.T)
func TestCreateUser_ValidationErrors(t *testing.T)

// Avoid
func TestCreateUser1(t *testing.T)
func TestUserStuff(t *testing.T)
```

### 5. **Test One Thing Per Test**
```go
// Good - focused test
func TestCreateUser_Success(t *testing.T) {
    // Only tests successful creation
}

// Avoid - testing multiple things
func TestUser(t *testing.T) {
    // Creates user, updates user, deletes user...
}
```

## Comparison: Before vs After

### Before (Original Structure)
```go
func TestCreateUser(t *testing.T) {
    db := test_utils.GetTestDB()
    app, err := test_utils.NewTestApplication(db, nil)
    ginEngine := server.NewHttpServer(true, app.UIRouter, ...)
    
    payload := map[string]interface{}{...}
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    ginEngine.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
    // ... more boilerplate
}
```

### After (Improved Structure)
```go
func TestCreateUser_Success(t *testing.T) {
    defer test_utils.ResetTestDB()
    
    client := test_utils.NewTestClient(t)
    
    payload := map[string]interface{}{
        "user": map[string]string{
            "email":    "test@example.com",
            "username": "testuser",
            "password": "password123",
        },
    }
    
    w := client.Post("/api/users", payload)
    test_utils.AssertStatus(t, w, http.StatusOK)
    
    var response map[string]interface{}
    test_utils.ParseJSON(t, w, &response)
    // ... assertions
}
```

**Lines of code reduced**: ~60% reduction in boilerplate

## Migration Guide

If you have existing tests, migrate them gradually:

1. **Start using http_helper** for new tests
2. **Convert validation tests** to table-driven format
3. **Extract common setup** to helper functions
4. **Refactor old tests** when you need to modify them

## Future Enhancements

As the project grows, consider:

1. **Test fixtures**: Reusable test data
2. **Mock utilities**: For external services
3. **Performance tests**: Load and stress testing
4. **Contract tests**: API contract verification
5. **E2E tests**: Full workflow testing

## Conclusion

This improved test structure provides:
- ✅ Less boilerplate code
- ✅ Easier to write new tests
- ✅ Better test organization
- ✅ Clearer test intent
- ✅ Easier maintenance

The structure scales well as you add more endpoints and test scenarios.
