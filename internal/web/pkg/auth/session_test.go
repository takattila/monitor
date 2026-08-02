package auth

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/suite"
)

type (
	WebSessionSuite struct {
		suite.Suite
	}
)

func (s WebSessionSuite) TestSetSessionGetSession() {
	recorder := httptest.NewRecorder()
	SetSession("/monitor/", "username", recorder)

	request := &http.Request{Header: http.Header{"Cookie": recorder.HeaderMap["Set-Cookie"]}}
	s.Equal("username", GetUserName(request))
}

func (s WebSessionSuite) TestClearSession() {
	recorder := httptest.NewRecorder()
	ClearSession("/monitor/", recorder)

	request := &http.Request{Header: http.Header{"Cookie": recorder.HeaderMap["Set-Cookie"]}}
	s.Equal("", GetUserName(request))
}

func (s WebSessionSuite) TestSaveCredentials() {
	authdb := "testauth.db"
	defer func() { _ = os.Remove(authdb) }()

	oldTerminalPrompt := terminalPrompt
	defer func() { terminalPrompt = oldTerminalPrompt }()

	terminalPrompt = func(prompt string) string {
		return prompt
	}

	err := SaveCredentials(authdb, true)
	s.Equal(nil, err)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	defer db.Close()

	var hash string
	err = db.QueryRow("SELECT password_hash FROM users WHERE username = ?", "username: ").Scan(&hash)
	s.Require().NoError(err)

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("password: "))
	s.Equal(nil, err)
}

func (s WebSessionSuite) TestSaveCredentialsBadPathError() {
	authdb := "/bad/path/testauth.db"
	defer func() { _ = os.Remove(authdb) }()

	oldTerminalPrompt := terminalPrompt
	defer func() { terminalPrompt = oldTerminalPrompt }()

	terminalPrompt = func(prompt string) string {
		return prompt
	}

	err := SaveCredentials(authdb, true)
	s.Contains(fmt.Sprint(err), "unable to open database file")
}

func (s WebSessionSuite) TestSaveCredentialsExistingUser() {
	authdb := "testauth2.db"
	defer func() { _ = os.Remove(authdb) }()

	oldTerminalPrompt := terminalPrompt
	defer func() { terminalPrompt = oldTerminalPrompt }()

	terminalPrompt = func(prompt string) string {
		return prompt
	}

	err := SaveCredentials(authdb, true)
	s.Equal(nil, err)

	err = SaveCredentials(authdb, true)
	s.Equal(nil, err)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "username: ").Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)
}

func (s WebSessionSuite) TestSaveCredentialsMigrateLegacy() {
	authdb := "testauth_legacy.db"
	defer func() {
		_ = os.Remove(authdb)
		_ = os.Remove(authdb + ".legacy")
	}()

	oldTerminalPrompt := terminalPrompt
	defer func() { terminalPrompt = oldTerminalPrompt }()

	terminalPrompt = func(prompt string) string {
		return prompt
	}

	f, err := os.OpenFile(authdb, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	s.Require().NoError(err)
	authString := base64.StdEncoding.EncodeToString([]byte("legacyuser:legacypass"))
	_, err = f.WriteString(authString + "\n")
	s.Require().NoError(err)
	f.Close()

	err = SaveCredentials(authdb, true)
	s.Equal(nil, err)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", "legacyuser").Scan(&count)
	s.Require().NoError(err)
	s.Equal(1, count)

	_, err = os.Stat(authdb + ".legacy")
	s.Require().NoError(err)
}

func (s WebSessionSuite) TestReadLegacyCredentialsNonExistent() {
	creds, err := readLegacyCredentials("nonexistent_legacy.db")
	s.Equal(0, len(creds))
	s.NotEqual(nil, err)
}

func (s WebSessionSuite) TestReadLegacyCredentialsCorrupted() {
	authdb := "corrupt_legacy.db"
	defer func() { _ = os.Remove(authdb) }()

	f, err := os.OpenFile(authdb, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	s.Require().NoError(err)
	_, err = f.WriteString("this is not valid base64!!!\n")
	s.Require().NoError(err)
	f.Close()

	creds, err := readLegacyCredentials(authdb)
	s.Equal(nil, err)
	s.Equal(0, len(creds))
}

func (s WebSessionSuite) TestReadLegacyCredentialsMixed() {
	authdb := "mixed_legacy.db"
	defer func() { _ = os.Remove(authdb) }()

	f, err := os.OpenFile(authdb, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	s.Require().NoError(err)
	validCred := base64.StdEncoding.EncodeToString([]byte("validuser:validpass"))
	_, err = f.WriteString(validCred + "\n")
	s.Require().NoError(err)
	_, err = f.WriteString("not valid base64!!!\n")
	s.Require().NoError(err)
	f.Close()

	creds, err := readLegacyCredentials(authdb)
	s.Equal(nil, err)
	s.Equal(1, len(creds))
	s.Equal("validuser:validpass", creds[0])
}

func (s WebSessionSuite) TestTerminalPrompt() {
	input := terminalPrompt("username: ")
	s.Equal("", input)
}

func TestWebSessionSuite(t *testing.T) {
	suite.Run(t, new(WebSessionSuite))
}
