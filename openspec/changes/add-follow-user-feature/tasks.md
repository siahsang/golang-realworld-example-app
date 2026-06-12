## 1. Schema and Middleware Setup

- [x] 1.1 Add FollowUserResponse schema to internal/schema/user_schema.go
- [x] 1.2 Implement RequiredAuthMiddleware() in internal/middleware/auth.go

## 2. Controller Implementation

- [x] 2.1 Add Follow() handler method to UserController
- [x] 2.2 Add Unfollow() handler method to UserController
- [x] 2.3 Implement error handling for 401, 400, 404 responses in handlers

## 3. Router Configuration

- [x] 3.1 Add RegisterAuthAPIRouter() method to BlogAPIRouter
- [x] 3.2 Register POST /profiles/:username/follow route
- [x] 3.3 Register DELETE /profiles/:username/follow route

## 4. Server Setup

- [x] 4.1 Add protected route group registration in internal/server/http.go
- [x] 4.2 Apply RequiredAuthMiddleware to protected routes

## 5. Testing and Verification

- [x] 5.1 Test follow endpoint with valid authentication - TestFollow_Success
- [x] 5.2 Test follow endpoint without authentication (401) - TestFollow_WithoutAuth_401
- [x] 5.3 Test follow endpoint with invalid token (401) - TestFollow_InvalidToken_401
- [x] 5.4 Test follow non-existent user (404) - TestFollow_NonExistentUser_404
- [x] 5.5 Test follow already followed user (400) - TestFollow_AlreadyFollowing_400
- [x] 5.6 Test unfollow endpoint successfully - TestUnfollow_Success
- [x] 5.7 Test unfollow user not being followed (400) - TestUnfollow_NotFollowing_400
- [x] 5.8 Verify response format matches specification - TestFollow_ResponseFormat, TestUnfollow_ResponseFormat
