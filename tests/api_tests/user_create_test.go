package api_tests

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/siahsang/blog/tests/test_utils"
)

func TestCreateUser_Success(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	payload := map[string]interface{}{
		"user": map[string]string{
			"email":    "test@example.com",
			"username": "testuser",
			"password": "password123",
		},
	}

	w := client.Post("/api/users", payload)

	test_utils.AssertStatus(t, w, http.StatusOK)

	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)

	user, ok := response["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Response does not contain user object")
	}

	if user["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", user["email"])
	}

	if user["username"] != "testuser" {
		t.Errorf("Expected username 'testuser', got '%v'", user["username"])
	}

	if user["token"] == nil || user["token"] == "" {
		t.Error("Expected token to be present in response")
	}
}

func TestCreateUser_ValidationErrors(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	tests := []struct {
		name       string
		payload    map[string]interface{}
		wantStatus int
	}{
		{
			name: "invalid email format",
			payload: map[string]interface{}{
				"user": map[string]string{
					"email":    "invalid-email",
					"username": "testuser",
					"password": "password123",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			payload: map[string]interface{}{
				"user": map[string]string{
					"email":    "user@example.com",
					"username": "testuser",
					"password": "short",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "username too short",
			payload: map[string]interface{}{
				"user": map[string]string{
					"email":    "user@example.com",
					"username": "usr",
					"password": "password123",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			payload: map[string]interface{}{
				"user": map[string]string{
					"username": "testuser",
					"password": "password123",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing username",
			payload: map[string]interface{}{
				"user": map[string]string{
					"email":    "user@example.com",
					"password": "password123",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: map[string]interface{}{
				"user": map[string]string{
					"email":    "user@example.com",
					"username": "testuser",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty request body",
			payload:    map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := client.Post("/api/users", tt.payload)
			test_utils.AssertStatus(t, w, tt.wantStatus)
		})
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Create first user
	payload1 := map[string]interface{}{
		"user": map[string]string{
			"email":    "duplicate@example.com",
			"username": "user1",
			"password": "password123",
		},
	}
	w1 := client.Post("/api/users", payload1)
	test_utils.AssertStatus(t, w1, http.StatusOK)

	// Try to create another user with same email
	payload2 := map[string]interface{}{
		"user": map[string]string{
			"email":    "duplicate@example.com",
			"username": "user2",
			"password": "password123",
		},
	}
	w2 := client.Post("/api/users", payload2)

	// Should fail with 400 or 500 (constraint violation)
	if w2.Code != http.StatusBadRequest && w2.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 400 or 500 for duplicate email, got %d", w2.Code)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	// Create first user
	payload1 := map[string]interface{}{
		"user": map[string]string{
			"email":    "user1@example.com",
			"username": "duplicateuser",
			"password": "password123",
		},
	}
	w1 := client.Post("/api/users", payload1)
	test_utils.AssertStatus(t, w1, http.StatusOK)

	// Try to create another user with same username
	payload2 := map[string]interface{}{
		"user": map[string]string{
			"email":    "user2@example.com",
			"username": "duplicateuser",
			"password": "password123",
		},
	}
	w2 := client.Post("/api/users", payload2)

	// Should fail with 400 or 500 (constraint violation)
	if w2.Code != http.StatusBadRequest && w2.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 400 or 500 for duplicate username, got %d", w2.Code)
	}
}

func TestCreateUser_DatabaseVerification(t *testing.T) {
	defer test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	db := test_utils.GetTestDB()

	payload := map[string]interface{}{
		"user": map[string]string{
			"email":    "dbtest@example.com",
			"username": "dbtestuser",
			"password": "password123",
		},
	}

	w := client.Post("/api/users", payload)
	test_utils.AssertStatus(t, w, http.StatusOK)

	// Verify user count
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "dbtest@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user in database, got %d", count)
	}

	// Verify username
	var username string
	err = db.QueryRow("SELECT username FROM users WHERE email = $1", "dbtest@example.com").Scan(&username)
	if err != nil {
		t.Fatalf("Failed to get username from database: %v", err)
	}
	if username != "dbtestuser" {
		t.Errorf("Expected username 'dbtestuser', got '%s'", username)
	}

	// Verify password is hashed
	var passwordBytes []byte
	err = db.QueryRow("SELECT password FROM users WHERE email = $1", "dbtest@example.com").Scan(&passwordBytes)
	if err != nil {
		t.Fatalf("Failed to get password from database: %v", err)
	}

	if string(passwordBytes) == "password123" {
		t.Error("Password is stored as plaintext, should be hashed")
	}
	if len(passwordBytes) == 0 {
		t.Error("Password should not be empty")
	}
}

func TestLoginUser_Success(t *testing.T) {
	test_utils.ResetTestDB()
	client := test_utils.NewTestClient(t)

	createTestUser(t, client, "john@smith.com", "john1", "password123")
	//w := client.Post("/api/users/login", map[string]any{
	//	"user": map[string]string{
	//		"email":    "john@smith.com",
	//		"password": "password123",
	//	},
	//})
	//
	//test_utils.AssertStatus(t, w, http.StatusOK)
}

// Helper function to create a test user and return the response
func createTestUser(t *testing.T, client *test_utils.TestClient, email, username, password string) map[string]interface{} {
	t.Helper()


	payload := map[string]interface{}{
		"user": map[string]string{
			"email":    email,
			"username": username,
			"password": password,
		},
	}

	w := client.Post("/api/users", payload)
	test_utils.AssertStatus(t, w, http.StatusOK)

	var response map[string]interface{}
	test_utils.ParseJSON(t, w, &response)
	return response
}

// Helper to verify user exists in database
func verifyUserInDB(t *testing.T, db *sql.DB, email string) {
	t.Helper()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user with email %s, got %d", email, count)
	}
}
