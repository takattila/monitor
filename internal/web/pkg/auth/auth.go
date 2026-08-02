package auth

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func initDB(authFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", authFile)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password_hash TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("CREATE TABLE: %w", err)
	}

	return db, nil
}

// Authenticate checks whether credentials exist in the auth database or not.
func Authenticate(authFile, name, pass string) bool {
	db, err := initDB(authFile)
	if err != nil {
		return false
	}
	defer db.Close()

	var hash string
	err = db.QueryRow("SELECT password_hash FROM users WHERE username = ?", name).Scan(&hash)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}
