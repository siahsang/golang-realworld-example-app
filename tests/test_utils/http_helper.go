package test_utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/siahsang/blog/internal/handler"
	"github.com/siahsang/blog/internal/server"
)

// TestClient provides convenient methods for making HTTP test requests
type TestClient struct {
	Engine *gin.Engine
}

// NewTestClient creates a new test client with the full application stack
func NewTestClient(t *testing.T) *TestClient {
	t.Helper()

	db := GetTestDB()
	logger := CreateTestLogger()
	app, err := NewTestApplication(db, logger)
	if err != nil {
		t.Fatalf("Failed to create test application: %v", err)
	}

	engine := server.NewHttpServer(
		true,
		app.UIRouter,
		app.UIConfig,
		app.BlogAPIRouter,
		app.Logger,
		*app.Config,
		app.Core,
		handler.NewHandler(app.Logger),
	)

	return &TestClient{Engine: engine}
}

// Post makes a POST request to the given path with the given body
func (c *TestClient) Post(path string, body any) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, req)
	return w
}

// Get makes a GET request to the given path
func (c *TestClient) Get(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, req)
	return w
}

// Put makes a PUT request to the given path with the given body
func (c *TestClient) Put(path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, req)
	return w
}

// Delete makes a DELETE request to the given path
func (c *TestClient) Delete(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, req)
	return w
}

// PostWithAuth makes a POST request with an Authorization header
func (c *TestClient) PostWithAuth(path string, body interface{}, token string) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)
	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, req)
	return w
}

// GetWithAuth makes a GET request with an Authorization header
func (c *TestClient) GetWithAuth(path string, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Token "+token)
	w := httptest.NewRecorder()
	c.Engine.ServeHTTP(w, req)
	return w
}

// ParseJSON parses JSON response body into the given interface
func ParseJSON(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, w.Body.String())
	}
}

// AssertStatus checks if the response status matches expected
func AssertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("Expected status %d, got %d. Body: %s", expected, w.Code, w.Body.String())
	}
}

// AssertJSONField checks if a JSON response contains a specific field with expected value
func AssertJSONField(t *testing.T, w *httptest.ResponseRecorder, field string, expected interface{}) {
	t.Helper()
	var response map[string]interface{}
	ParseJSON(t, w, &response)

	if response[field] != expected {
		t.Errorf("Expected %s to be %v, got %v", field, expected, response[field])
	}
}

// ParseJSONFromRecorder parses JSON response body into the given interface
func ParseJSONFromRecorder(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("Failed to parse JSON response: %v. Body: %s", err, w.Body.String())
	}
}
