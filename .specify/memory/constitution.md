<!--
SYNC IMPACT REPORT
==================
Version change: 1.0.0 → 1.0.1
Modified principles:
  - I. Idiomatic Go: Updated to reference .github/instructions/go-instructions.md as primary authority
Added sections: (none)
Removed sections: (none)
Templates requiring updates: (none)
Follow-up TODOs: (none)
-->

# Blog App Constitution

## Core Principles

### I. Idiomatic Go

All Go code MUST follow `.github/instructions/go-instructions.md`. These instructions are based on Effective Go and the Go Code Review Comments wiki. Code MUST be simple, clear, and favor readability over cleverness. Key requirements:

- **Error Handling**: Check errors immediately, wrap with context using `%w`, handle at appropriate level
- **Naming**: Use mixedCaps, descriptive names, single-letter variables only for short scopes
- **Formatting**: Always use `gofmt`, manage imports with `goimports`
- **Comments**: Write self-documenting code; comments explain why, not what
- **Zero Values**: Make zero values useful; design types accordingly
- **Dependencies**: Prefer standard library; minimize external dependencies
- **Documentation**: Document all exported symbols; English by default

**Rationale**: Idiomatic Go code is maintainable, readable, and aligns with community standards, reducing cognitive load for developers.

### II. API-First (RealWorld Spec)

All API endpoints MUST conform to the RealWorld API specification. The API is the contract; implementation details may change but the API must remain stable. Requirements:

- **RESTful Design**: Use proper HTTP methods (GET, POST, PUT, DELETE) and status codes
- **JSON Format**: All requests and responses use JSON; follow RealWorld schema exactly
- **Authentication**: JWT token-based auth via `Authorization: Token <token>` header
- **Pagination**: Support offset/limit pagination for list endpoints
- **Error Responses**: Consistent error format with descriptive messages
- **Validation**: Validate all input; return 422 for validation errors

**Rationale**: RealWorld API compliance ensures interoperability with existing clients and demonstrates production-ready API design.

### III. Database Integrity

All database operations MUST maintain data integrity through proper migrations, transactions, and constraint enforcement. Requirements:

- **Migrations**: All schema changes require migration files in `migrations/`; use `golang-migrate/migrate`
- **Transactions**: Use transactions for multi-step operations; rollback on errors
- **Constraints**: Define foreign keys, unique constraints, and NOT NULL at database level
- **Connection Management**: Use connection pooling; always close rows and statements
- **SQL Injection**: Use parameterized queries; never concatenate user input into SQL
- **Schema Changes**: Backward-compatible migrations only; plan for zero-downtime deployments

**Rationale**: Database integrity is critical for data correctness; migrations provide audit trail and enable reproducible deployments.

### IV. Authentication & Security

All user-facing endpoints MUST enforce proper authentication and authorization. Security is non-negotiable. Requirements:

- **JWT Tokens**: Use HS256 or stronger; store secrets in environment variables; set reasonable expiration
- **Password Hashing**: Use bcrypt with cost factor >= 10; never store plaintext passwords
- **Input Validation**: Validate all user input on server-side; use `go-playground/validator` for struct validation
- **Authorization**: Verify user ownership before modifying resources; return 403 for unauthorized access
- **Rate Limiting**: Implement rate limiting on authentication endpoints
- **HTTPS**: Require TLS in production; set secure cookie flags

**Rationale**: Security vulnerabilities can compromise user data; defense-in-depth approach minimizes attack surface.

### V. Test Coverage

All critical paths MUST have test coverage. Integration tests are mandatory for API endpoints. Requirements:

- **API Tests**: Integration tests for all endpoints in `tests/api_tests/`
- **Test Structure**: Follow `TEST_STRUCTURE.md`; use table-driven tests where applicable
- **Test Data**: Use test database; truncate tables between tests; never test against production
- **Critical Paths**: Test authentication flows, CRUD operations, error cases
- **Makefile**: Use `make audit` for quality checks before committing
- **CI**: All tests must pass in CI; no disabled tests without documented reason

**Rationale**: Tests provide confidence for refactoring, catch regressions, and document expected behavior.

## Development Workflow

All development MUST follow established workflow to maintain code quality and team coordination:

- **Git Branches**: Feature branches from main; naming convention `<issue>-<feature-name>`
- **Code Review**: All PRs require at least one review; address all comments before merge
- **Commit Messages**: Conventional commits; reference issues; atomic commits
- **Pre-commit**: Run `go fmt`, `go vet`, `staticcheck` before committing
- **Makefile Targets**: Use provided Makefile targets for common operations
- **Environment**: Use `.envrc` for environment variables; document required variables
- **Documentation**: Update README for user-facing changes; document new endpoints

**Quality Gates**:
1. Code formatted with `gofmt`
2. No `go vet` warnings
3. All tests pass (`go test -race ./...`)
4. Static analysis passes (`staticcheck ./...`)
5. PR reviewed and approved

## Security Requirements

All code MUST adhere to security best practices. Security issues take priority over feature development:

- **Secrets Management**: Never commit secrets; use environment variables or secret managers
- **Input Sanitization**: Sanitize HTML content in user-generated content (use `bluemonday`)
- **SQL Injection**: Parameterized queries only; ORM usage must escape properly
- **XSS Prevention**: Set `Content-Type` headers; escape output in templates
- **CORS**: Configure CORS explicitly; avoid wildcard origins in production
- **Logging**: Never log sensitive data (passwords, tokens, PII); use structured logging
- **Dependencies**: Regularly update dependencies; monitor for security advisories

## Governance

This constitution supersedes all other development practices in this repository. Amendments require:

1. **Proposal**: Document proposed change with rationale
2. **Review**: At least one team member review
3. **Documentation**: Update constitution with version bump
4. **Migration**: Plan for migrating existing code if needed

**Versioning Policy**:
- MAJOR: Backward-incompatible principle changes or removals
- MINOR: New principles added or existing principles expanded
- PATCH: Clarifications, wording improvements, typo fixes

**Compliance**:
- All PRs MUST be verified against constitution principles
- Use `Constitution Check` section in implementation plans
- Document any necessary violations in plan's Complexity Tracking table

**Version**: 1.0.1 | **Ratified**: 2026-06-14 | **Last Amended**: 2026-06-14
