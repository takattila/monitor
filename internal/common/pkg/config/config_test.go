package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/settings-manager"
)

type (
	ConfigSuite struct {
		suite.Suite
	}
)

func (s ConfigSuite) TestTGetBool() {
	content := `
other:
  content:
    int: 1
    string: 'text'
    bool: true
    slice:
      - a
      - b
      - c
`
	sm := settings.NewFromContent(content)
	result := GetBool(sm, "other.content.bool")
	s.Equal(true, result)
}

func (s ConfigSuite) TestGetInt() {
	content := `
other:
  content:
    int: 1
    string: 'text'
    bool: true
    slice:
      - a
      - b
      - c
`
	sm := settings.NewFromContent(content)
	result := GetInt(sm, "other.content.int")
	s.Equal(1, result)
}

func (s ConfigSuite) TestGetString() {
	content := `
other:
  content:
    int: 1
    string: 'text'
    bool: true
    slice:
      - a
      - b
      - c
`
	sm := settings.NewFromContent(content)
	result := GetString(sm, "other.content.string")
	s.Equal("text", result)
}

func (s ConfigSuite) TestGetStringSlice() {
	content := `
other:
  content:
    int: 1
    string: 'text'
    bool: true
    slice:
      - a
      - b
      - c
`
	sm := settings.NewFromContent(content)
	result := GetStringSlice(sm, "other.content.slice")
	s.Equal([]string{"a", "b", "c"}, result)
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
