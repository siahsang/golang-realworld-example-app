package api_tests

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/siahsang/blog/models"
	"github.com/siahsang/blog/tests/test_utils"
)

type profileResponse struct {
	Profile *models.Profile `json:"profile"`
}

type profileErrorResponse struct {
	Errors map[string][]string `json:"errors"`
}

// TestGetProfile_WithoutAuth_Success tests getting a profile without authentication
// Acceptance Criteria: GET /api/profiles/:username returns profile with following=false
func TestGetProfile_WithoutAuth_Success(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create a test user
	createTestUser(t, client, "test@example.com", "testuser", "password123")

	// Act: Get profile without auth header
	w := client.Get("/api/profiles/testuser")

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Response contains profile object
	var response profileResponse
	test_utils.ParseJSON(t, w, &response)

	if response.Profile == nil {
		t.Fatal("Response does not contain profile object")
	}

	// Assert: Profile fields
	if response.Profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", response.Profile.Username)
	}
	if response.Profile.Following {
		t.Error("Expected following=false for unauthenticated request")
	}
}

// TestGetProfile_WithAuth_Success tests getting a profile with valid authentication
// Acceptance Criteria: GET /api/profiles/:username with valid token returns profile
func TestGetProfile_WithAuth_Success(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create authenticated user
	authResponse := createTestUser(t, client, "auth@example.com", "authuser", "password123")
	token := getUserToken(authResponse)

	// Setup: Create profile to fetch
	createTestUser(t, client, "profile@example.com", "profileuser", "password123")

	// Act: Get profile with auth header
	w := client.GetWithAuth("/api/profiles/profileuser", token)

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Response structure
	var response profileResponse
	test_utils.ParseJSON(t, w, &response)

	if response.Profile == nil {
		t.Fatal("Response does not contain profile object")
	}

	if response.Profile.Username != "profileuser" {
		t.Errorf("Expected username 'profileuser', got '%s'", response.Profile.Username)
	}
	if response.Profile.Following {
		t.Error("Expected following=false when authenticated user is not following profile")
	}
}

// TestGetProfile_WithAuth_FollowingTrue tests following status when user follows profile
// Acceptance Criteria: following=true when authenticated user follows the profile
func TestGetProfile_WithAuth_FollowingTrue(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create follower user
	followerResponse := createTestUser(t, client, "follower@example.com", "follower", "password123")
	followerToken := getUserToken(followerResponse)

	// Setup: Create followed user
	createTestUser(t, client, "followed@example.com", "followed", "password123")

	// Setup: Create following relationship in database
	createFollowing(t, db, "follower", "followed")

	// Act: Get profile with auth
	w := client.GetWithAuth("/api/profiles/followed", followerToken)

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: following=true
	var response profileResponse
	test_utils.ParseJSON(t, w, &response)

	if response.Profile == nil {
		t.Fatal("Response does not contain profile object")
	}

	if !response.Profile.Following {
		t.Error("Expected following=true")
	}
}

// TestGetProfile_NotFound_404 tests error handling for non-existent profile
// Acceptance Criteria: 404 status with error object for non-existent username
func TestGetProfile_NotFound_404(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Act: Get non-existent profile
	w := client.Get("/api/profiles/nonexistent")

	// Assert: Status 404
	test_utils.AssertStatus(t, w, http.StatusNotFound)

	// Assert: Error response format
	var response profileErrorResponse
	test_utils.ParseJSON(t, w, &response)

	bodyErrors, ok := response.Errors["body"]
	if !ok {
		t.Fatal("Response does not contain errors object")
	}
	if len(bodyErrors) == 0 {
		t.Error("Expected error message in body array")
	}
}

// TestGetProfile_InvalidToken_TreatedAsUnauthenticated tests invalid token handling
// Acceptance Criteria: Invalid/expired tokens are silently ignored, following=false
func TestGetProfile_InvalidToken_TreatedAsUnauthenticated(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Setup: Create a test user
	createTestUser(t, client, "test@example.com", "testuser", "password123")

	// Act: Get profile with invalid token
	w := client.GetWithAuth("/api/profiles/testuser", "invalid.token.here")

	// Assert: Status 200 (not 401 - auth is optional)
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Treated as unauthenticated (following=false)
	var response profileResponse
	test_utils.ParseJSON(t, w, &response)

	if response.Profile == nil {
		t.Fatal("Response does not contain profile object")
	}

	if response.Profile.Following {
		t.Error("Expected following=false for invalid token")
	}
}

// TestGetProfile_EmptyUsername_404 tests empty username parameter
// Acceptance Criteria: Empty/missing username returns 404
func TestGetProfile_EmptyUsername_404(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Act: Get profile with empty username (route should not match)
	// This tests the edge case where username is empty or whitespace
	w := client.Get("/api/profiles/")

	// Assert: Should return 404 (route not found or resource not found)
	// Note: Gin may return 404 for route not found
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for empty username, got %d", w.Code)
	}
}

