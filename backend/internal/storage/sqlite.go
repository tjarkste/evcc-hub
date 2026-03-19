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

func (db *DB) migrate() error {
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
	return err
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

// AuthenticateMQTT looks up a user by MQTT username and verifies the MQTT password (plain comparison).
func (db *DB) AuthenticateMQTT(mqttUsername, mqttPassword string) (*models.User, error) {
	u, err := db.GetUserByMQTTUsername(mqttUsername)
	if err != nil {
		return nil, errors.New("invalid mqtt credentials")
	}
	if u.MQTTPassword != mqttPassword {
		return nil, errors.New("invalid mqtt credentials")
	}
	return u, nil
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
