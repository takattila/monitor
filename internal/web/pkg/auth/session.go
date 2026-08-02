package auth

import (
	"fmt"
	"net/http"

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
	}
	return nil
}
