package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	auth "github.com/volunteersync/backend/internal/core/auth"
)

// AuthUserRepository implements auth.UserRepository using Postgres
type AuthUserRepository struct {
	db *sql.DB
}

func NewAuthUserRepository(db *sql.DB) *AuthUserRepository { return &AuthUserRepository{db: db} }

// CreateUser creates a new user record
func (r *AuthUserRepository) CreateUser(ctx context.Context, user *auth.User) error {
	const q = `INSERT INTO users (id, email, name, password_hash, email_verified, google_id, last_login, failed_login_attempts, locked_until, created_at, updated_at)
               VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.db.ExecContext(ctx, q,
		user.ID, user.Email, user.Name, user.PasswordHash, user.EmailVerified, user.GoogleID, user.LastLogin,
		user.FailedLoginAttempts, user.LockedUntil, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

// GetUserByID fetches a user by ID
func (r *AuthUserRepository) GetUserByID(ctx context.Context, id string) (*auth.User, error) {
	const q = `SELECT id, email, name, password_hash, email_verified, google_id, last_login, failed_login_attempts, locked_until, created_at, updated_at
               FROM users WHERE id=$1`
	var u auth.User
	var pwd sql.NullString
	var gid sql.NullString
	var last sql.NullTime
	var locked sql.NullTime
	if err := r.db.QueryRowContext(ctx, q, id).Scan(&u.ID, &u.Email, &u.Name, &pwd, &u.EmailVerified, &gid, &last, &u.FailedLoginAttempts, &locked, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	if pwd.Valid {
		p := pwd.String
		u.PasswordHash = &p
	}
	if gid.Valid {
		g := gid.String
		u.GoogleID = &g
	}
	if last.Valid {
		t := last.Time
		u.LastLogin = &t
	}
	if locked.Valid {
		t := locked.Time
		u.LockedUntil = &t
	}
	return &u, nil
}

// GetUserByEmail fetches a user by email
func (r *AuthUserRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	const q = `SELECT id, email, name, password_hash, email_verified, google_id, last_login, failed_login_attempts, locked_until, created_at, updated_at
               FROM users WHERE LOWER(email)=LOWER($1)`
	var u auth.User
	var pwd sql.NullString
	var gid sql.NullString
	var last sql.NullTime
	var locked sql.NullTime
	if err := r.db.QueryRowContext(ctx, q, email).Scan(&u.ID, &u.Email, &u.Name, &pwd, &u.EmailVerified, &gid, &last, &u.FailedLoginAttempts, &locked, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	if pwd.Valid {
		p := pwd.String
		u.PasswordHash = &p
	}
	if gid.Valid {
		g := gid.String
		u.GoogleID = &g
	}
	if last.Valid {
		t := last.Time
		u.LastLogin = &t
	}
	if locked.Valid {
		t := locked.Time
		u.LockedUntil = &t
	}
	return &u, nil
}

// GetUserByGoogleID fetches a user by google_id
func (r *AuthUserRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*auth.User, error) {
	const q = `SELECT id, email, name, password_hash, email_verified, google_id, last_login, failed_login_attempts, locked_until, created_at, updated_at
               FROM users WHERE google_id=$1`
	var u auth.User
	var pwd sql.NullString
	var gid sql.NullString
	var last sql.NullTime
	var locked sql.NullTime
	if err := r.db.QueryRowContext(ctx, q, googleID).Scan(&u.ID, &u.Email, &u.Name, &pwd, &u.EmailVerified, &gid, &last, &u.FailedLoginAttempts, &locked, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	if pwd.Valid {
		p := pwd.String
		u.PasswordHash = &p
	}
	if gid.Valid {
		g := gid.String
		u.GoogleID = &g
	}
	if last.Valid {
		t := last.Time
		u.LastLogin = &t
	}
	if locked.Valid {
		t := locked.Time
		u.LockedUntil = &t
	}
	return &u, nil
}

// UpdateUser updates basic fields
func (r *AuthUserRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	const q = `UPDATE users SET email=$1, name=$2, password_hash=$3, email_verified=$4, google_id=$5, updated_at=NOW() WHERE id=$6`
	_, err := r.db.ExecContext(ctx, q, user.Email, user.Name, user.PasswordHash, user.EmailVerified, user.GoogleID, user.ID)
	return err
}

// UpdateUserLoginAttempts updates failed attempts and locked_until
func (r *AuthUserRepository) UpdateUserLoginAttempts(ctx context.Context, userID string, attempts int, lockedUntil *time.Time) error {
	const q = `UPDATE users SET failed_login_attempts=$1, locked_until=$2, updated_at=NOW() WHERE id=$3`
	_, err := r.db.ExecContext(ctx, q, attempts, lockedUntil, userID)
	return err
}

// UpdateLastLogin sets last_login to now
func (r *AuthUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET last_login=NOW(), updated_at=NOW() WHERE id=$1`, userID)
	return err
}

// EmailExists checks if email is registered
func (r *AuthUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE LOWER(email)=LOWER($1))`, strings.ToLower(email)).Scan(&exists)
	return exists, err
}

// RefreshTokenRepository implements auth.RefreshTokenRepository using Postgres
type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) CreateRefreshToken(ctx context.Context, token *auth.RefreshToken) error {
	const q = `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, revoked_at) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := r.db.ExecContext(ctx, q, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.CreatedAt, token.RevokedAt)
	return err
}

func (r *RefreshTokenRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*auth.RefreshToken, error) {
	const q = `SELECT id, user_id, token_hash, expires_at, created_at, revoked_at FROM refresh_tokens WHERE token_hash=$1`
	var t auth.RefreshToken
	var revoked sql.NullTime
	if err := r.db.QueryRowContext(ctx, q, tokenHash).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &revoked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("refresh token not found")
		}
		return nil, err
	}
	if revoked.Valid {
		x := revoked.Time
		t.RevokedAt = &x
	}
	return &t, nil
}

func (r *RefreshTokenRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at=NOW() WHERE token_hash=$1 AND revoked_at IS NULL`, tokenHash)
	return err
}

func (r *RefreshTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET revoked_at=NOW() WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	return err
}

func (r *RefreshTokenRepository) DeleteExpiredTokens(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM refresh_tokens WHERE (revoked_at IS NOT NULL) OR (expires_at < NOW())`)
	return err
}

func (r *RefreshTokenRepository) CountActiveTokensForUser(ctx context.Context, userID string) (int, error) {
	var cnt int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id=$1 AND revoked_at IS NULL AND expires_at > NOW()`, userID).Scan(&cnt)
	return cnt, err
}
