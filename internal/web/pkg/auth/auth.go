package auth

import (
	"bufio"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const sqliteHeader = "SQLite format 3\x00"

func isSQLiteDatabase(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, 16)
	_, err = f.Read(header)
	if err != nil {
		return false
	}

	return string(header) == sqliteHeader
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) || info == nil {
		return false
	}
	return !info.IsDir()
}

func initDB(authFile string) (*sql.DB, error) {
	if fileExists(authFile) && !isSQLiteDatabase(authFile) {
		return nil, fmt.Errorf("not a valid SQLite database: %s", authFile)
	}

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

func authenticateLegacy(authFile, name, pass string) bool {
	file, err := os.Open(authFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	authString := base64.StdEncoding.EncodeToString([]byte(name + ":" + pass))

	for scanner.Scan() {
		if authString == scanner.Text() {
			return true
		}
	}

	return false
}

// Authenticate checks whether credentials exist in the auth database or not.
func Authenticate(authFile, name, pass string) bool {
	db, err := initDB(authFile)
	if err != nil {
		return authenticateLegacy(authFile, name, pass)
	}
	defer db.Close()

	var hash string
	err = db.QueryRow("SELECT password_hash FROM users WHERE username = ?", name).Scan(&hash)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		return authenticateLegacy(authFile, name, pass)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	return err == nil
}