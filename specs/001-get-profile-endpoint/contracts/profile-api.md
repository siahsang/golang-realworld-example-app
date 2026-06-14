# API Contract: Get Profile Endpoint

**Created**: 2026-06-14  
**Feature**: Get Profile Endpoint  
**Spec**: [../spec.md](../spec.md)  
**Plan**: [../plan.md](../plan.md)

## Endpoint Specification

### Get User Profile

**Endpoint**: `GET /api/profiles/:username`

**Authentication**: Optional

**Description**: Retrieve a user's public profile information including username, bio, profile image, and following status.

---

## Request

### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | string | Yes | The username of the profile to retrieve (case-sensitive, exact match) |

### Headers

| Header | Type | Required | Description |
|--------|------|----------|-------------|
| `Authorization` | string | No | JWT token in format `Token <jwt_token>` for authentication |
| `Content-Type` | string | No | Should be `application/json` (not used for GET but good practice) |

### Query Parameters

None

### Request Body

None (GET request)

---

## Response

### Success Response (200 OK)

**Status Code**: `200`

**Content-Type**: `application/json`

**Schema**:
```json
{
  "profile": {
    "username": "string",
    "bio": "string or null",
    "image": "string or null",
    "following": "boolean"
  }
}
```

**Example** (with image):
```json
{
  "profile": {
    "username": "jake",
    "bio": "I work at statefarm",
    "image": "https://api.realworld.io/images/smiley-cyrus.jpg",
    "following": false
  }
}
```

**Example** (without image):
```json
{
  "profile": {
    "username": "noimage",
    "bio": "Hello world",
    "image": null,
    "following": false
  }
}
```

**Example** (authenticated, following):
```json
{
  "profile": {
    "username": "jake",
    "bio": "I work at statefarm",
    "image": "https://api.realworld.io/images/smiley-cyrus.jpg",
    "following": true
  }
}
```

---

### Error Responses

#### User Not Found (404)

**Status Code**: `404`

**Content-Type**: `application/json`

**Schema**:
```json
{
  "errors": {
    "body": ["User not found"]
  }
}
```

**Example**:
```json
{
  "errors": {
    "body": ["User not found"]
  }
}
```

---

## Authentication Behavior

### Without Authentication

When no `Authorization` header is provided:
- Request succeeds with status 200
- Response includes `following: false`
- All other profile fields returned normally

### With Valid Authentication

When valid JWT token is provided in `Authorization: Token <token>` header:
- Request succeeds with status 200
- Response includes accurate `following` status (true if viewer follows profile user, false otherwise)
- All profile fields returned normally

### With Invalid Authentication

When invalid/expired JWT token is provided:
- Request succeeds with status 200 (authentication is optional)
- Treated as unauthenticated request
- Response includes `following: false`

**Note**: Unlike protected endpoints, this endpoint does NOT return 401 for invalid tokens because authentication is optional.

---

## Field Specifications

### username

- **Type**: string
- **Required**: Yes
- **Format**: Alphanumeric with underscores/hyphens allowed
- **Constraints**: 
  - Must exactly match the requested `:username` parameter (case-sensitive)
  - Cannot be empty
  - Maximum 255 characters (enforced at database level)

### bio

- **Type**: string or null
- **Required**: No
- **Format**: Plain text (no markdown, HTML, or rich text)
- **Constraints**:
  - Maximum 500 characters
  - Can be empty string or null
  - No sanitization required (stored as-is, displayed as plain text)

### image

- **Type**: string (URL) or null
- **Required**: No
- **Format**: Valid HTTP/HTTPS URL
- **Constraints**:
  - Must be absolute URL (not relative path)
  - Should point to an image resource (jpg, png, gif, webp)
  - Can be null if user hasn't uploaded a profile image
  - No validation of URL accessibility (assumed valid if stored)

### following

- **Type**: boolean
- **Required**: Yes
- **Format**: `true` or `false`
- **Constraints**:
  - Always `false` for unauthenticated requests
  - For authenticated requests: `true` if viewer follows profile user, `false` otherwise
  - Never null or undefined

---

## Edge Cases

### Case-Sensitive Username

**Input**: `GET /api/profiles/Jake` (capital J)  
**Expected**: 404 if only "jake" exists (lowercase)  
**Rationale**: Usernames are case-sensitive for exact matching

### Unicode in Username

**Input**: `GET /api/profiles/用户名`  
**Expected**: 200 if user exists with that username  
**Rationale**: Usernames support Unicode characters

### Very Long Username

**Input**: `GET /api/profiles/<256-character-username>`  
**Expected**: 404 (no user with that username exists)  
**Rationale**: Database constraint limits username length

### Profile with Null Bio

**Response**: `"bio": null` or `"bio": ""`  
**Rationale**: Both acceptable - depends on how bio is stored in database

### User Following Themselves

**Scenario**: User requests their own profile while authenticated  
**Expected**: `following: false` (users don't "follow" themselves)  
**Rationale**: Following relationship is between different users

---

## Testing Checklist

- [ ] GET existing user profile without authentication → 200 with following: false
- [ ] GET existing user profile with valid JWT (not following) → 200 with following: false
- [ ] GET existing user profile with valid JWT (following) → 200 with following: true
- [ ] GET non-existent user → 404 with proper error format
- [ ] GET profile with null image → 200 with image: null
- [ ] GET profile with null bio → 200 with bio: null or ""
- [ ] GET profile with invalid JWT → 200 treated as unauthenticated
- [ ] Case-sensitive username matching (Jake vs jake)
- [ ] Unicode username support
- [ ] Response Content-Type is application/json

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-06-14 | Initial contract specification |

---

## References

- RealWorld API Spec: https://github.com/gothinkster/realworld/tree/main/api
- Feature Spec: [../spec.md](../spec.md)
- Data Model: [../data-model.md](../data-model.md)
