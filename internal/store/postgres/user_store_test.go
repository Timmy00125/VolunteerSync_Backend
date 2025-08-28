package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/volunteersync/backend/internal/core/user"
)

func setupTestDB(t *testing.T) *sql.DB {
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

	// Run migrations first
	err := MigrateUp(opts)
	if err != nil {
		t.Skipf("Migration failed: %v", err)
	}

	// Open database connection
	db, err := Open(opts)
	if err != nil {
		t.Skipf("Database connection failed: %v", err)
	}

	return db
}

func createTestUser(t *testing.T, db *sql.DB, userID string) {
	// Insert a test user directly into the database
	query := `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`
	now := time.Now().UTC()
	_, err := db.Exec(query, userID, "Test User", "test@example.com", "hashed_password", now, now, true)
	require.NoError(t, err)
}

func TestUserStorePG_GetProfile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-1"

	// Create test user
	createTestUser(t, db, userID)

	t.Run("successful profile retrieval", func(t *testing.T) {
		profile, err := store.GetProfile(ctx, userID)
		
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, userID, profile.ID)
		assert.Equal(t, "Test User", profile.Name)
		assert.Equal(t, "test@example.com", profile.Email)
		assert.True(t, profile.IsVerified)
	})

	t.Run("user not found", func(t *testing.T) {
		profile, err := store.GetProfile(ctx, "nonexistent-user")
		
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserStorePG_UpdateProfile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-2"

	// Create test user
	createTestUser(t, db, userID)

	t.Run("successful profile update", func(t *testing.T) {
		input := user.UpdateProfileInput{
			Name: stringPtr("Updated Name"),
			Bio:  stringPtr("Updated bio"),
		}

		profile, err := store.UpdateProfile(ctx, userID, input)
		
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "Updated Name", profile.Name)
		assert.Equal(t, "Updated bio", *profile.Bio)
	})

	t.Run("update nonexistent user", func(t *testing.T) {
		input := user.UpdateProfileInput{
			Name: stringPtr("Updated Name"),
		}

		profile, err := store.UpdateProfile(ctx, "nonexistent", input)
		
		assert.Error(t, err)
		assert.Nil(t, profile)
	})
}

func TestUserStorePG_SetProfilePicture(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-3"

	// Create test user
	createTestUser(t, db, userID)

	t.Run("successful profile picture update", func(t *testing.T) {
		pictureURL := "https://example.com/profile.jpg"
		
		err := store.SetProfilePicture(ctx, userID, pictureURL)
		
		require.NoError(t, err)

		// Verify the update
		profile, err := store.GetProfile(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, profile.ProfilePictureURL)
		assert.Equal(t, pictureURL, *profile.ProfilePictureURL)
	})

	t.Run("update nonexistent user", func(t *testing.T) {
		err := store.SetProfilePicture(ctx, "nonexistent", "url")
		
		assert.Error(t, err)
	})
}

func TestUserStorePG_InterestManagement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-4"

	// Create test user
	createTestUser(t, db, userID)

	// Create test interests
	interestQuery := `
		INSERT INTO interests (id, name, category)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`
	_, err := db.Exec(interestQuery, "int1", "Environment", "causes")
	require.NoError(t, err)
	_, err = db.Exec(interestQuery, "int2", "Education", "causes")
	require.NoError(t, err)

	t.Run("list all interests", func(t *testing.T) {
		interests, err := store.ListInterests(ctx)
		
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(interests), 2)
		
		// Check that our test interests are present
		foundEnv := false
		foundEdu := false
		for _, interest := range interests {
			if interest.Name == "Environment" {
				foundEnv = true
			}
			if interest.Name == "Education" {
				foundEdu = true
			}
		}
		assert.True(t, foundEnv)
		assert.True(t, foundEdu)
	})

	t.Run("replace user interests", func(t *testing.T) {
		interestIDs := []string{"int1", "int2"}
		
		interests, err := store.ReplaceInterests(ctx, userID, interestIDs)
		
		require.NoError(t, err)
		assert.Len(t, interests, 2)
	})

	t.Run("list user interests", func(t *testing.T) {
		interests, err := store.ListUserInterests(ctx, userID)
		
		require.NoError(t, err)
		assert.Len(t, interests, 2)
	})

	t.Run("replace with empty interests", func(t *testing.T) {
		interests, err := store.ReplaceInterests(ctx, userID, []string{})
		
		require.NoError(t, err)
		assert.Len(t, interests, 0)
	})
}

