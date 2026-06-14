# Quickstart Validation: Get Profile Endpoint

**Created**: 2026-06-14  
**Feature**: Get Profile Endpoint  
**Spec**: [spec.md](spec.md)  
**Plan**: [plan.md](plan.md)

## Purpose

This guide provides manual validation steps to verify the Get Profile endpoint implementation works correctly before running automated tests.

---

## Prerequisites

1. **Database**: PostgreSQL running with migrations applied
2. **Application**: Server running on `http://localhost:8080` (or configured port)
3. **Test Data**: At least one user in the database
4. **Tools**: `curl` or similar HTTP client (Postman, HTTPie, etc.)

---

## Setup Commands

### 1. Start the Application

```bash
# From project root
make run/api
# Or manually:
go run ./cmd/main.go
```

Wait for log message: "Starting application..."

### 2. Create Test User

```bash
# Register a new user (adjust port if needed)
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "username": "testuser",
      "email": "test@example.com",
      "password": "password123"
    }
  }'
```

**Expected**: 201 Created with user object including JWT token

### 3. (Optional) Update Profile with Bio and Image

```bash
# Get token from previous response and use it
TOKEN="<your-jwt-token-from-registration>"

curl -X PUT http://localhost:8080/api/user \
  -H "Content-Type: application/json" \
  -H "Authorization: Token $TOKEN" \
  -d '{
    "user": {
      "bio": "I work at statefarm",
      "image": "https://api.realworld.io/images/smiley-cyrus.jpg"
    }
  }'
```

**Expected**: 200 OK with updated user object

---

## Validation Scenarios

### Scenario 1: Get Profile Without Authentication

**Command**:
```bash
curl -v http://localhost:8080/api/profiles/testuser
```

**Expected Output**:
```json
{
  "profile": {
    "username": "testuser",
    "bio": "I work at statefarm",
    "image": "https://api.realworld.io/images/smiley-cyrus.jpg",
    "following": false
  }
}
```

**Validation Checklist**:
- [ ] Status code: 200 OK
- [ ] Username matches requested username
- [ ] Bio matches what was set
- [ ] Image URL is present (or null if not set)
- [ ] `following` is `false`
- [ ] Response Content-Type: `application/json`

---

### Scenario 2: Get Profile with Authentication (Not Following)

**Command**:
```bash
curl -v http://localhost:8080/api/profiles/testuser \
  -H "Authorization: Token $TOKEN"
```

**Expected Output**: Same as Scenario 1, but `following` reflects actual status

**Validation Checklist**:
- [ ] Status code: 200 OK
- [ ] `following` is `false` (assuming you're not following testuser)
- [ ] All other fields same as Scenario 1

---

### Scenario 3: Get Non-Existent User Profile

**Command**:
```bash
curl -v http://localhost:8080/api/profiles/unknownuser
```

**Expected Output**:
```json
{
  "errors": {
    "body": ["User not found"]
  }
}
```

**Validation Checklist**:
- [ ] Status code: 404 Not Found
- [ ] Error format matches RealWorld spec
- [ ] Error message is "User not found"

---

### Scenario 4: Get Profile with Null Image

**Setup**: Create a user without setting an image

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "username": "noimage",
      "email": "noimage@example.com",
      "password": "password123"
    }
  }'
```

**Command**:
```bash
curl -v http://localhost:8080/api/profiles/noimage
```

**Expected Output**:
```json
{
  "profile": {
    "username": "noimage",
    "bio": null,
    "image": null,
    "following": false
  }
}
```

**Validation Checklist**:
- [ ] Status code: 200 OK
- [ ] `image` is `null` (not empty string)
- [ ] `bio` is `null` or empty string

---

### Scenario 5: Case-Sensitive Username

**Command**:
```bash
curl -v http://localhost:8080/api/profiles/TestUser
```

**Expected**: 404 (if only "testuser" exists with lowercase)

**Validation Checklist**:
- [ ] Status code: 404 Not Found
- [ ] Usernames are case-sensitive

---

## Manual Testing with JWT from Different User

### Setup: Create Second User and Follow First User

```bash
# Login as second user
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "user": {
      "email": "follower@example.com",
      "password": "password123"
    }
  }'

# Get token and follow testuser
FOLLOWER_TOKEN="<token-from-login>"

curl -X POST http://localhost:8080/api/profiles/testuser/follow \
  -H "Authorization: Token $FOLLOWER_TOKEN"
```

### Validate Following Status

```bash
curl -v http://localhost:8080/api/profiles/testuser \
  -H "Authorization: Token $FOLLOWER_TOKEN"
```

**Expected**:
```json
{
  "profile": {
    "username": "testuser",
    "bio": "I work at statefarm",
    "image": "https://api.realworld.io/images/smiley-cyrus.jpg",
    "following": true
  }
}
```

**Validation Checklist**:
- [ ] Status code: 200 OK
- [ ] `following` is `true`

---

## Common Issues and Troubleshooting

### Issue: 401 Unauthorized

**Cause**: Endpoint incorrectly requires authentication  
**Fix**: Remove authentication requirement from handler

### Issue: 500 Internal Server Error

**Cause**: Database connection issue or query error  
**Fix**: Check database is running, verify query syntax

### Issue: following Always Returns false

**Cause**: JWT not being parsed correctly or follower query wrong  
**Fix**: Verify token extraction and following status query logic

### Issue: Null Pointer Exception

**Cause**: Not handling null bio/image from database  
**Fix**: Use `sql.NullString` or pointer types in Go struct

---

## Next Steps After Manual Validation

1. **Run Automated Tests**:
   ```bash
   go test -v ./tests/api_tests/profile_test.go
   ```

2. **Check Code Quality**:
   ```bash
   make audit
   ```

3. **Review Implementation**:
   - Verify handler follows Go best practices
   - Check error handling and logging
   - Ensure parameterized queries (no SQL injection)

---

## References

- API Contract: [contracts/profile-api.md](contracts/profile-api.md)
- Data Model: [data-model.md](data-model.md)
- Feature Spec: [spec.md](spec.md)
- RealWorld API Spec: https://github.com/gothinkster/realworld/tree/main/api
