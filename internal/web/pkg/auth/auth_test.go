package auth

// Testing:
// go test -coverprofile="coverage.out" -v ./...
// go tool cover -html="coverage.out"

import (
	"encoding/base64"
	"os"
	"testing"

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

	f, err := os.OpenFile(auth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		s.T().Fatalf("os.OpenFile: %v", err)
	}
	defer f.Close()

	authString := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	if _, err := f.WriteString(authString + "\n"); err != nil {
		s.T().Fatalf("f.WriteString: %v", err)
	}

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
