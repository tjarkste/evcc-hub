package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"evcc-cloud/backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// BcryptCost is the bcrypt cost factor used for password hashing.
const BcryptCost = 12

// DB wraps a PostgreSQL connection pool.
type DB struct {
	pool *pgxpool.Pool
}

// Open connects to PostgreSQL using the given database URL and runs migrations.
func Open(databaseURL string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MinConns = 2
	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres: %w", err)
	}

	db := &DB{pool: pool}
	if err := db.migrate(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// Close closes the underlying connection pool.
func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

// Ping verifies the database connection is alive.
func (db *DB) Ping() error {
	return db.pool.Ping(context.Background())
}

func (db *DB) migrate() error {
	_, err := db.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id            UUID PRIMARY KEY,
			email         TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			mqtt_username TEXT UNIQUE NOT NULL,
			mqtt_password TEXT NOT NULL,
			topic_prefix  TEXT NOT NULL,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	_, err = db.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS sites (
			id            UUID PRIMARY KEY,
			user_id       UUID NOT NULL REFERENCES users(id),
			name          TEXT NOT NULL,
			mqtt_username TEXT UNIQUE NOT NULL,
			mqtt_password TEXT NOT NULL,
			topic_prefix  TEXT UNIQUE NOT NULL,
			timezone      TEXT,
			created_at    TIMESTAMPTZ NOT NULL,
			updated_at    TIMESTAMPTZ NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create sites table: %w", err)
	}

	_, err = db.pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id         UUID PRIMARY KEY,
			user_id    UUID NOT NULL REFERENCES users(id),
			token_hash TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create refresh_tokens table: %w", err)
	}

	// Migrate existing users: create a default site for each user without one
	if err := db.migrateExistingUsersToSites(); err != nil {
		return fmt.Errorf("migrate users to sites: %w", err)
	}

	return nil
}

