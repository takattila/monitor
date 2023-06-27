package terminal

import (
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

func (s WebSessionSuite) TestPrompt() {
	input := Prompt("username: ")
	s.Equal("", input)
}

func TestWebSessionSuite(t *testing.T) {
	suite.Run(t, new(WebSessionSuite))
}
