package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/volunteersync/backend/internal/config"
	authcore "github.com/volunteersync/backend/internal/core/auth"
	usercore "github.com/volunteersync/backend/internal/core/user"
	"github.com/volunteersync/backend/internal/graph"
	"github.com/volunteersync/backend/internal/graph/generated"
	mw "github.com/volunteersync/backend/internal/middleware"
	pg "github.com/volunteersync/backend/internal/store/postgres"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Setup database
	db, err := setupDatabase(cfg)
	if err != nil {
		log.Fatalf("database setup: %v", err)
	}
	defer db.Close()

	// Setup HTTP server
	srv, err := setupHTTPServer(cfg, db)
	if err != nil {
		log.Fatalf("server setup: %v", err)
	}

	// Start server and handle graceful shutdown
	startServerWithGracefulShutdown(srv, cfg)
}

// setupDatabase connects to the database and runs migrations
func setupDatabase(cfg *config.Config) (*sql.DB, error) {
	dbOptions := pg.DBOptions{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	}

	// Connect to database
	db, err := pg.Open(dbOptions)
	if err != nil {
		return nil, fmt.Errorf("db open: %w", err)
	}

	// Run migrations
	if err := pg.MigrateUp(dbOptions); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrations: %w", err)
	}

	return db, nil
}

// setupHTTPServer creates and configures the HTTP server
func setupHTTPServer(cfg *config.Config, db *sql.DB) (*http.Server, error) {
	r := gin.Default()

	// Setup CORS
	setupCORS(r, cfg)

	// Setup routes
	setupRoutes(r, db, cfg)

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: r,
	}, nil
}

// setupCORS configures CORS middleware
func setupCORS(r *gin.Engine, cfg *config.Config) {
	corsCfg := cors.Config{
		AllowOrigins: cfg.CORS.AllowOrigins,
		AllowMethods: cfg.CORS.AllowMethods,
		AllowHeaders: cfg.CORS.AllowHeaders,
	}
	corsCfg.AllowCredentials = true
	r.Use(cors.New(corsCfg))
}

// setupRoutes configures all application routes
func setupRoutes(r *gin.Engine, db *sql.DB, cfg *config.Config) {
	// Health endpoint
	r.GET("/healthz", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Static uploads - local file service
	if cfg.Uploads.BaseURL != "" && cfg.Uploads.BaseDir != "" {
		r.Static(cfg.Uploads.BaseURL, cfg.Uploads.BaseDir)
	}

	// GraphQL server
	// Wire user service
	var userSvc *usercore.Service
	{
		// local file service
		maxBytes := int64(cfg.Uploads.MaxMB) * 1024 * 1024
		files := usercore.NewLocalFileService(cfg.Uploads.BaseDir, cfg.Uploads.BaseURL, maxBytes)
		// Postgres user store
		store := pg.NewUserStore(db)
		userSvc = usercore.NewService(store, files, nil, nil)
	}

	// Wire auth service (uses user store for user lookup and refresh token repo from Postgres store)
	var authSvc *authcore.AuthService
	{
		// For demo, reuse user store for user repo via an adapter implemented on UserStorePG
		userRepo := pg.NewAuthUserRepository(db)
		refreshRepo := pg.NewRefreshTokenRepository(db)
		pwd := authcore.NewPasswordService(12)
		jwtSvc, err := authcore.NewJWTService(authcore.JWTConfig{
			AccessSecret:  cfg.JWT.AccessSecret,
			RefreshSecret: cfg.JWT.RefreshSecret,
			AccessExpiry:  time.Duration(cfg.JWT.AccessTTLMin) * time.Minute,
			RefreshExpiry: time.Duration(cfg.JWT.RefreshTTLDays) * 24 * time.Hour,
			Issuer:        "volunteersync",
		})
		if err != nil {
			log.Fatalf("jwt service: %v", err)
		}
		logger := slog.Default()
		authSvc = authcore.NewAuthService(userRepo, refreshRepo, pwd, jwtSvc, logger)
	}

	// Auth middleware
	authMW := mw.NewAuthMiddleware(authSvc, slog.Default())

	gql := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{DB: db, UserService: userSvc}}))
	r.POST("/graphql", authMW.OptionalAuth(), gin.WrapH(gql))
	r.GET("/graphql", authMW.OptionalAuth(), func(c *gin.Context) {
		playground.Handler("GraphQL", "/graphql").ServeHTTP(c.Writer, c.Request)
	})
}

// startServerWithGracefulShutdown starts the server and handles graceful shutdown
func startServerWithGracefulShutdown(srv *http.Server, cfg *config.Config) {
	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()
	log.Printf("server started on http://%s:%d", cfg.Host, cfg.Port)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	log.Println("server exited")
}
