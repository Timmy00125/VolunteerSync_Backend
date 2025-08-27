package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/volunteersync/backend/internal/core/user"
)

// UserStorePG is a Postgres implementation of user.UserStore.
type UserStorePG struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStorePG { return &UserStorePG{db: db} }

func (s *UserStorePG) GetProfile(ctx context.Context, userID string) (*user.UserProfile, error) {
	const q = `SELECT id, name, email, bio, profile_picture_url, city, state, country, latitude, longitude,
		profile_visibility, show_email, show_location, allow_messaging,
		email_notifications, push_notifications, sms_notifications,
		event_reminders, new_opportunities, newsletter_subscription,
		created_at, updated_at, last_active_at, is_verified
	  FROM users WHERE id = $1`
	var (
		id, name, email                   string
		bio, pic, city, state, country    sql.NullString
		lat, lng                          sql.NullFloat64
		visibility                        string
		showEmail, showLocation, allowMsg bool
		emailNotif, pushNotif, smsNotif   bool
		eventRem, newOpp, newsSub         bool
		createdAt, updatedAt              time.Time
		lastActive                        sql.NullTime
		isVerified                        bool
	)
	err := s.db.QueryRowContext(ctx, q, userID).Scan(&id, &name, &email, &bio, &pic, &city, &state, &country, &lat, &lng,
		&visibility, &showEmail, &showLocation, &allowMsg,
		&emailNotif, &pushNotif, &smsNotif,
		&eventRem, &newOpp, &newsSub,
		&createdAt, &updatedAt, &lastActive, &isVerified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	prof := &user.UserProfile{
		ID:                id,
		Name:              name,
		Email:             email,
		Bio:               nullStringPtr(bio),
		ProfilePictureURL: nullStringPtr(pic),
		Privacy:           user.PrivacySettings{ProfileVisibility: strings.ToUpper(visibility), ShowEmail: showEmail, ShowLocation: showLocation, AllowMessaging: allowMsg},
		Notifications:     user.NotificationPreferences{EmailNotifications: emailNotif, PushNotifications: pushNotif, SMSNotifications: smsNotif, EventReminders: eventRem, NewOpportunities: newOpp, NewsletterSubscription: newsSub},
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		LastActiveAt:      nullTimePtr(lastActive),
		IsVerified:        isVerified,
	}
	if city.Valid || state.Valid || country.Valid || lat.Valid || lng.Valid {
		prof.Location = &user.Location{City: nullStringPtr(city), State: nullStringPtr(state), Country: nullStringPtr(country), Lat: nullFloatPtr(lat), Lng: nullFloatPtr(lng)}
	}
	// Interests/Skills can be loaded later on demand
	return prof, nil
}

func (s *UserStorePG) UpdateProfile(ctx context.Context, userID string, input user.UpdateProfileInput) (*user.UserProfile, error) {
	var sets []string
	var args []any
	i := 1
	if input.Name != nil {
		sets = append(sets, fmt.Sprintf("name=$%d", i))
		args = append(args, *input.Name)
		i++
	}
	if input.Bio != nil {
		sets = append(sets, fmt.Sprintf("bio=$%d", i))
		args = append(args, *input.Bio)
		i++
	}
	if input.Location != nil {
		loc := input.Location
		if loc.City != nil {
			sets = append(sets, fmt.Sprintf("city=$%d", i))
			args = append(args, *loc.City)
			i++
		}
		if loc.State != nil {
			sets = append(sets, fmt.Sprintf("state=$%d", i))
			args = append(args, *loc.State)
			i++
		}
		if loc.Country != nil {
			sets = append(sets, fmt.Sprintf("country=$%d", i))
			args = append(args, *loc.Country)
			i++
		}
		if loc.Lat != nil {
			sets = append(sets, fmt.Sprintf("latitude=$%d", i))
			args = append(args, *loc.Lat)
			i++
		}
		if loc.Lng != nil {
			sets = append(sets, fmt.Sprintf("longitude=$%d", i))
			args = append(args, *loc.Lng)
			i++
		}
	}
	if len(sets) == 0 {
		return s.GetProfile(ctx, userID)
	}
	args = append(args, userID)
	q := "UPDATE users SET " + strings.Join(sets, ", ") + ", updated_at=NOW() WHERE id=$" + fmt.Sprint(i)
	if _, err := s.db.ExecContext(ctx, q, args...); err != nil {
		return nil, err
	}
	return s.GetProfile(ctx, userID)
}

func (s *UserStorePG) SetProfilePicture(ctx context.Context, userID, url string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET profile_picture_url=$1, updated_at=NOW() WHERE id=$2`, url, userID)
	return err
}

func (s *UserStorePG) ReplaceInterests(ctx context.Context, userID string, interestIDs []string) ([]user.Interest, error) {
	return nil, nil
}

func (s *UserStorePG) ListInterests(ctx context.Context) ([]user.Interest, error) {
	const q = `SELECT i.id, i.name, c.name FROM interests i JOIN interest_categories c ON c.id=i.category_id ORDER BY c.name, i.name`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []user.Interest
	for rows.Next() {
		var it user.Interest
		if err := rows.Scan(&it.ID, &it.Name, &it.Category); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *UserStorePG) AddSkill(ctx context.Context, userID string, in user.SkillInput) (*user.Skill, error) {
	return nil, nil
}
func (s *UserStorePG) RemoveSkill(ctx context.Context, userID, skillID string) error { return nil }
func (s *UserStorePG) ListSkills(ctx context.Context, userID string) ([]user.Skill, error) {
	return nil, nil
}
func (s *UserStorePG) UpdatePrivacy(ctx context.Context, userID string, in user.PrivacySettings) (user.PrivacySettings, error) {
	return user.PrivacySettings{}, nil
}
func (s *UserStorePG) UpdateNotifications(ctx context.Context, userID string, in user.NotificationPreferences) (user.NotificationPreferences, error) {
	return user.NotificationPreferences{}, nil
}
func (s *UserStorePG) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}
func (s *UserStorePG) SetUserRoles(ctx context.Context, userID string, roles []string, assignedBy string) error {
	return nil
}
func (s *UserStorePG) SearchUsers(ctx context.Context, filter user.UserSearchFilter, limit, offset int) ([]user.UserProfile, error) {
	return nil, nil
}
func (s *UserStorePG) LogActivity(ctx context.Context, log user.ActivityLog) error { return nil }
func (s *UserStorePG) ListActivityLogs(ctx context.Context, userID string, limit, offset int) ([]user.ActivityLog, error) {
	return nil, nil
}

func nullStringPtr(ns sql.NullString) *string {
	if ns.Valid {
		v := ns.String
		return &v
	}
	return nil
}
func nullFloatPtr(nf sql.NullFloat64) *float64 {
	if nf.Valid {
		v := nf.Float64
		return &v
	}
	return nil
}
func nullTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		v := nt.Time
		return &v
	}
	return nil
}
