package auth

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"modernc.org/sqlite"
)

type (
	WebAuthSuite struct {
		suite.Suite
	}
)

type roConnector struct {
	d *sqlite.Driver
	n string
}

func (c *roConnector) Driver() driver.Driver { return c.d }

func (c *roConnector) Connect(_ context.Context) (driver.Conn, error) {
	return c.d.Open(c.n)
}

func (c *roConnector) Close() error { return nil }

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

func (s WebAuthSuite) TestIsSQLiteDatabaseDirectory() {
	s.Equal(false, isSQLiteDatabase("/tmp"))
}

func (s WebAuthSuite) TestIsSQLiteDatabaseNonExistent() {
	s.Equal(false, isSQLiteDatabase("nonexistent.db"))
}

func (s WebAuthSuite) TestAuthenticateQueryErrorFallback() {
	auth := "query_error_auth.db"
	user := "username"
	pass := "password"

	_ = os.Remove(auth)

	db, err := sql.Open("sqlite", auth)
	s.Require().NoError(err)
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE users (
			username TEXT PRIMARY KEY
		)
	`)
	s.Require().NoError(err)

	exists := Authenticate(auth, user, pass)
	s.Equal(false, exists)

	_ = os.Remove(auth)
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

func (s WebAuthSuite) TestAuthenticateScannerErrorFallback() {
	auth := "scanner_error_auth.db"
	user := "username"
	pass := "password"

	_ = os.Remove(auth)

	f, err := os.OpenFile(auth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	s.Require().NoError(err)
	_, err = f.WriteString(strings.Repeat("A", 70000) + "\n")
	s.Require().NoError(err)
	f.Close()
	defer os.Remove(auth)

	exists := Authenticate(auth, user, pass)
	s.Equal(false, exists)
}

func (s WebAuthSuite) TestAuthenticateSQLOpenError() {
	auth := "sqlopen_error.db"

	_ = os.Remove(auth)

	err := os.WriteFile(auth, []byte(sqliteHeader), 0640)
	s.Require().NoError(err)
	defer os.Remove(auth)

	patch := monkey.Patch(sql.Open, func(driverName, dataSourceName string) (*sql.DB, error) {
		return nil, errors.New("mock sql.Open error")
	})
	defer patch.Unpatch()

	exists := Authenticate(auth, "username", "password")
	s.Equal(false, exists)
}

func (s WebAuthSuite) TestAuthenticateLegacyOpenError() {
	auth := "nonexistent_dir/legacy_open_error.db"

	exists := Authenticate(auth, "username", "password")
	s.Equal(false, exists)
}

func (s WebAuthSuite) TestInitDBCreateUserSettingsError() {
	authdb := "initdb_error.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	// Create the database file with just the users table, but no user_settings table
	// and make it read-only to cause CREATE TABLE to fail
	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	_, err = db.Exec("CREATE TABLE users (username TEXT PRIMARY KEY, password_hash TEXT NOT NULL)")
	s.Require().NoError(err)
	db.Close()

	// Open the file in read-only mode by using a URI.
	// The modernc sqlite driver supports ?mode=ro.
	patch := monkey.Patch(sql.Open, func(driverName, dataSourceName string) (*sql.DB, error) {
		// Build a read-only connection via sql.OpenDB so the patched sql.Open
		// is never called again (which would cause infinite recursion).
		return sql.OpenDB(&roConnector{d: &sqlite.Driver{}, n: "file:" + dataSourceName + "?mode=ro"}), nil
	})
	defer patch.Unpatch()

	_, err = initDB(authdb)
	s.NotEqual(nil, err)
}

func TestWebAuthSuite(t *testing.T) {
	suite.Run(t, new(WebAuthSuite))
}
