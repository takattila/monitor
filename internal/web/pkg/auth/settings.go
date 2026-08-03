package auth

import (
	"database/sql"
	"fmt"
)

// SaveUserSetting stores or updates a user setting (skin, css, logo, preset, etc.)
// in the user_settings table of the auth database.
func SaveUserSetting(authFile, username, key, value string) error {
	db, err := initDB(authFile)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(
		"INSERT OR REPLACE INTO user_settings (username, key_name, value) VALUES (?, ?, ?)",
		username, key, value,
	)
	if err != nil {
		return fmt.Errorf("INSERT user_settings: %w", err)
	}
	return nil
}

// GetUserSettings retrieves all settings for a given user as a map of key-value
// string pairs from the user_settings table.
func GetUserSettings(authFile, username string) (map[string]string, error) {
	db, err := initDB(authFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT key_name, value FROM user_settings WHERE username = ?",
		username,
	)
	if err != nil {
		return nil, fmt.Errorf("SELECT user_settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan user_settings row: %w", err)
		}
		settings[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return settings, nil
}

// GetUserSetting retrieves a single setting value for a given user and key.
// Returns the value and nil error if found, or an empty string and sql.ErrNoRows if not found.
func GetUserSetting(authFile, username, key string) (string, error) {
	db, err := initDB(authFile)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var value string
	err = db.QueryRow(
		"SELECT value FROM user_settings WHERE username = ? AND key_name = ?",
		username, key,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("SELECT user_settings: %w", err)
	}
	return value, nil
}