func TestUserStorePG_SkillManagement(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-5"

	// Create test user
	createTestUser(t, db, userID)

	t.Run("add skill", func(t *testing.T) {
		skillInput := user.SkillInput{
			Name:        "JavaScript",
			Proficiency: "INTERMEDIATE",
		}
		
		skill, err := store.AddSkill(ctx, userID, skillInput)
		
		require.NoError(t, err)
		assert.NotNil(t, skill)
		assert.Equal(t, "JavaScript", skill.Name)
		assert.Equal(t, "INTERMEDIATE", skill.Proficiency)
		assert.NotEmpty(t, skill.ID)
	})

	t.Run("list skills", func(t *testing.T) {
		skills, err := store.ListSkills(ctx, userID)
		
		require.NoError(t, err)
		assert.Len(t, skills, 1)
		assert.Equal(t, "JavaScript", skills[0].Name)
	})

	t.Run("remove skill", func(t *testing.T) {
		// First get the skill ID
		skills, err := store.ListSkills(ctx, userID)
		require.NoError(t, err)
		require.Len(t, skills, 1)
		
		skillID := skills[0].ID
		
		err = store.RemoveSkill(ctx, userID, skillID)
		require.NoError(t, err)
		
		// Verify skill is removed
		skills, err = store.ListSkills(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, skills, 0)
	})

	t.Run("remove nonexistent skill", func(t *testing.T) {
		err := store.RemoveSkill(ctx, userID, "nonexistent-skill")
		
		assert.Error(t, err)
	})
}

func TestUserStorePG_PrivacySettings(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-6"

	// Create test user
	createTestUser(t, db, userID)

	t.Run("update privacy settings", func(t *testing.T) {
		privacy := user.PrivacySettings{
			ProfileVisibility: "VOLUNTEERS_ONLY",
			ShowEmail:         false,
			ShowLocation:      true,
			AllowMessaging:    true,
		}
		
		result, err := store.UpdatePrivacy(ctx, userID, privacy)
		
		require.NoError(t, err)
		assert.Equal(t, "VOLUNTEERS_ONLY", result.ProfileVisibility)
		assert.False(t, result.ShowEmail)
		assert.True(t, result.ShowLocation)
		assert.True(t, result.AllowMessaging)
	})
}

func TestUserStorePG_NotificationPreferences(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-7"

	// Create test user
	createTestUser(t, db, userID)

	t.Run("update notification preferences", func(t *testing.T) {
		prefs := user.NotificationPreferences{
			EmailNotifications:     false,
			PushNotifications:      true,
			SMSNotifications:       false,
			EventReminders:         true,
			NewOpportunities:       true,
			NewsletterSubscription: false,
		}
		
		result, err := store.UpdateNotifications(ctx, userID, prefs)
		
		require.NoError(t, err)
		assert.False(t, result.EmailNotifications)
		assert.True(t, result.PushNotifications)
		assert.False(t, result.SMSNotifications)
		assert.True(t, result.EventReminders)
		assert.True(t, result.NewOpportunities)
		assert.False(t, result.NewsletterSubscription)
	})
}

func TestUserStorePG_ActivityLogs(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewUserStore(db)
	ctx := context.Background()
	userID := "test-user-8"

	// Create test user
	createTestUser(t, db, userID)

	t.Run("log activity", func(t *testing.T) {
		log := user.ActivityLog{
			UserID:  userID,
			Action:  "profile.update",
			Details: map[string]any{"field": "name"},
		}
		
		err := store.LogActivity(ctx, log)
		
		require.NoError(t, err)
	})

	t.Run("list activity logs", func(t *testing.T) {
		logs, err := store.ListActivityLogs(ctx, userID, 10, 0)
		
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 1)
		
		if len(logs) > 0 {
			assert.Equal(t, userID, logs[0].UserID)
			assert.Equal(t, "profile.update", logs[0].Action)
		}
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}