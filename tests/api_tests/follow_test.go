package api_tests

import (
	"net/http"
	"testing"

	"github.com/siahsang/blog/tests/test_utils"
)

// TestFollow_Success tests successfully following a user
// Acceptance Criteria: POST /api/profiles/:username/follow returns profile with following=true
func TestFollow_Success(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create authenticated user (follower)
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Setup: Create user to follow
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Act: Follow user with auth header
	w := client.PostWithAuth("/api/profiles/followed/follow", nil, followerToken)

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Response contains profile object
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	profile, ok := response["profile"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain profile object")
	}

	// Assert: Profile fields
	if profile["username"] != "followed" {
		t.Errorf("Expected username 'followed', got '%v'", profile["username"])
	}
	if profile["following"] != true {
		t.Errorf("Expected following=true, got %v", profile["following"])
	}
}

// TestFollow_WithoutAuth_401 tests following without authentication
// Acceptance Criteria: 401 Unauthorized when no token provided
func TestFollow_WithoutAuth_401(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create user to follow
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Act: Follow user without auth header
	w := client.Post("/api/profiles/followed/follow", nil)

	// Assert: Status 401
	test_utils.AssertStatus(t, w, http.StatusUnauthorized)

	// Assert: Error response format
	var response map[string]string
	test_utils.ParseJSON(t, w, &response)

	if response["error"] == "" {
		t.Error("Expected error message in response")
	}
}

// TestFollow_InvalidToken_401 tests following with invalid token
// Acceptance Criteria: 401 Unauthorized when invalid token provided
func TestFollow_InvalidToken_401(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create user to follow
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Act: Follow user with invalid token
	w := client.PostWithAuth("/api/profiles/followed/follow", nil, "invalid.token.here")

	// Assert: Status 401
	test_utils.AssertStatus(t, w, http.StatusUnauthorized)

	// Assert: Error response format
	var response map[string]string
	test_utils.ParseJSON(t, w, &response)

	if response["error"] == "" {
		t.Error("Expected error message in response")
	}
}

// TestFollow_NonExistentUser_404 tests following a user that doesn't exist
// Acceptance Criteria: 404 Not Found when username doesn't exist
func TestFollow_NonExistentUser_404(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Act: Follow non-existent user
	w := client.PostWithAuth("/api/profiles/nonexistent/follow", nil, followerToken)

	// Assert: Status 404
	test_utils.AssertStatus(t, w, http.StatusNotFound)

	// Assert: Error response format
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	errors, ok := response["errors"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain errors object")
	}

	bodyErrors, ok := errors["body"].([]interface{})
	if !ok {
		t.Fatal("Errors does not contain 'body' array")
	}

	if len(bodyErrors) == 0 {
		t.Error("Expected error message in body array")
	}
}

// TestFollow_AlreadyFollowing_400 tests following a user already being followed
// Acceptance Criteria: 400 Bad Request when already following
func TestFollow_AlreadyFollowing_400(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Setup: Create user to follow
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Setup: Create existing following relationship
	_, err := db.Exec(`INSERT INTO followers (user_id, follower_id) VALUES (2, 1)`)
	if err != nil {
		t.Fatalf("Failed to create following relationship: %v", err)
	}

	// Act: Follow user again
	w := client.PostWithAuth("/api/profiles/followed/follow", nil, followerToken)

	// Assert: Status 400
	test_utils.AssertStatus(t, w, http.StatusBadRequest)

	// Assert: Error response format
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	errors, ok := response["errors"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain errors object")
	}

	bodyErrors, ok := errors["body"].([]interface{})
	if !ok {
		t.Fatal("Errors does not contain 'body' array")
	}

	if len(bodyErrors) == 0 {
		t.Error("Expected error message in body array")
	}
}

// TestUnfollow_Success tests successfully unfollowing a user
// Acceptance Criteria: DELETE /api/profiles/:username/follow returns profile with following=false
func TestUnfollow_Success(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Setup: Create user to unfollow
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Setup: Create existing following relationship
	_, err := db.Exec(`INSERT INTO followers (user_id, follower_id) VALUES (2, 1)`)
	if err != nil {
		t.Fatalf("Failed to create following relationship: %v", err)
	}

	// Act: Unfollow user
	w := client.DeleteWithAuth("/api/profiles/followed/follow", followerToken)

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Response contains profile object
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	profile, ok := response["profile"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain profile object")
	}

	// Assert: Profile fields
	if profile["username"] != "followed" {
		t.Errorf("Expected username 'followed', got '%v'", profile["username"])
	}
	if profile["following"] != false {
		t.Errorf("Expected following=false, got %v", profile["following"])
	}
}

