## 1. Core Layer Implementation

- [x] 1.1 Add `CannotFollowSelf` error constant in `internal/core/profiles.go`
- [x] 1.2 Add self-follow validation in `FollowUser()` method
- [x] 1.3 Return appropriate error when user attempts to follow themselves

## 2. Controller Layer Implementation

- [x] 2.1 Add error case handling for `CannotFollowSelf` in `Follow()` controller method
- [x] 2.2 Return 400 Bad Request with "Cannot follow yourself" error message
- [x] 2.3 Optional: Add early validation check before calling core layer

## 3. Test Implementation

- [x] 3.1 Add test case `TestFollow_SelfFollow_400` in `tests/api_tests/follow_test.go`
- [x] 3.2 Verify error response format matches specification
- [x] 3.3 Run all existing tests to ensure no regressions

## 4. Verification

- [x] 4.1 Run `go build ./...` to verify compilation
- [x] 4.2 Run all API tests to confirm passing
- [x] 4.3 Manually test self-follow scenario if needed
