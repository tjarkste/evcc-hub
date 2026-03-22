package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"evcc-cloud/backend/internal/models"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Open opens (or creates) the SQLite database at the given path and runs migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Ping verifies the database connection is alive.
func (db *DB) Ping() error {
	return db.conn.Ping()
}

func (db *DB) migrate() error {
	// Original users table
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            TEXT PRIMARY KEY,
			email         TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			mqtt_username TEXT UNIQUE NOT NULL,
			mqtt_password TEXT NOT NULL,
			topic_prefix  TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("create users table: %w", err)
	}

	// Add updated_at to users (idempotent — ignore error if column exists)
	db.conn.Exec(`ALTER TABLE users ADD COLUMN updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`)

	// Sites table
	_, err = db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS sites (
			id            TEXT PRIMARY KEY,
			user_id       TEXT NOT NULL REFERENCES users(id),
			name          TEXT NOT NULL,
			mqtt_username TEXT UNIQUE NOT NULL,
			mqtt_password TEXT NOT NULL,
			topic_prefix  TEXT UNIQUE NOT NULL,
			timezone      TEXT,
			created_at    DATETIME NOT NULL,
			updated_at    DATETIME NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("create sites table: %w", err)
	}

	// Refresh tokens table
	_, err = db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id         TEXT PRIMARY KEY,
			user_id    TEXT NOT NULL REFERENCES users(id),
			token_hash TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL
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
	rows, err := db.conn.Query(`
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
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcryptCost)
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
	_, err = db.conn.Exec(
		`INSERT INTO users (id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, email, string(hash), mqttUsername, mqttPassword, topicPrefix, now,
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
	}, nil
}

// GetUserByEmail retrieves a user by email. Returns sql.ErrNoRows if not found.
func (db *DB) GetUserByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := db.conn.QueryRow(
		`SELECT id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at
		 FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.MQTTUsername, &u.MQTTPassword, &u.TopicPrefix, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUserByMQTTUsername retrieves a user by MQTT username. Returns sql.ErrNoRows if not found.
func (db *DB) GetUserByMQTTUsername(mqttUsername string) (*models.User, error) {
	u := &models.User{}
	err := db.conn.QueryRow(
		`SELECT id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at
		 FROM users WHERE mqtt_username = ?`, mqttUsername,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.MQTTUsername, &u.MQTTPassword, &u.TopicPrefix, &u.CreatedAt)
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
	// Use base64 URL encoding and strip non-alphanumeric chars until we have enough.
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

	_, err = db.conn.Exec(
		`INSERT INTO sites (id, user_id, name, mqtt_username, mqtt_password, topic_prefix, timezone, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
	rows, err := db.conn.Query(
		`SELECT id, user_id, name, mqtt_username, topic_prefix, timezone, created_at, updated_at
		 FROM sites WHERE user_id = ? ORDER BY created_at`, userID,
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

// UpdateSite updates a site's name and/or timezone.
func (db *DB) UpdateSite(siteID, userID string, name *string, timezone *string) (*models.Site, error) {
	// Verify ownership
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM sites WHERE id = ? AND user_id = ?`, siteID, userID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("site not found")
	}

	now := time.Now().UTC()
	if name != nil {
		if _, err := db.conn.Exec(`UPDATE sites SET name = ?, updated_at = ? WHERE id = ?`, *name, now, siteID); err != nil {
			return nil, err
		}
	}
	if timezone != nil {
		if _, err := db.conn.Exec(`UPDATE sites SET timezone = ?, updated_at = ? WHERE id = ?`, *timezone, now, siteID); err != nil {
			return nil, err
		}
	}

	// Return updated site
	var s models.Site
	err = db.conn.QueryRow(
		`SELECT id, user_id, name, mqtt_username, topic_prefix, timezone, created_at, updated_at
		 FROM sites WHERE id = ?`, siteID,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.MQTTUsername, &s.TopicPrefix, &s.Timezone, &s.CreatedAt, &s.UpdatedAt)
	return &s, err
}

// DeleteSite removes a site if it belongs to the given user.
func (db *DB) DeleteSite(siteID, userID string) error {
	result, err := db.conn.Exec(`DELETE FROM sites WHERE id = ? AND user_id = ?`, siteID, userID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return errors.New("site not found")
	}
	return nil
}

// GetSiteByMQTTUsername retrieves a site by its MQTT username.
func (db *DB) GetSiteByMQTTUsername(mqttUsername string) (*models.Site, error) {
	var s models.Site
	err := db.conn.QueryRow(
		`SELECT s.id, s.user_id, s.name, s.mqtt_username, s.mqtt_password, s.topic_prefix, s.timezone, s.created_at, s.updated_at
		 FROM sites s WHERE s.mqtt_username = ?`, mqttUsername,
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

	_, err := db.conn.Exec(
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
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
	err := db.conn.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens WHERE token_hash = ? AND expires_at > ?`,
		tokenHash, time.Now().UTC(),
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// DeleteRefreshToken removes a refresh token by its hash (used for rotation and logout).
func (db *DB) DeleteRefreshToken(tokenHash string) error {
	_, err := db.conn.Exec(`DELETE FROM refresh_tokens WHERE token_hash = ?`, tokenHash)
	return err
}

// DeleteRefreshTokensByUserID removes all refresh tokens for a user (force logout all sessions).
func (db *DB) DeleteRefreshTokensByUserID(userID string) error {
	_, err := db.conn.Exec(`DELETE FROM refresh_tokens WHERE user_id = ?`, userID)
	return err
}

// CleanupExpiredRefreshTokens removes all expired refresh tokens.
func (db *DB) CleanupExpiredRefreshTokens() (int64, error) {
	result, err := db.conn.Exec(`DELETE FROM refresh_tokens WHERE expires_at <= ?`, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CountSitesByUserID returns the number of sites for a user.
func (db *DB) CountSitesByUserID(userID string) (int, error) {
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM sites WHERE user_id = ?`, userID).Scan(&count)
	return count, err
}

// GetUserByID retrieves a user by their UUID.
func (db *DB) GetUserByID(userID string) (*models.User, error) {
	u := &models.User{}
	err := db.conn.QueryRow(
		`SELECT id, email, password_hash, mqtt_username, mqtt_password, topic_prefix, created_at
		 FROM users WHERE id = ?`, userID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.MQTTUsername, &u.MQTTPassword, &u.TopicPrefix, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