// TestUnfollow_NotFollowing_400 tests unfollowing a user not being followed
// Acceptance Criteria: 400 Bad Request when not following
func TestUnfollow_NotFollowing_400(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Setup: Create user to unfollow
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Act: Unfollow user (not following)
	w := client.DeleteWithAuth("/api/profiles/followed/follow", followerToken)

	// Assert: Status 400
	test_utils.AssertStatus(t, w, http.StatusBadRequest)

	// Assert: Error response format
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	errors, ok := response["errors"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain errors object")
	}

	bodyErrors, ok := errors["body"].([]interface{})
	if !ok {
		t.Fatal("Errors does not contain 'body' array")
	}

	if len(bodyErrors) == 0 {
		t.Error("Expected error message in body array")
	}
}

// TestUnfollow_NonExistentUser_404 tests unfollowing a user that doesn't exist
// Acceptance Criteria: 404 Not Found when username doesn't exist
func TestUnfollow_NonExistentUser_404(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Act: Unfollow non-existent user
	w := client.DeleteWithAuth("/api/profiles/nonexistent/follow", followerToken)

	// Assert: Status 404
	test_utils.AssertStatus(t, w, http.StatusNotFound)

	// Assert: Error response format
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	errors, ok := response["errors"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain errors object")
	}

	bodyErrors, ok := errors["body"].([]interface{})
	if !ok {
		t.Fatal("Errors does not contain 'body' array")
	}

	if len(bodyErrors) == 0 {
		t.Error("Expected error message in body array")
	}
}

// TestFollow_ResponseFormat tests that response matches RealWorld spec
// Acceptance Criteria: Response contains profile with username, bio, image, following
func TestFollow_ResponseFormat(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Setup: Create user with bio and image
	_, err := db.Exec(`
		INSERT INTO users (email, username, password, bio, image) 
		VALUES ('followed@example.com', 'followed', 'hashedpw', 'I work at statefarm', 'https://api.realworld.io/images/smiley-cyrus.jpg')
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Act: Follow user
	w := client.PostWithAuth("/api/profiles/followed/follow", nil, followerToken)

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Response structure
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	profile, ok := response["profile"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain profile object")
	}

	// Assert: All required fields present
	requiredFields := []string{"username", "bio", "image", "following"}
	for _, field := range requiredFields {
		if _, exists := profile[field]; !exists {
			t.Errorf("Response missing required field: %s", field)
		}
	}

	// Assert: Field values
	if profile["username"] != "followed" {
		t.Errorf("Expected username 'followed', got '%v'", profile["username"])
	}
	if profile["bio"] != "I work at statefarm" {
		t.Errorf("Expected bio 'I work at statefarm', got '%v'", profile["bio"])
	}
	if profile["image"] != "https://api.realworld.io/images/smiley-cyrus.jpg" {
		t.Errorf("Expected correct image URL, got '%v'", profile["image"])
	}
	if profile["following"] != true {
		t.Errorf("Expected following=true, got %v", profile["following"])
	}
}

// TestUnfollow_ResponseFormat tests unfollow response format
// Acceptance Criteria: Response contains profile with username, bio, image, following=false
func TestUnfollow_ResponseFormat(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Setup: Create user with bio and image
	_, err := db.Exec(`
		INSERT INTO users (email, username, password, bio, image) 
		VALUES ('followed@example.com', 'followed', 'hashedpw', 'I work at statefarm', 'https://api.realworld.io/images/smiley-cyrus.jpg')
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Setup: Create following relationship
	_, err = db.Exec(`INSERT INTO followers (user_id, follower_id) VALUES (2, 1)`)
	if err != nil {
		t.Fatalf("Failed to create following relationship: %v", err)
	}

	// Act: Unfollow user
	w := client.DeleteWithAuth("/api/profiles/followed/follow", followerToken)

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Response structure
	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	profile, ok := response["profile"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain profile object")
	}

	// Assert: All required fields present
	requiredFields := []string{"username", "bio", "image", "following"}
	for _, field := range requiredFields {
		if _, exists := profile[field]; !exists {
			t.Errorf("Response missing required field: %s", field)
		}
	}

	// Assert: following=false
	if profile["following"] != false {
		t.Errorf("Expected following=false, got %v", profile["following"])
	}
}

// TestFollow_EmptyUsername_400 tests following with empty username
// Acceptance Criteria: 400 Bad Request when username is empty
func TestFollow_EmptyUsername_400(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(authResponse)

	// Act: Follow with empty username (route should not match)
	w := client.PostWithAuth("/api/profiles//follow", nil, followerToken)

	// Assert: Should return 400 or 404
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("Expected 400 or 404 for empty username, got %d", w.Code)
	}
}

// TestUnfollow_WithoutAuth_401 tests unfollowing without authentication
// Acceptance Criteria: 401 Unauthorized when no token provided
func TestUnfollow_WithoutAuth_401(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create user to unfollow
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Act: Unfollow user without auth header
	w := client.Delete("/api/profiles/followed/follow")

	// Assert: Status 401
	test_utils.AssertStatus(t, w, http.StatusUnauthorized)

	// Assert: Error response format
	var response map[string]string
	test_utils.ParseJSON(t, w, &response)

	if response["error"] == "" {
		t.Error("Expected error message in response")
	}
}
