package api_tests

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/siahsang/blog/internal/auth"
	"github.com/siahsang/blog/models"
	"github.com/siahsang/blog/tests/test_utils"
)

type ProfileResponse struct {
	Profile *models.Profile `json:"profile"`
}

type ErrorResponse struct {
	Errors map[string][]string `json:"errors"`
}

func TestGetProfile_Unauthenticated(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	testUser := &auth.User{
		Email:             "testuser@example.com",
		Username:          "testuser",
		PlaintextPassword: "password123",
		Bio:               strPtr("Test bio"),
		Image:             strPtr("https://example.com/image.jpg"),
	}

	createUser(test_utils.GetTestDB(), testUser)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/testuser", nil)
	w := httptest.NewRecorder()
	client.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp ProfileResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp.Profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", resp.Profile.Username)
	}

	if resp.Profile.Bio == nil || *resp.Profile.Bio != "Test bio" {
		t.Errorf("Expected bio 'Test bio', got '%v'", resp.Profile.Bio)
	}

	if resp.Profile.Image == nil || *resp.Profile.Image != "https://example.com/image.jpg" {
		t.Errorf("Expected image URL, got '%v'", resp.Profile.Image)
	}

	if resp.Profile.Following != false {
		t.Errorf("Expected following=false for unauthenticated request, got true")
	}
}

func TestGetProfile_Authenticated_NotFollowing(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	testUser := &auth.User{
		Email:             "testuser@example.com",
		Username:          "testuser",
		PlaintextPassword: "password123",
		Bio:               strPtr("Test bio"),
		Image:             strPtr("https://example.com/image.jpg"),
	}
	createUser(test_utils.GetTestDB(), testUser)

	viewerUser := &auth.User{
		Email:             "viewer@example.com",
		Username:          "viewer",
		PlaintextPassword: "password123",
	}
	createUser(test_utils.GetTestDB(), viewerUser)

	token := generateToken(viewerUser)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/testuser", nil)
	req.Header.Set("Authorization", "Token "+token)
	w := httptest.NewRecorder()
	client.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp ProfileResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp.Profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", resp.Profile.Username)
	}

	if resp.Profile.Following != false {
		t.Errorf("Expected following=false when not following, got true")
	}
}

func TestGetProfile_Authenticated_Following(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	testUser := &auth.User{
		Email:             "testuser@example.com",
		Username:          "testuser",
		PlaintextPassword: "password123",
	}
	createUser(test_utils.GetTestDB(), testUser)

	viewerUser := &auth.User{
		Email:             "viewer@example.com",
		Username:          "viewer",
		PlaintextPassword: "password123",
	}
	createUser(test_utils.GetTestDB(), viewerUser)

	createFollower(test_utils.GetTestDB(), viewerUser.ID, testUser.ID)

	token := generateToken(viewerUser)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/testuser", nil)
	req.Header.Set("Authorization", "Token "+token)
	w := httptest.NewRecorder()
	client.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp ProfileResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp.Profile.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", resp.Profile.Username)
	}

	if resp.Profile.Following != true {
		t.Errorf("Expected following=true when following, got false")
	}
}

func TestGetProfile_NonExistentUser(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/nonexistent", nil)
	w := httptest.NewRecorder()
	client.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
		return
	}

	t.Logf("Response body: %s", w.Body.String())

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, w.Body.String())
	}

	t.Logf("Parsed response: %+v", resp)
	
	errorsField, ok := resp["errors"]
	if !ok {
		t.Fatalf("Expected 'errors' field in response, got: %+v", resp)
	}
	
	errorsMap, ok := errorsField.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected errors to be an object, got: %T", errorsField)
	}
	
	bodyField, ok := errorsMap["body"]
	if !ok {
		t.Fatalf("Expected 'errors.body' field in response, got: %+v", errorsMap)
	}
	
	bodyArray, ok := bodyField.([]interface{})
	if !ok {
		t.Fatalf("Expected errors.body to be an array, got: %T", bodyField)
	}
	
	if len(bodyArray) == 0 {
		t.Errorf("Expected error message in errors.body, got empty array")
	}
}

func TestGetProfile_NullImage(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	testUser := &auth.User{
		Email:             "noimage@example.com",
		Username:          "noimage",
		PlaintextPassword: "password123",
		Bio:               strPtr("Hello world"),
		Image:             nil,
	}
	createUser(test_utils.GetTestDB(), testUser)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/noimage", nil)
	w := httptest.NewRecorder()
	client.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp ProfileResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp.Profile.Image != nil {
		t.Errorf("Expected image to be null, got '%v'", resp.Profile.Image)
	}

	if resp.Profile.Bio == nil || *resp.Profile.Bio != "Hello world" {
		t.Errorf("Expected bio 'Hello world', got '%v'", resp.Profile.Bio)
	}
}

func TestGetProfile_InvalidToken(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	testUser := &auth.User{
		Email:             "testuser@example.com",
		Username:          "testuser",
		PlaintextPassword: "password123",
	}
	createUser(test_utils.GetTestDB(), testUser)

	req := httptest.NewRequest(http.MethodGet, "/api/profiles/testuser", nil)
	req.Header.Set("Authorization", "Token invalid-token")
	w := httptest.NewRecorder()
	client.Engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for invalid token (treated as unauthenticated), got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp ProfileResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if resp.Profile.Following != false {
		t.Errorf("Expected following=false for invalid token, got true")
	}
}

func createUser(db *sql.DB, user *auth.User) {
	if err := user.SetPassword(user.PlaintextPassword); err != nil {
		panic(err)
	}
	
	query := `
		INSERT INTO users (email, username, password, bio, image)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := db.QueryRow(query, user.Email, user.Username, user.Password, user.Bio, user.Image).Scan(&user.ID)
	if err != nil {
		panic(err)
	}
}

func createFollower(db *sql.DB, followerID, userID int64) {
	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
	`
	_, err := db.Exec(query, userID, followerID)
	if err != nil {
		panic(err)
	}
}

func generateToken(user *auth.User) string {
	token, err := user.GenerateToken(time.Hour*24, "test-secret-key")
	if err != nil {
		panic(err)
	}
	return token
}

func strPtr(s string) *string {
	return &s
}
