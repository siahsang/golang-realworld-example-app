package api_tests

import (
	"testing"

	"github.com/siahsang/blog/tests/test_utils"
)

func TestDatabaseConnection(t *testing.T) {
	// Get test database connection
	db := test_utils.GetTestDB()
	if db == nil {
		t.Fatal("Test database is not initialized")
	}

	// Test ping
	err := db.Ping()
	if err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	t.Log("Test database connection is working")
}

func TestApplicationCreation(t *testing.T) {
	// Get test database connection
	db := test_utils.GetTestDB()
	if db == nil {
		t.Fatal("Test database is not initialized")
	}

	// Create test application instance
	app, err := test_utils.NewTestApplication(db, nil)
	if err != nil {
		t.Fatalf("Failed to create test application: %v", err)
	}

	if app == nil {
		t.Fatal("Test application is nil")
	}

	if app.DB == nil {
		t.Fatal("Application database is nil")
	}

	t.Log("Test application created successfully")
}
