package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var (
	ErrUserNotFound    = errors.New("kullanıcı bulunamadı")
	ErrInvalidPassword = errors.New("şifre hatalı")
	ErrDBNotReady      = errors.New("veritabanı hazır değil")
)

type Database struct {
	db *sql.DB
	mu sync.RWMutex
}

var (
	globalDB   *Database
	globalDBMu sync.Mutex
)

// DB returns the global Database instance.
func DB() *Database {
	return globalDB
}

// InitDB opens the SQLite database and applies WAL pragmas.
func InitDB(dbPath string) (*Database, error) {
	globalDBMu.Lock()
	defer globalDBMu.Unlock()

	if globalDB != nil {
		return globalDB, nil
	}

	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	// Performance pragmas matching anthology-dav
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA busy_timeout = 5000;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("failed to set pragma %q: %w", p, err)
		}
	}

	d := &Database{db: db}
	if err := d.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	globalDB = d
	log.Printf("📦 [SQLite] Veritabanı bağlandı: %s", dbPath)
	return d, nil
}

func (d *Database) migrate() error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL COLLATE NOCASE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := d.db.Exec(query)
	return err
}

// Close closes the database connection.
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// Authenticate checks user credentials using bcrypt.
func (d *Database) Authenticate(username, password string) (*User, error) {
	if d == nil || d.db == nil {
		return nil, ErrDBNotReady
	}

	u, err := d.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return u, nil
}

// GetUserByUsername retrieves a user by case-insensitive username.
func (d *Database) GetUserByUsername(username string) (*User, error) {
	if d == nil || d.db == nil {
		return nil, ErrDBNotReady
	}

	username = strings.TrimSpace(username)
	if username == "" {
		return nil, ErrUserNotFound
	}

	row := d.db.QueryRow(`SELECT id, username, password_hash, created_at FROM users WHERE username = ? COLLATE NOCASE`, username)
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

// HasUsers checks if there are any registered users.
func (d *Database) HasUsers() bool {
	if d == nil || d.db == nil {
		return false
	}
	var count int
	err := d.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return err == nil && count > 0
}

// Register creates a new user if needed.
func (d *Database) Register(username, password string) (*User, error) {
	if d == nil || d.db == nil {
		return nil, ErrDBNotReady
	}
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(password) < 4 {
		return nil, errors.New("kullanıcı adı en az 3, şifre en az 4 karakter olmalıdır")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	res, err := d.db.Exec(`INSERT INTO users (username, password_hash, created_at) VALUES (?, ?, ?)`,
		username, string(hash), now)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           id,
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    now,
	}, nil
}
