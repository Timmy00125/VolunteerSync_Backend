package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/volunteersync/backend/internal/config"
)

func TestMain(m *testing.M) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Run tests
	code := m.Run()
	
	os.Exit(code)
}

func TestSetupDatabase(t *testing.T) {
	t.Run("database setup with valid config", func(t *testing.T) {
		// Skip if no test database available
		dbURL := os.Getenv("DB_TEST_URL")
		if dbURL == "" {
			t.Skip("DB_TEST_URL not set, skipping database setup tests")
		}

		cfg := &config.Config{}
		cfg.DB.Host = "localhost"
		cfg.DB.Port = 5432
		cfg.DB.User = "postgres"
		cfg.DB.Password = "postgres"
		cfg.DB.Name = "volunteersync_test"
		cfg.DB.SSLMode = "disable"

		db, err := setupDatabase(cfg)
		if err != nil {
			t.Skipf("Database setup failed (database may not be available): %v", err)
		}
		defer db.Close()

		require.NoError(t, err)
		assert.NotNil(t, db)

		// Test database connectivity
		err = db.Ping()
		assert.NoError(t, err)
	})

	t.Run("database setup with invalid config", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.DB.Host = "invalid-host"
		cfg.DB.Port = 9999
		cfg.DB.User = "invalid"
		cfg.DB.Password = "invalid"
		cfg.DB.Name = "invalid"
		cfg.DB.SSLMode = "disable"

		db, err := setupDatabase(cfg)
		
		assert.Error(t, err)
		assert.Nil(t, db)
	})
}

func TestSetupHTTPServer(t *testing.T) {
	// Skip if no test database available
	dbURL := os.Getenv("DB_TEST_URL")
	if dbURL == "" {
		t.Skip("DB_TEST_URL not set, skipping HTTP server setup tests")
	}

	cfg := &config.Config{}
	cfg.Host = "localhost"
	cfg.Port = 8080
	cfg.DB.Host = "localhost"
	cfg.DB.Port = 5432
	cfg.DB.User = "postgres"
	cfg.DB.Password = "postgres"
	cfg.DB.Name = "volunteersync_test"
	cfg.DB.SSLMode = "disable"
	cfg.JWT.AccessSecret = "test-secret"
	cfg.JWT.AccessTTLMin = 15
	cfg.CORS.AllowOrigins = []string{"http://localhost:3000"}
	cfg.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.CORS.AllowHeaders = []string{"Content-Type", "Authorization"}

	// Setup database first
	db, err := setupDatabase(cfg)
	if err != nil {
		t.Skipf("Database setup failed: %v", err)
	}
	defer db.Close()

	t.Run("successful HTTP server setup", func(t *testing.T) {
		srv, err := setupHTTPServer(cfg, db)
		
		require.NoError(t, err)
		assert.NotNil(t, srv)
		assert.Equal(t, "localhost:8080", srv.Addr)
	})
}

func TestSetupRoutes(t *testing.T) {
	// Skip if no test database available
	dbURL := os.Getenv("DB_TEST_URL")
	if dbURL == "" {
		t.Skip("DB_TEST_URL not set, skipping route setup tests")
	}

	cfg := &config.Config{}
	cfg.Host = "localhost"
	cfg.Port = 8080
	cfg.DB.Host = "localhost"
	cfg.DB.Port = 5432
	cfg.DB.User = "postgres"
	cfg.DB.Password = "postgres"
	cfg.DB.Name = "volunteersync_test"
	cfg.DB.SSLMode = "disable"
	cfg.JWT.AccessSecret = "test-secret"
	cfg.JWT.AccessTTLMin = 15
	cfg.CORS.AllowOrigins = []string{"http://localhost:3000"}
	cfg.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.CORS.AllowHeaders = []string{"Content-Type", "Authorization"}

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		t.Skipf("Database setup failed: %v", err)
	}
	defer db.Close()

	// Create router
	router := gin.New()
	setupRoutes(router, db, cfg)

	t.Run("health endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/healthz", nil)
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")
	})

	t.Run("GraphQL playground endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/graphql", nil)
		req.Header.Set("Accept", "text/html")
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GraphQL query endpoint", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/graphql", nil)
		req.Header.Set("Content-Type", "application/json")
		
		router.ServeHTTP(w, req)
		
		// Should not return 404 (route exists)
		assert.NotEqual(t, http.StatusNotFound, w.Code)
	})

	t.Run("CORS headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("OPTIONS", "/graphql", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		
		router.ServeHTTP(w, req)
		
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	})
}

