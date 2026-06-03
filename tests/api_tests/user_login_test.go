package api_tests

import (
	"net/http"
	"testing"

	"github.com/siahsang/blog/tests/test_utils"
)

func TestLoginUser_WrongPassword(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	createTestUser(t, client, "alice@example.com", "aliceuser", "correctpassword")

	w := client.Post("/api/users/login", map[string]any{
		"user": map[string]string{
			"email":    "alice@example.com",
			"password": "wrongpassword",
		},
	})

	test_utils.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestLoginUser_NonExistentEmail(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)

	w := client.Post("/api/users/login", map[string]any{
		"user": map[string]string{
			"email":    "nobody@example.com",
			"password": "password123",
		},
	})

	if w.Code == http.StatusOK {
		t.Errorf("expected non-200 status for non-existent user, got %d", w.Code)
	}
}

func TestLoginUser_ValidationErrors(t *testing.T) {
	test_utils.ResetTestDB()

	tests := []struct {
		name       string
		payload    map[string]any
		wantStatus int
	}{
		{
			name: "missing email",
			payload: map[string]any{
				"user": map[string]string{
					"password": "password123",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			payload: map[string]any{
				"user": map[string]string{
					"email": "user@example.com",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			payload: map[string]any{
				"user": map[string]string{
					"email":    "not-an-email",
					"password": "password123",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			payload: map[string]any{
				"user": map[string]string{
					"email":    "user@example.com",
					"password": "short",
				},
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			payload:    map[string]any{},
			wantStatus: http.StatusBadRequest,
		},
	}

	client := test_utils.NewTestClient(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := client.Post("/api/users/login", tt.payload)
			test_utils.AssertStatus(t, w, tt.wantStatus)
		})
	}
}

func TestLoginUser_PasswordExactlyEightChars(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	createTestUser(t, client, "bob@example.com", "bobsmith", "12345678")

	w := client.Post("/api/users/login", map[string]any{
		"user": map[string]string{
			"email":    "bob@example.com",
			"password": "12345678",
		},
	})

	test_utils.AssertStatus(t, w, http.StatusOK)
}

func TestLoginUser_ResponseContainsUserAndToken(t *testing.T) {
	test_utils.ResetTestDB()

	client := test_utils.NewTestClient(t)
	createTestUser(t, client, "carol@example.com", "caroluser", "password123")

	w := client.Post("/api/users/login", map[string]any{
		"user": map[string]string{
			"email":    "carol@example.com",
			"password": "password123",
		},
	})

	test_utils.AssertStatus(t, w, http.StatusOK)

	var response map[string]any
	test_utils.ParseJSON(t, w, &response)

	user, ok := response["user"].(map[string]any)
	if !ok {
		t.Fatal("response does not contain a user object")
	}
	if user["email"] != "carol@example.com" {
		t.Errorf("expected email 'carol@example.com', got '%v'", user["email"])
	}
	if user["username"] != "caroluser" {
		t.Errorf("expected username 'caroluser', got '%v'", user["username"])
	}
	if user["token"] == nil || user["token"] == "" {
		t.Error("expected token to be present in login response")
	}
}
