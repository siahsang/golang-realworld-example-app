## ADDED Requirements

### Requirement: Follow User Endpoint

The system SHALL allow authenticated users to follow other users by username via POST /api/profiles/:username/follow.

#### Scenario: Successfully follow a user
- **WHEN** authenticated user sends POST request to /api/profiles/:username/follow with valid Authorization header
- **THEN** system returns 200 OK with profile object containing following: true

#### Scenario: Follow without authentication
- **WHEN** user sends POST request without Authorization header
- **THEN** system returns 401 Unauthorized with error message

#### Scenario: Follow with invalid token
- **WHEN** user sends POST request with expired or invalid token
- **THEN** system returns 401 Unauthorized with error message

#### Scenario: Follow non-existent user
- **WHEN** authenticated user tries to follow a username that doesn't exist
- **THEN** system returns 404 Not Found with error message

#### Scenario: Follow already followed user
- **WHEN** authenticated user tries to follow a user they already follow
- **THEN** system returns 400 Bad Request with error message

### Requirement: Unfollow User Endpoint

The system SHALL allow authenticated users to unfollow users they currently follow via DELETE /api/profiles/:username/follow.

#### Scenario: Successfully unfollow a user
- **WHEN** authenticated user sends DELETE request to /api/profiles/:username/follow for a user they follow
- **THEN** system returns 200 OK with profile object containing following: false

#### Scenario: Unfollow without authentication
- **WHEN** user sends DELETE request without Authorization header
- **THEN** system returns 401 Unauthorized with error message

#### Scenario: Unfollow user not being followed
- **WHEN** authenticated user tries to unfollow a user they don't follow
- **THEN** system returns 400 Bad Request with error message

#### Scenario: Unfollow non-existent user
- **WHEN** authenticated user tries to unfollow a username that doesn't exist
- **THEN** system returns 404 Not Found with error message

### Requirement: Profile Response Format

The system SHALL return a profile object with username, bio, image, and following status in follow/unfollow responses.

#### Scenario: Follow response structure
- **WHEN** user successfully follows another user
- **THEN** response contains profile object with username, bio, image, and following fields

#### Scenario: Unfollow response structure
- **WHEN** user successfully unfollows another user
- **THEN** response contains profile object with username, bio, image, and following fields
