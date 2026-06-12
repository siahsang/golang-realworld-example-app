---
description: 'General project instructions for all development'
---

# Project Development Guidelines

## Go Development
**CRITICAL**: Before generating any Go code, always consult and follow `.github/instructions/go-instructions.md`.

Key reminders:
- Follow idiomatic Go practices
- Use table-driven tests
- Never duplicate package declarations
- Check errors immediately
- Keep the happy path left-aligned

## Code Generation Checklist
When writing Go code:
1. ✅ Check `.github/instructions/go-instructions.md`
2. ✅ Follow naming conventions
3. ✅ Handle errors properly
4. ✅ Write tests
5. ✅ Document exported symbols

## Testing
- Use the test utilities in `tests/test_utils/`
- Follow the structure in `tests/TEST_STRUCTURE.md`
- Always reset test database with `defer test_utils.ResetTestDB()`
- Do not write unit tests

## Commands
- Build: `go build ./...`
- Test: `go test -v ./...`
- Format: `go fmt ./...`
- Lint: `golangci-lint run`