// migrateExistingUsersToSites creates a default "My Home" site for any user that doesn't have one yet.
func (db *DB) migrateExistingUsersToSites() error {
	rows, err := db.pool.Query(context.Background(), `
		SELECT u.id FROM users u
		LEFT JOIN sites s ON s.user_id = u.id
		WHERE s.id IS NULL
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		userIDs = append(userIDs, id)
	}

	for _, userID := range userIDs {
		if _, err := db.CreateSite(userID, "My Home", nil); err != nil {
			return fmt.Errorf("create default site for user %s: %w", userID, err)
		}
	}
	return nil
}

// TruncateAll removes all data from all tables. For test cleanup only.
func (db *DB) TruncateAll() error {
	_, err := db.pool.Exec(context.Background(), `TRUNCATE refresh_tokens, sites, users CASCADE`)
	return err
}

// EnsureDefaultSite creates a "My Home" site if the user has none. Called after CreateUser.
func (db *DB) EnsureDefaultSite(userID string) (*models.Site, error) {
	sites, err := db.GetSitesByUserID(userID)
	if err != nil {
		return nil, err
	}
	if len(sites) > 0 {
		return &sites[0], nil
	}
	return db.CreateSite(userID, "My Home", nil)
}

// CreateUser hashes the password, generates MQTT credentials, persists the user and returns it.
func (db *DB) CreateUser(email, plainPassword string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	userID := uuid.New().String()
	mqttUsername := GenerateMQTTUsername(userID)
	mqttPassword, err := GenerateRandomPassword(24)
	if err != nil {
		return nil, fmt.Errorf("generate mqtt password: %w", err)
	}
	topicPrefix := "user/" + userID + "/evcc"

	now := time.Now().UTC()
	_, err = db.pool.Exec(context.Background(),
		`INSERT INTO users (id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		userID, email, string(hash), mqttUsername, mqttPassword, topicPrefix, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	// Create default site for new user
	if _, err := db.EnsureDefaultSite(userID); err != nil {
		return nil, fmt.Errorf("create default site: %w", err)
	}

	return &models.User{
		ID:           userID,
		Email:        email,
		PasswordHash: string(hash),
		MQTTUsername: mqttUsername,
		MQTTPassword: mqttPassword,
		TopicPrefix:  topicPrefix,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetUserByEmail retrieves a user by email. Returns pgx.ErrNoRows if not found.
func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := db.pool.QueryRow(context.Background(),
		`SELECT id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.MQTTUsername, &u.MQTTPassword, &u.TopicPrefix, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByMQTTUsername retrieves a user by MQTT username. Returns pgx.ErrNoRows if not found.
func (db *DB) GetUserByMQTTUsername(mqttUsername string) (*models.User, error) {
	u := &models.User{}
	err := db.pool.QueryRow(context.Background(),
		`SELECT id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at, updated_at
		 FROM users WHERE mqtt_username = $1`, mqttUsername,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.MQTTUsername, &u.MQTTPassword, &u.TopicPrefix, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByID retrieves a user by their UUID.
func (db *DB) GetUserByID(userID string) (*models.User, error) {
	u := &models.User{}
	err := db.pool.QueryRow(context.Background(),
		`SELECT id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at, updated_at
		 FROM users WHERE id = $1`, userID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.MQTTUsername, &u.MQTTPassword, &u.TopicPrefix, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// VerifyPassword checks a plain-text password against the stored hash.
func VerifyPassword(hash, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
}

// AuthenticateUser looks up a user by email and verifies the password.
// Returns the user on success, or an error if credentials are invalid.
func (db *DB) AuthenticateUser(email, plainPassword string) (*models.User, error) {
	u, err := db.GetUserByEmail(email)
	if err != nil {
		// Return a generic error so we don't leak whether the email exists.
		return nil, errors.New("invalid credentials")
	}
	if err := VerifyPassword(u.PasswordHash, plainPassword); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// MQTTCredentialType indicates whether the authenticated credential belongs to a user or a site.
type MQTTCredentialType int

const (
	MQTTCredUser MQTTCredentialType = iota
	MQTTCredSite
)

// MQTTAuthResult holds the result of MQTT authentication.
type MQTTAuthResult struct {
	CredType    MQTTCredentialType
	UserID      string
	TopicPrefix string
}

// LookupMQTTCredentialByUsername identifies whether a username belongs to a user or site (no password check).
func (db *DB) LookupMQTTCredentialByUsername(mqttUsername string) (*MQTTAuthResult, error) {
	u, err := db.GetUserByMQTTUsername(mqttUsername)
	if err == nil {
		return &MQTTAuthResult{
			CredType:    MQTTCredUser,
			UserID:      u.ID,
			TopicPrefix: "user/" + u.ID + "/site",
		}, nil
	}

	s, err := db.GetSiteByMQTTUsername(mqttUsername)
	if err == nil {
		return &MQTTAuthResult{
			CredType:    MQTTCredSite,
			UserID:      s.UserID,
			TopicPrefix: s.TopicPrefix,
		}, nil
	}

	return nil, errors.New("unknown mqtt username")
}

// AuthenticateMQTT checks both user and site credentials and returns the result.
func (db *DB) AuthenticateMQTT(mqttUsername, mqttPassword string) (*MQTTAuthResult, error) {
	u, err := db.GetUserByMQTTUsername(mqttUsername)
	if err == nil && u.MQTTPassword == mqttPassword {
		return &MQTTAuthResult{
			CredType:    MQTTCredUser,
			UserID:      u.ID,
			TopicPrefix: "user/" + u.ID + "/site",
		}, nil
	}

	s, err := db.GetSiteByMQTTUsername(mqttUsername)
	if err == nil && s.MQTTPassword == mqttPassword {
		return &MQTTAuthResult{
			CredType:    MQTTCredSite,
			UserID:      s.UserID,
			TopicPrefix: s.TopicPrefix,
		}, nil
	}

	return nil, errors.New("invalid mqtt credentials")
}

// GenerateMQTTUsername derives a stable MQTT username from a user UUID.
func GenerateMQTTUsername(userID string) string {
	return "user_" + strings.ReplaceAll(userID, "-", "")[:16]
}

// GenerateRandomPassword generates a cryptographically random alphanumeric password of length n.
func GenerateRandomPassword(n int) (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 0, n)
	for len(result) < n {
		buf := make([]byte, n*2)
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		encoded := base64.RawURLEncoding.EncodeToString(buf)
		for _, c := range encoded {
			if strings.ContainsRune(alphabet, c) {
				result = append(result, byte(c))
				if len(result) == n {
					break
				}
			}
		}
	}
	return string(result), nil
}

// CreateSite creates a new site with generated MQTT credentials.
func (db *DB) CreateSite(userID, name string, timezone *string) (*models.Site, error) {
	siteID := uuid.New().String()
	mqttUsername := "site_" + strings.ReplaceAll(siteID, "-", "")[:16]
	mqttPassword, err := GenerateRandomPassword(24)
	if err != nil {
		return nil, fmt.Errorf("generate site mqtt password: %w", err)
	}
	topicPrefix := fmt.Sprintf("user/%s/site/%s/evcc", userID, siteID)
	now := time.Now().UTC()

	_, err = db.pool.Exec(context.Background(),
		`INSERT INTO sites (id, user_id, name, mqtt_username, mqtt_password, topic_prefix, timezone, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		siteID, userID, name, mqttUsername, mqttPassword, topicPrefix, timezone, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert site: %w", err)
	}

	return &models.Site{
		ID:           siteID,
		UserID:       userID,
		Name:         name,
		MQTTUsername: mqttUsername,
		MQTTPassword: mqttPassword,
		TopicPrefix:  topicPrefix,
		Timezone:     timezone,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// GetSitesByUserID returns all sites belonging to a user.
func (db *DB) GetSitesByUserID(userID string) ([]models.Site, error) {
	rows, err := db.pool.Query(context.Background(),
		`SELECT id, user_id, name, mqtt_username, topic_prefix, timezone, created_at, updated_at
		 FROM sites WHERE user_id = $1 ORDER BY created_at`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var s models.Site
		if err := rows.Scan(&s.ID, &s.UserID, &s.Name, &s.MQTTUsername, &s.TopicPrefix, &s.Timezone, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		sites = append(sites, s)
	}
	return sites, nil
}

// GetSiteByID retrieves a site by its ID, scoped to a user.
func (db *DB) GetSiteByID(siteID, userID string) (*models.Site, error) {
	var s models.Site
	err := db.pool.QueryRow(context.Background(),
		`SELECT id, user_id, name, mqtt_username, mqtt_password, topic_prefix, timezone, created_at, updated_at
		 FROM sites WHERE id = $1 AND user_id = $2`, siteID, userID,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.MQTTUsername, &s.MQTTPassword, &s.TopicPrefix, &s.Timezone, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSite updates a site's name and/or timezone.
func (db *DB) UpdateSite(siteID, userID string, name *string, timezone *string) (*models.Site, error) {
	// Verify ownership
	var count int
	err := db.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM sites WHERE id = $1 AND user_id = $2`, siteID, userID,
	).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("site not found")
	}

	now := time.Now().UTC()
	if name != nil {
		if _, err := db.pool.Exec(context.Background(),
			`UPDATE sites SET name = $1, updated_at = $2 WHERE id = $3`, *name, now, siteID,
		); err != nil {
			return nil, err
		}
	}
	if timezone != nil {
		if _, err := db.pool.Exec(context.Background(),
			`UPDATE sites SET timezone = $1, updated_at = $2 WHERE id = $3`, *timezone, now, siteID,
		); err != nil {
			return nil, err
		}
	}

	// Return updated site
	var s models.Site
	err = db.pool.QueryRow(context.Background(),
		`SELECT id, user_id, name, mqtt_username, topic_prefix, timezone, created_at, updated_at
		 FROM sites WHERE id = $1`, siteID,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.MQTTUsername, &s.TopicPrefix, &s.Timezone, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

// DeleteSite removes a site if it belongs to the given user.
func (db *DB) DeleteSite(siteID, userID string) error {
	result, err := db.pool.Exec(context.Background(),
		`DELETE FROM sites WHERE id = $1 AND user_id = $2`, siteID, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("site not found")
	}
	return nil
}

// GetSiteByMQTTUsername retrieves a site by its MQTT username.
func (db *DB) GetSiteByMQTTUsername(mqttUsername string) (*models.Site, error) {
	var s models.Site
	err := db.pool.QueryRow(context.Background(),
		`SELECT s.id, s.user_id, s.name, s.mqtt_username, s.mqtt_password, s.topic_prefix, s.timezone, s.created_at, s.updated_at
		 FROM sites s WHERE s.mqtt_username = $1`, mqttUsername,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.MQTTUsername, &s.MQTTPassword, &s.TopicPrefix, &s.Timezone, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

const refreshTokenDuration = 30 * 24 * time.Hour // 30 days

// CreateRefreshToken stores a hashed refresh token for a user, returning the row ID.
func (db *DB) CreateRefreshToken(userID, tokenHash string) (*models.RefreshToken, error) {
	id := uuid.New().String()
	now := time.Now().UTC()
	expiresAt := now.Add(refreshTokenDuration)

	_, err := db.pool.Exec(context.Background(),
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, userID, tokenHash, expiresAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert refresh token: %w", err)
	}
	return &models.RefreshToken{
		ID:        id,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

// GetRefreshTokenByHash looks up a non-expired refresh token by its hash.
func (db *DB) GetRefreshTokenByHash(tokenHash string) (*models.RefreshToken, error) {
	var rt models.RefreshToken
	err := db.pool.QueryRow(context.Background(),
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1 AND expires_at > $2`,
		tokenHash, time.Now().UTC(),
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// DeleteRefreshToken removes a refresh token by its hash (used for rotation and logout).
func (db *DB) DeleteRefreshToken(tokenHash string) error {
	_, err := db.pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash,
	)
	return err
}

// DeleteRefreshTokensByUserID removes all refresh tokens for a user (force logout all sessions).
func (db *DB) DeleteRefreshTokensByUserID(userID string) error {
	_, err := db.pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE user_id = $1`, userID,
	)
	return err
}

// CleanupExpiredRefreshTokens removes all expired refresh tokens.
func (db *DB) CleanupExpiredRefreshTokens() (int64, error) {
	result, err := db.pool.Exec(context.Background(),
		`DELETE FROM refresh_tokens WHERE expires_at <= $1`, time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// CountSitesByUserID returns the number of sites for a user.
func (db *DB) CountSitesByUserID(userID string) (int, error) {
	var count int
	err := db.pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM sites WHERE user_id = $1`, userID,
	).Scan(&count)
	return count, err
}

// UpdateUserPassword updates a user's password hash and sets updated_at.
func (db *DB) UpdateUserPassword(userID, newHash string) error {
	_, err := db.pool.Exec(context.Background(),
		`UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		newHash, userID,
	)
	return err
}
