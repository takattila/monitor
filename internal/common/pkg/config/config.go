package config

import (
	"strings"

	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/monitor/pkg/logger"
	"github.com/takattila/settings-manager"
)

// GetBool returns the value associated with the key as a boolean.
func GetBool(s *settings.Settings, key string) bool {
	ret, err := s.GetBool(key)
	common.ErrorIfErr(err)
	return ret
}

// GetInt returns the value associated with the key as an integer.
func GetInt(s *settings.Settings, key string) int {
	ret, err := s.GetInt(key)
	common.ErrorIfErr(err)
	return ret
}

// GetString returns the value associated with the key as a string.
func GetString(s *settings.Settings, key string) string {
	ret, err := s.GetString(key)
	common.ErrorIfErr(err)
	return ret
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(s *settings.Settings, key string) []string {
	ret, err := s.GetStringSlice(key)
	common.ErrorIfErr(err)
	return ret
}

// GetLogLevel returns the value associated with the key as a logger.LogLevel.
func GetLogLevel(s *settings.Settings, key string) logger.LogLevel {
	level, err := s.GetString(key)
	common.ErrorIfErr(err)

	switch strings.ToLower(level) {
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warning":
		return logger.WarningLevel
	case "error":
		return logger.ErrorLevel
	case "fatal":
		return logger.FatalLevel
	default:
		return logger.NoneLevel
	}
}

// GetLogColor returns the value associated with the key as a logger.Color.
func GetLogColor(s *settings.Settings, key string) logger.Color {
	colorOn, err := s.GetBool(key)
	common.ErrorIfErr(err)

	if colorOn {
		return logger.ColorOn
	}

	return logger.ColorOff
}