// TestGetProfile_WithBioAndImage tests profile with bio and image fields
// Acceptance Criteria: Bio and image fields are correctly returned
func TestGetProfile_WithBioAndImage(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create user with bio and image using direct DB insert
	_, err := db.Exec(`
		INSERT INTO users (email, username, password, bio, image) 
		VALUES ('test@example.com', 'testuser', 'hashedpw', 'I work at statefarm', 'https://api.realworld.io/images/smiley-cyrus.jpg')
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Act: Get profile
	w := client.Get("/api/profiles/testuser")

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Assert: Bio and image present
	var response profileResponse
	test_utils.ParseJSON(t, w, &response)

	if response.Profile == nil {
		t.Fatal("Response does not contain profile object")
	}

	if response.Profile.Bio == nil || *response.Profile.Bio != "I work at statefarm" {
		t.Errorf("Expected bio 'I work at statefarm', got '%v'", response.Profile.Bio)
	}
	if response.Profile.Image == nil || *response.Profile.Image != "https://api.realworld.io/images/smiley-cyrus.jpg" {
		t.Errorf("Expected correct image URL, got '%v'", response.Profile.Image)
	}
}

// TestGetProfile_NullImage verifies that an absent image is encoded as JSON null.
func TestGetProfile_NullImage(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	_, err := db.Exec(`
		INSERT INTO users (email, username, password, bio, image)
		VALUES ('noimage@example.com', 'noimage', 'hashedpw', 'Hello world', NULL)
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	w := client.Get("/api/profiles/noimage")
	test_utils.AssertStatus(t, w, http.StatusOK)

	var response profileResponse
	test_utils.ParseJSON(t, w, &response)
	if response.Profile == nil {
		t.Fatal("Response does not contain profile object")
	}
	if response.Profile.Image != nil {
		t.Errorf("Expected image to be null, got '%v'", response.Profile.Image)
	}
	if response.Profile.Bio == nil || *response.Profile.Bio != "Hello world" {
		t.Errorf("Expected bio 'Hello world', got '%v'", response.Profile.Bio)
	}
}

// TestGetProfile_MultipleFollowingRelationships tests complex following scenarios
// Acceptance Criteria: Following status is specific to the requesting user
func TestGetProfile_MultipleFollowingRelationships(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create user A (follower)
	userAResponse := createTestUser(t, client, "usera@example.com", "usera", "password123")
	tokenA := getUserToken(userAResponse)

	// Setup: Create user B (not follower)
	userBResponse := createTestUser(t, client, "userb@example.com", "userb", "password123")
	tokenB := getUserToken(userBResponse)

	// Setup: Create profile user
	createTestUser(t, client, "profile@example.com", "profileuser", "password123")

	// Setup: User A follows profile user (user B does not)
	createFollowing(t, db, "usera", "profileuser")

	// Act: User A gets profile (should see following=true)
	wA := client.GetWithAuth("/api/profiles/profileuser", tokenA)
	var responseA profileResponse
	test_utils.ParseJSON(t, wA, &responseA)
	if responseA.Profile == nil {
		t.Fatal("Response for user A does not contain profile object")
	}

	// Act: User B gets profile (should see following=false)
	wB := client.GetWithAuth("/api/profiles/profileuser", tokenB)
	var responseB profileResponse
	test_utils.ParseJSON(t, wB, &responseB)
	if responseB.Profile == nil {
		t.Fatal("Response for user B does not contain profile object")
	}

	// Assert: Different following status for different users
	if !responseA.Profile.Following {
		t.Error("User A: Expected following=true")
	}
	if responseB.Profile.Following {
		t.Error("User B: Expected following=false")
	}
}

// TestGetProfile_SpecialCharactersInUsername tests URL encoding
// Acceptance Criteria: Special characters in username are handled correctly
func TestGetProfile_SpecialCharactersInUsername(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	// Setup: Create user with special characters in username
	_, err := db.Exec(`
		INSERT INTO users (email, username, password) 
		VALUES ('test@example.com', 'test-user_123', 'hashedpw')
	`)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Act: Get profile with URL-encoded username
	w := client.Get("/api/profiles/test-user_123")

	// Assert: Status 200
	test_utils.AssertStatus(t, w, http.StatusOK)

	var response profileResponse
	test_utils.ParseJSON(t, w, &response)

	if response.Profile == nil {
		t.Fatal("Response does not contain profile object")
	}

	if response.Profile.Username != "test-user_123" {
		t.Errorf("Expected username 'test-user_123', got '%s'", response.Profile.Username)
	}
}

func createFollowing(t *testing.T, db *sql.DB, followerUsername, followedUsername string) {
	t.Helper()

	result, err := db.Exec(`
		INSERT INTO followers (follower_id, user_id)
		SELECT follower.id, followed.id
		FROM users AS follower, users AS followed
		WHERE follower.username = $1 AND followed.username = $2
	`, followerUsername, followedUsername)
	if err != nil {
		t.Fatalf("Failed to create following relationship: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("Failed to verify following relationship: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("Expected one following relationship, inserted %d", rowsAffected)
	}
}

// Helper function to extract token from user creation response
func getUserToken(response map[string]interface{}) string {
	user, ok := response["user"].(map[string]interface{})
	if !ok {
		return ""
	}
	token, ok := user["token"].(string)
	if !ok {
		return ""
	}
	return token
}
