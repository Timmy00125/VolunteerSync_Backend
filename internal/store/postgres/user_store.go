package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_interests WHERE user_id=$1`, userID); err != nil {
		return nil, err
	}
	if len(interestIDs) > 0 {
		// Build bulk insert VALUES placeholders
		values := make([]string, 0, len(interestIDs))
		args := make([]any, 0, len(interestIDs)+1)
		args = append(args, userID)
		for i, id := range interestIDs {
			values = append(values, fmt.Sprintf("($1,$%d,NOW())", i+2))
			args = append(args, id)
		}
		q := `INSERT INTO user_interests (user_id, interest_id, created_at) VALUES ` + strings.Join(values, ",") + ` ON CONFLICT DO NOTHING`
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	const sel = `SELECT i.id, i.name, c.name
				 FROM user_interests ui
				 JOIN interests i ON i.id=ui.interest_id
				 JOIN interest_categories c ON c.id=i.category_id
				 WHERE ui.user_id=$1
				 ORDER BY c.name, i.name`
	rows, err := s.db.QueryContext(ctx, sel, userID)
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

func (s *UserStorePG) ListUserInterests(ctx context.Context, userID string) ([]user.Interest, error) {
	const q = `SELECT i.id, i.name, c.name FROM user_interests ui JOIN interests i ON i.id=ui.interest_id JOIN interest_categories c ON c.id=i.category_id WHERE ui.user_id=$1 ORDER BY c.name, i.name`
	rows, err := s.db.QueryContext(ctx, q, userID)
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
	const q = `INSERT INTO user_skills (user_id, name, proficiency, verified)
			   VALUES ($1,$2,$3,false)
			   RETURNING id, name, proficiency, verified, created_at, updated_at`
	var sk user.Skill
	if err := s.db.QueryRowContext(ctx, q, userID, in.Name, strings.ToUpper(in.Proficiency)).Scan(
		&sk.ID, &sk.Name, &sk.Proficiency, &sk.Verified, &sk.CreatedAt, &sk.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &sk, nil
}
func (s *UserStorePG) RemoveSkill(ctx context.Context, userID, skillID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM user_skills WHERE id=$1 AND user_id=$2`, skillID, userID)
	return err
}
func (s *UserStorePG) ListSkills(ctx context.Context, userID string) ([]user.Skill, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, proficiency, verified, created_at, updated_at FROM user_skills WHERE user_id=$1 ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []user.Skill
	for rows.Next() {
		var sk user.Skill
		if err := rows.Scan(&sk.ID, &sk.Name, &sk.Proficiency, &sk.Verified, &sk.CreatedAt, &sk.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, sk)
	}
	return out, rows.Err()
}
func (s *UserStorePG) UpdatePrivacy(ctx context.Context, userID string, in user.PrivacySettings) (user.PrivacySettings, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET profile_visibility=$1, show_email=$2, show_location=$3, allow_messaging=$4, updated_at=NOW() WHERE id=$5`,
		strings.ToUpper(in.ProfileVisibility), in.ShowEmail, in.ShowLocation, in.AllowMessaging, userID,
	)
	if err != nil {
		return user.PrivacySettings{}, err
	}
	return in, nil
}
func (s *UserStorePG) UpdateNotifications(ctx context.Context, userID string, in user.NotificationPreferences) (user.NotificationPreferences, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET email_notifications=$1, push_notifications=$2, sms_notifications=$3, event_reminders=$4, new_opportunities=$5, newsletter_subscription=$6, updated_at=NOW() WHERE id=$7`,
		in.EmailNotifications, in.PushNotifications, in.SMSNotifications, in.EventReminders, in.NewOpportunities, in.NewsletterSubscription, userID,
	)
	if err != nil {
		return user.NotificationPreferences{}, err
	}
	return in, nil
}
func (s *UserStorePG) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	return nil, nil
}
func (s *UserStorePG) SetUserRoles(ctx context.Context, userID string, roles []string, assignedBy string) error {
	return nil
}
func (s *UserStorePG) SearchUsers(ctx context.Context, filter user.UserSearchFilter, limit, offset int) ([]user.UserProfile, error) {
	// Minimal baseline: return empty set until full search is implemented
	return []user.UserProfile{}, nil
}
func (s *UserStorePG) LogActivity(ctx context.Context, l user.ActivityLog) error {
	var detailsJSON any
	if l.Details != nil {
		b, err := json.Marshal(l.Details)
		if err != nil {
			return err
		}
		detailsJSON = string(b)
	} else {
		detailsJSON = nil
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO user_activity_logs (user_id, action, details, ip_address, user_agent) VALUES ($1,$2,COALESCE($3::jsonb, NULL),$4,$5)`,
		l.UserID, l.Action, detailsJSON, l.IPAddress, l.UserAgent,
	)
	return err
}
func (s *UserStorePG) ListActivityLogs(ctx context.Context, userID string, limit, offset int) ([]user.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, action, details, ip_address, user_agent, created_at FROM user_activity_logs WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []user.ActivityLog
	for rows.Next() {
		var al user.ActivityLog
		var details sql.NullString
		var ip, ua sql.NullString
		if err := rows.Scan(&al.ID, &al.Action, &details, &ip, &ua, &al.CreatedAt); err != nil {
			return nil, err
		}
		al.UserID = userID
		if details.Valid {
			var m map[string]any
			if err := json.Unmarshal([]byte(details.String), &m); err == nil {
				al.Details = m
			}
		}
		al.IPAddress = nullStringPtr(ip)
		al.UserAgent = nullStringPtr(ua)
		out = append(out, al)
	}
	return out, rows.Err()
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
