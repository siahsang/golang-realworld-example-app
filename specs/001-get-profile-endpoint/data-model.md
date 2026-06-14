# Data Model: Get Profile Endpoint

**Created**: 2026-06-14  
**Feature**: Get Profile Endpoint  
**Spec**: [spec.md](spec.md)  
**Plan**: [plan.md](plan.md)

## Entities

### Profile (Read-Only View Model)

**Description**: A read-only projection of user information for public display. Not a database table - constructed from the `users` table and optional `followers` relationship.

**Fields**:
- `username` (string, required): User's unique username, exact match from URL parameter
- `bio` (string, optional): User's biography, plain text max 500 characters
- `image` (string or null, optional): Profile image URL, null if not uploaded
- `following` (boolean, required): Whether the current viewer follows this user (always false if unauthenticated)

**Validation Rules**:
- `username`: Must match the requested `:username` parameter exactly (case-sensitive)
- `bio`: Maximum 500 characters, plain text only (no markdown/HTML)
- `image`: Valid URL format or null
- `following`: Boolean, calculated based on authentication status and follower relationship

**Source Mapping**:
```go
// Database mapping
Profile.username ← users.username
Profile.bio      ← users.bio
Profile.image    ← users.image (can be NULL)
Profile.following ← Calculated from followers table OR false
```

---

### User (Existing Database Table)

**Description**: Existing users table that stores user account information. This is the source of profile data.

**Fields** (relevant to this feature):
- `id` (bigint, primary key): Internal user identifier
- `username` (string, unique): Public username, indexed for fast lookup
- `bio` (text, nullable): User biography
- `image` (text, nullable): Profile image URL
- `email` (string, unique): User email (not exposed in profile)
- `password` (string, hashed): User password (not exposed in profile)
- `created_at` (timestamp): Account creation time
- `updated_at` (timestamp): Last profile update time

**Constraints**:
- `username`: UNIQUE, NOT NULL, indexed
- `email`: UNIQUE, NOT NULL
- `password`: NOT NULL (hashed with bcrypt)

---

### Follower Relationship (Existing Database Table)

**Description**: Many-to-many relationship table tracking who follows whom.

**Fields**:
- `follower_id` (bigint, foreign key → users.id): The user who is following
- `followed_id` (bigint, foreign key → users.id): The user being followed
- `created_at` (timestamp): When the follow relationship was created

**Constraints**:
- Primary key: `(follower_id, followed_id)` composite
- Foreign keys: Both reference `users.id`
- Cannot follow self (should be prevented at application level)

**Index**: Recommended index on `(followed_id, follower_id)` for fast following lookups

---

## Relationships

```
┌─────────────┐         ┌──────────────────┐         ┌─────────────┐
│   User      │         │    Follower      │         │   User      │
│ (Profile)   │◄────────│   Relationship   │────────►│ (Viewer)    │
└─────────────┘         └──────────────────┘         └─────────────┘
     ▲                        │                          │
     │                        │                          │
     └────────────────────────┴──────────────────────────┘
                    Query JOIN path
```

**Relationship Rules**:
1. One User has one Profile (1:1 mapping for display purposes)
2. One User can follow many Users (1:N via followers table)
3. One User can be followed by many Users (1:N via followers table)
4. Following relationship is unidirectional (A follows B ≠ B follows A)
5. Profile query may or may not include following status based on authentication

---

## State Transitions

**Profile is read-only** - no state transitions for this feature. The profile is constructed from existing user data on each request.

**Data Flow**:
```
Request: GET /api/profiles/:username
         ↓
Extract: username from URL parameter
         ↓
Query:   SELECT user data + optional following status
         ↓
Construct: Profile object from query result
         ↓
Return:  JSON response with profile data
```

---

## Query Patterns

### Unauthenticated Request

```sql
SELECT 
    u.username,
    u.bio,
    u.image,
    false AS following
FROM users u
WHERE u.username = $1
```

### Authenticated Request

```sql
SELECT 
    u.username,
    u.bio,
    u.image,
    CASE 
        WHEN f.following_id IS NOT NULL THEN true 
        ELSE false 
    END AS following
FROM users u
LEFT JOIN followers f 
    ON f.followed_id = u.id 
    AND f.following_id = $1
WHERE u.username = $2
```

**Parameters**:
- `$1` (authenticated): Current user's ID from JWT token
- `$2`: Requested username from URL parameter

---

## Validation Rules from Requirements

| Requirement | Validation Rule | Implementation |
|-------------|----------------|----------------|
| FR-004 | Username must match :username | WHERE clause exact match |
| FR-005 | Bio max 500 chars, plain text | Database constraint, no rendering |
| FR-006 | Image is URL or null | Store as text, validate on write |
| FR-007 | Following is boolean | CASE expression returns boolean |
| FR-008 | Unauthenticated → following: false | Use simpler query without JOIN |
| FR-009 | Authenticated → actual following status | LEFT JOIN with current user ID |

---

## Notes

- No new database tables or schema changes required
- Existing indexes on `users.username` support fast lookups
- Recommended: Add composite index on `followers(followed_id, follower_id)` if not present
- Profile construction is stateless - no caching strategy defined (can be added later)
- NULL handling: Database NULL → JSON `null` (Go's `sql.NullString` or pointer types)
