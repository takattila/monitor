package auth

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/securecookie"
	"github.com/takattila/monitor/internal/web/pkg/terminal"
	"golang.org/x/crypto/bcrypt"
)

var (
	CookieHandler = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))

	terminalPrompt = func(prompt string) string {
		return terminal.Prompt(prompt)
	}
)

// SetSession creates session cookie.
func SetSession(path, userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := CookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  path,
		}

		http.SetCookie(response, cookie)
	}
}

// ClearSession removes session cookie.
func ClearSession(path string, response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   path,
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

// GetUserName takes out userName from session cookie.
func GetUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = CookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

// SaveCredentials writes user credentials into the AuthFile (SQLite database).
func SaveCredentials(authFile string, saveCredentials bool) error {
	if saveCredentials == true || !fileExists(authFile) {
		user := terminalPrompt("username: ")
		pass := terminalPrompt("password: ")

		var legacyCreds []string
		if fileExists(authFile) && !isSQLiteDatabase(authFile) {
			legacyCreds, _ = readLegacyCredentials(authFile)
			backupPath := authFile + ".legacy"
			_ = os.Rename(authFile, backupPath)
		}

		db, err := initDB(authFile)
		if err != nil {
			return err
		}
		defer db.Close()

		hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
		}

		_, err = db.Exec("INSERT OR REPLACE INTO users (username, password_hash) VALUES (?, ?)", user, string(hash))
		if err != nil {
			return fmt.Errorf("INSERT: %w", err)
		}

		for _, cred := range legacyCreds {
			parts := strings.SplitN(cred, ":", 2)
			if len(parts) == 2 {
				hash, err := bcrypt.GenerateFromPassword([]byte(parts[1]), bcrypt.DefaultCost)
				if err != nil {
					return fmt.Errorf("bcrypt.GenerateFromPassword: %w", err)
				}
				_, err = db.Exec("INSERT OR IGNORE INTO users (username, password_hash) VALUES (?, ?)", parts[0], string(hash))
				if err != nil {
					return fmt.Errorf("INSERT: %w", err)
				}
			}
		}
	}
	return nil
}

func readLegacyCredentials(authFile string) ([]string, error) {
	file, err := os.Open(authFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var creds []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		decoded, err := base64.StdEncoding.DecodeString(line)
		if err == nil {
			creds = append(creds, string(decoded))
		}
	}
	return creds, scanner.Err()
}
