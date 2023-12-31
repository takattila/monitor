package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/takattila/monitor/pkg/logger"
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

func (s ConfigSuite) TestGetLogLevel() {
	for _, check := range []struct {
		replace string
		level   logger.LogLevel
	}{
		{
			replace: "debug",
			level:   logger.DebugLevel,
		},
		{
			replace: "info",
			level:   logger.InfoLevel,
		},
		{
			replace: "warning",
			level:   logger.WarningLevel,
		},
		{
			replace: "error",
			level:   logger.ErrorLevel,
		},
		{
			replace: "fatal",
			level:   logger.FatalLevel,
		},
		{
			replace: "none",
			level:   logger.NoneLevel,
		},
	} {
		content := `
logger:
  level: CHANGE_ME
  color: on
`

		content = strings.ReplaceAll(content, "CHANGE_ME", check.replace)

		sm := settings.NewFromContent(content)
		result := GetLogLevel(sm, "logger.level")
		s.Equal(check.level, result)
	}
}

func (s ConfigSuite) TestGetLogColor() {
	for _, check := range []struct {
		replace string
		color   logger.Color
	}{
		{
			replace: "on",
			color:   logger.ColorOn,
		},
		{
			replace: "off",
			color:   logger.ColorOff,
		},
	} {
		content := `
logger:
  level: CHANGE_ME
  color: on
`

		content = strings.ReplaceAll(content, "CHANGE_ME", check.replace)

		sm := settings.NewFromContent(content)
		result := GetLogColor(sm, "logger.level")
		s.Equal(check.color, result)
	}
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigSuite))
}
