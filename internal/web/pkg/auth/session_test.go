package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Testing:
// go test -coverprofile="coverage.out" -v ./...
// go tool cover -html="coverage.out"

type (
	WebSessionSuite struct {
		suite.Suite
	}
)

func (s WebSessionSuite) TestSetSessionGetSession() {
	recorder := httptest.NewRecorder()
	SetSession("/monitor/web/", "username", recorder)

	request := &http.Request{Header: http.Header{"Cookie": recorder.HeaderMap["Set-Cookie"]}}
	s.Equal("username", GetUserName(request))
}

func (s WebSessionSuite) TestClearSession() {
	recorder := httptest.NewRecorder()
	ClearSession("/monitor/web/", recorder)

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
	s.Contains(fmt.Sprint(err), "no such file or directory")
}

func (s WebSessionSuite) TestSaveCredentialsWriteStringError() {
	authdb := "testauth.db"
	defer func() { _ = os.Remove(authdb) }()

	oldRriteString := writeString
	defer func() { writeString = oldRriteString }()

	oldTerminalPrompt := terminalPrompt
	defer func() { terminalPrompt = oldTerminalPrompt }()

	terminalPrompt = func(prompt string) string {
		return prompt
	}

	writeString = func(f *os.File, s string) (n int, err error) {
		return 0, fmt.Errorf("error: %s", "file.WriteString")
	}

	err := SaveCredentials(authdb, true)
	s.Contains(fmt.Sprint(err), "error: file.WriteString")
}

func (s WebSessionSuite) TestTerminalPrompt() {
	input := terminalPrompt("username: ")
	s.Equal("", input)
}

func TestWebSessionSuite(t *testing.T) {
	suite.Run(t, new(WebSessionSuite))
}
