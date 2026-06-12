## Why

Users can currently follow themselves through the follow API endpoints, which doesn't make semantic sense in a social platform context. This gap allows invalid follower relationships that could confuse users, skew follower counts, and violate the expected behavior of a follow system.

## What Changes

- Add validation in `FollowUser()` core method to prevent users from following themselves
- Add error handling in `Follow()` controller to return 400 Bad Request for self-follow attempts
- Add test coverage for self-follow scenario
- Update existing follow/unfollow spec to document this constraint

## Capabilities

### New Capabilities
<!-- None - this is a validation addition to existing capability -->

### Modified Capabilities
<!-- Existing capabilities whose REQUIREMENTS are changing -->
- `user-follow`: Add requirement that users cannot follow themselves

## Impact

- **Core layer**: `internal/core/profiles.go` - new validation and error type
- **Controller layer**: `internal/controller/user_controller.go` - new error case handling
- **Tests**: `tests/api_tests/follow_test.go` - new test cases
- **API behavior**: POST `/api/profiles/:username/follow` will return 400 when user attempts to follow themselves
