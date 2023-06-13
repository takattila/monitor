package auth

import (
	"encoding/base64"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/takattila/monitor/internal/web/pkg/terminal"
	"github.com/takattila/monitor/pkg/common"
)

var (
	CookieHandler = securecookie.New(
		securecookie.GenerateRandomKey(64),
		securecookie.GenerateRandomKey(32))
)

// SetSession creates session cookie.
func SetSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := CookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}

		http.SetCookie(response, cookie)
	}
}

// ClearSession removes session cookie.
func ClearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

// WriteSessionToFile saves session into SessionFile.
func WriteSessionToFile(session string, sessionFile string) error {
	f, err := os.OpenFile(sessionFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.WriteString(session + "\n"); err != nil {
		return err
	}
	return nil
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

// SaveCredentials writes user credentials into the AuthFile.
func SaveCredentials(authFile string, saveCredentials bool) error {
	if saveCredentials == true || !common.FileExists(authFile) {
		user := terminal.Prompt("username: ")
		pass := terminal.Prompt("password: ")

		f, err := os.OpenFile(authFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return err
		}
		defer f.Close()

		authString := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
		if _, err := f.WriteString(authString + "\n"); err != nil {
			return err
		}
	}
	return nil
}