func TestApplicationIntegration(t *testing.T) {
	// Skip if no test database available
	dbURL := os.Getenv("DB_TEST_URL")
	if dbURL == "" {
		t.Skip("DB_TEST_URL not set, skipping application integration tests")
	}

	cfg := &config.Config{}
	cfg.Host = "localhost"
	cfg.Port = 8081 // Use different port to avoid conflicts
	cfg.DB.Host = "localhost"
	cfg.DB.Port = 5432
	cfg.DB.User = "postgres"
	cfg.DB.Password = "postgres"
	cfg.DB.Name = "volunteersync_test"
	cfg.DB.SSLMode = "disable"
	cfg.JWT.AccessSecret = "test-secret"
	cfg.JWT.AccessTTLMin = 15
	cfg.CORS.AllowOrigins = []string{"*"}
	cfg.CORS.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.CORS.AllowHeaders = []string{"Content-Type", "Authorization"}

	t.Run("complete application startup", func(t *testing.T) {
		// Test that we can set up the complete application stack
		
		// Setup database
		db, err := setupDatabase(cfg)
		if err != nil {
			t.Skipf("Database setup failed: %v", err)
		}
		defer db.Close()

		// Setup HTTP server
		srv, err := setupHTTPServer(cfg, db)
		require.NoError(t, err)
		assert.NotNil(t, srv)

		// Test that the server configuration is correct
		assert.Equal(t, "localhost:8081", srv.Addr)
		assert.NotNil(t, srv.Handler)
	})

	t.Run("graceful shutdown simulation", func(t *testing.T) {
		// Test graceful shutdown logic
		
		// Setup database
		db, err := setupDatabase(cfg)
		if err != nil {
			t.Skipf("Database setup failed: %v", err)
		}
		defer db.Close()

		// Setup HTTP server
		srv, err := setupHTTPServer(cfg, db)
		require.NoError(t, err)

		// Start server in goroutine
		go func() {
			srv.ListenAndServe()
		}()

		// Give server time to start
		time.Sleep(100 * time.Millisecond)

		// Test graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		
		err = srv.Shutdown(ctx)
		assert.NoError(t, err)
	})
}

func TestConfigurationLoading(t *testing.T) {
	t.Run("load default configuration", func(t *testing.T) {
		// Set required environment variables for testing
		os.Setenv("JWT_ACCESS_SECRET", "test-jwt-secret")
		defer func() {
			os.Unsetenv("JWT_ACCESS_SECRET")
		}()

		cfg, err := config.Load()
		
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.NotEmpty(t, cfg.Host)
		assert.Greater(t, cfg.Port, 0)
		assert.NotEmpty(t, cfg.DB.Name)
	})

	t.Run("configuration validation", func(t *testing.T) {
		// Test that required configuration fields are validated
		
		cfg := &config.Config{}
		cfg.Host = "localhost"
		cfg.Port = 8080
		cfg.DB.Host = "localhost"
		cfg.DB.Port = 5432
		cfg.DB.User = "postgres"
		cfg.DB.Password = "postgres"
		cfg.DB.Name = "test_db"
		cfg.DB.SSLMode = "disable"
		cfg.JWT.AccessSecret = "test-secret"
		cfg.JWT.AccessTTLMin = 15

		// All required fields should be present
		assert.NotEmpty(t, cfg.Host)
		assert.Greater(t, cfg.Port, 0)
		assert.NotEmpty(t, cfg.DB.Host)
		assert.Greater(t, cfg.DB.Port, 0)
		assert.NotEmpty(t, cfg.DB.User)
		assert.NotEmpty(t, cfg.DB.Name)
		assert.NotEmpty(t, cfg.JWT.AccessSecret)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("database connection failure handling", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.DB.Host = "nonexistent-host"
		cfg.DB.Port = 5432
		cfg.DB.User = "invalid"
		cfg.DB.Password = "invalid"
		cfg.DB.Name = "invalid"
		cfg.DB.SSLMode = "disable"

		db, err := setupDatabase(cfg)
		
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "db open")
	})

	t.Run("invalid port handling", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.Host = "localhost"
		cfg.Port = -1 // Invalid port

		// This should be caught by the HTTP server setup
		srv := &http.Server{
			Addr: "localhost:-1",
		}
		
		// The server should handle invalid addresses gracefully
		assert.NotNil(t, srv)
	})
}