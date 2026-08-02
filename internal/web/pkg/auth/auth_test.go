package auth

import (
	"database/sql"
	"encoding/base64"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/suite"
)

type (
	WebAuthSuite struct {
		suite.Suite
	}
)

func (s WebAuthSuite) TestAuthenticate() {
	auth := "auth.db"
	user := "username"
	pass := "password"

	_ = os.Remove(auth)

	db, err := sql.Open("sqlite", auth)
	s.Require().NoError(err)
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			username TEXT PRIMARY KEY,
			password_hash TEXT NOT NULL
		)
	`)
	s.Require().NoError(err)

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	s.Require().NoError(err)

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", user, string(hash))
	s.Require().NoError(err)

	exists := Authenticate(auth, "bad", "credentials")
	s.Equal(false, exists)

	exists = Authenticate(auth, user, pass)
	s.Equal(true, exists)

	exists = Authenticate("bad.db", user, pass)
	s.Equal(false, exists)

	_ = os.Remove(auth)
}

func (s WebAuthSuite) TestAuthenticateLegacyFallback() {
	auth := "legacy_auth.db"
	user := "legacyuser"
	pass := "legacypass"

	_ = os.Remove(auth)

	f, err := os.OpenFile(auth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	s.Require().NoError(err)
	defer f.Close()

	authString := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	_, err = f.WriteString(authString + "\n")
	s.Require().NoError(err)

	exists := Authenticate(auth, "bad", "credentials")
	s.Equal(false, exists)

	exists = Authenticate(auth, user, pass)
	s.Equal(true, exists)

	_ = os.Remove(auth)
}

func (s WebAuthSuite) TestAuthenticateNoTableFallback() {
	auth := "no_table_auth.db"
	user := "username"
	pass := "password"

	_ = os.Remove(auth)

	db, err := sql.Open("sqlite", auth)
	s.Require().NoError(err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE other_table (id INTEGER)")
	s.Require().NoError(err)

	exists := Authenticate(auth, user, pass)
	s.Equal(false, exists)

	_ = os.Remove(auth)
}

func (s WebAuthSuite) TestIsSQLiteDatabaseNonExistent() {
	s.Equal(false, isSQLiteDatabase("nonexistent.db"))
}

func (s WebAuthSuite) TestAuthenticateCorruptDBFallback() {
	auth := "corrupt_auth.db"
	user := "username"
	pass := "password"

	_ = os.Remove(auth)

	f, err := os.OpenFile(auth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	s.Require().NoError(err)
	_, err = f.WriteString("this is not a valid sqlite database\n")
	s.Require().NoError(err)
	f.Close()
	defer os.Remove(auth)

	exists := Authenticate(auth, user, pass)
	s.Equal(false, exists)
}

func TestWebAuthSuite(t *testing.T) {
	suite.Run(t, new(WebAuthSuite))
}
