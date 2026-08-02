package auth

import (
	"database/sql"
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

func TestWebAuthSuite(t *testing.T) {
	suite.Run(t, new(WebAuthSuite))
}
