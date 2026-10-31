package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAuthWithSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_anthology.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()

	// 1. Initial state - no users
	if db.HasUsers() {
		t.Errorf("Expected HasUsers to be false initially")
	}

	// 2. Register user
	user, err := db.Register("falsis", "secret1234")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if user.Username != "falsis" {
		t.Errorf("Expected username falsis, got %s", user.Username)
	}

	// 3. Authenticate with correct password
	authU, err := db.Authenticate("falsis", "secret1234")
	if err != nil {
		t.Fatalf("Authenticate failed with correct password: %v", err)
	}
	if authU.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, authU.ID)
	}

	// 4. Case-insensitivity check
	authCase, err := db.Authenticate("FaLsIs", "secret1234")
	if err != nil {
		t.Fatalf("Authenticate failed with mixed-case username: %v", err)
	}
	if authCase.Username != "falsis" {
		t.Errorf("Expected canonical username falsis, got %s", authCase.Username)
	}

	// 5. Authenticate with wrong password
	_, err = db.Authenticate("falsis", "wrongpass")
	if err == nil {
		t.Errorf("Expected error with wrong password, got nil")
	}

	// 6. Session token verification
	token := GenerateSessionToken("falsis")
	u, ok := ValidateSessionToken(token)
	if !ok || u != "falsis" {
		t.Errorf("ValidateSessionToken failed: %s, ok=%v", u, ok)
	}

	_ = os.Remove(dbPath)
}
