package api_tests

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/siahsang/blog/internal/utils/databaseutils"
	"github.com/siahsang/blog/tests/test_utils"
)

func TestMain(m *testing.M) {
	logger := slog.Default()
	
	// Set test database DSN
	testDBDSN := "postgres://postgres:postgres@localhost/myblog_test?sslmode=disable"
	os.Setenv("DB_DSN", testDBDSN)
	os.Setenv("JWT_SECRET", "test-secret-key")
	
	// Create test database
	logger.Info("Creating test database...")
	test_utils.CreateTestDB()
	
	// Connect to test database
	logger.Info("Connecting to test database...")
	db, err := databaseutils.OpenDBConnection(logger, testDBDSN)
	if err != nil {
		logger.Error("Failed to connect to test database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	
	// Run migrations
	logger.Info("Running migrations on test database...")
	projectRoot := "/home/javad/Project/go-sample/blog-app/my-blog-app"
	migrationsPath := filepath.Join(projectRoot, "migrations")
	if err := test_utils.RunMigrations(db, migrationsPath); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}
	
	// Setup test database connection for test utilities
	test_utils.SetupTestDB()
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	test_utils.TearDownTestDB()
	
	os.Exit(code)
}
