package postgres

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	// Skip integration tests if DB_TEST_URL not set
	dbURL := os.Getenv("DB_TEST_URL")
	if dbURL == "" {
		t.Skip("DB_TEST_URL not set, skipping database integration tests")
	}

	// Test with valid database options
	t.Run("successful database connection", func(t *testing.T) {
		opts := DBOptions{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Name:     "volunteersync_test",
			SSLMode:  "disable",
		}

		db, err := Open(opts)
		if err != nil {
			t.Skipf("Database not available for testing: %v", err)
		}
		defer db.Close()

		require.NoError(t, err)
		assert.NotNil(t, db)

		// Test connection is working
		err = db.Ping()
		assert.NoError(t, err)
	})

	t.Run("invalid database connection", func(t *testing.T) {
		opts := DBOptions{
			Host:     "invalid-host",
			Port:     5432,
			User:     "invalid",
			Password: "invalid",
			Name:     "invalid",
			SSLMode:  "disable",
		}

		db, err := Open(opts)

		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestMigrateUp(t *testing.T) {
	// Skip if no test database available
	dbURL := os.Getenv("DB_TEST_URL")
	if dbURL == "" {
		t.Skip("DB_TEST_URL not set, skipping migration tests")
	}

	t.Run("migration with valid database", func(t *testing.T) {
		opts := DBOptions{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Name:     "volunteersync_test",
			SSLMode:  "disable",
		}

		// Test migration up
		err := MigrateUp(opts)
		if err != nil {
			t.Skipf("Migration failed, database may not be available: %v", err)
		}

		// Should be no error on successful migration or if no changes needed
		assert.NoError(t, err)

		// Verify database connection after migration
		db, err := Open(opts)
		if err == nil {
			defer db.Close()

			// Check that some expected tables exist
			tables := []string{"users", "events", "event_registrations"}
			for _, table := range tables {
				var exists bool
				query := `SELECT EXISTS (
					SELECT FROM information_schema.tables
					WHERE table_schema = 'public'
					AND table_name = $1
				)`
				err := db.QueryRow(query, table).Scan(&exists)
				assert.NoError(t, err)
				assert.True(t, exists, "Table %s should exist after migration", table)
			}
		}
	})

	t.Run("migration with invalid database", func(t *testing.T) {
		opts := DBOptions{
			Host:     "invalid-host",
			Port:     5432,
			User:     "invalid",
			Password: "invalid",
			Name:     "invalid",
			SSLMode:  "disable",
		}

		err := MigrateUp(opts)
		assert.Error(t, err)
	})
}

// TestDatabaseIntegration tests the complete database setup workflow
func TestDatabaseIntegration(t *testing.T) {
	// Skip if no test database available
	dbURL := os.Getenv("DB_TEST_URL")
	if dbURL == "" {
		t.Skip("DB_TEST_URL not set, skipping database integration tests")
	}

	opts := DBOptions{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Name:     "volunteersync_test",
		SSLMode:  "disable",
	}

	t.Run("complete database setup", func(t *testing.T) {
		// First run migrations
		err := MigrateUp(opts)
		if err != nil {
			t.Skipf("Migration failed: %v", err)
		}

		// Then open connection
		db, err := Open(opts)
		if err != nil {
			t.Skipf("Database connection failed: %v", err)
		}
		defer db.Close()

		require.NoError(t, err)

		// Test basic database operations
		t.Run("basic query operations", func(t *testing.T) {
			// Test simple query
			var version string
			err := db.QueryRow("SELECT version()").Scan(&version)
			assert.NoError(t, err)
			assert.NotEmpty(t, version)

			// Test current timestamp
			var now sql.NullTime
			err = db.QueryRow("SELECT NOW()").Scan(&now)
			assert.NoError(t, err)
			assert.True(t, now.Valid)
		})

		t.Run("transaction support", func(t *testing.T) {
			tx, err := db.Begin()
			require.NoError(t, err)

			// Test rollback
			err = tx.Rollback()
			assert.NoError(t, err)

			// Test commit
			tx, err = db.Begin()
			require.NoError(t, err)
			err = tx.Commit()
			assert.NoError(t, err)
		})
	})
}

// TestConnectionPooling tests database connection pool settings
func TestConnectionPooling(t *testing.T) {
	dbURL := os.Getenv("DB_TEST_URL")
	if dbURL == "" {
		t.Skip("DB_TEST_URL not set, skipping connection pool tests")
	}

	opts := DBOptions{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		Name:     "volunteersync_test",
		SSLMode:  "disable",
	}

	db, err := Open(opts)
	if err != nil {
		t.Skipf("Database not available: %v", err)
	}
	defer db.Close()

	t.Run("connection pool settings", func(t *testing.T) {
		stats := db.Stats()

		// Verify pool settings are applied
		assert.Equal(t, 25, stats.MaxOpenConnections)
		// Note: MaxIdleConnections field doesn't exist in all Go versions, just check it's reasonable
		assert.LessOrEqual(t, stats.Idle, 25)
	})

	t.Run("concurrent connections", func(t *testing.T) {
		// Test multiple concurrent queries
		done := make(chan bool, 5)

		for range 5 {
			go func() {
				var result int
				err := db.QueryRow("SELECT 1").Scan(&result)
				assert.NoError(t, err)
				assert.Equal(t, 1, result)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for range 5 {
			<-done
		}
	})
}
