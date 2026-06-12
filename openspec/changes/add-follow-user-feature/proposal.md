## Why

Users need the ability to follow other users to build a social network and receive updates from authors they're interested in. This is a core feature of the RealWorld blog platform specification that enables content discovery and user engagement.

## What Changes

- Add `POST /api/profiles/:username/follow` endpoint to follow a user
- Add `DELETE /api/profiles/:username/follow` endpoint to unfollow a user
- Implement required authentication middleware for protected routes
- Add new controller handlers for follow/unfollow operations
- Return updated profile with `following` status in response

## Capabilities

### New Capabilities
- `user-follow`: Ability for authenticated users to follow/unfollow other users and view following status

### Modified Capabilities
- None

## Impact

- **New endpoints**: Two new API routes for follow/unfollow operations
- **Authentication**: New required auth middleware (separate from existing optional auth)
- **Controller**: UserController gains Follow() and Unfollow() methods
- **Router**: BlogAPIRouter needs new method to register protected routes
- **Server**: HTTP server setup needs to register protected route group
- **Database**: Existing followers table already supports this feature
