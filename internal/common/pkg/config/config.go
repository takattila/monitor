package config

import (
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

// GetBool returns the value associated with the key as a boolean.
func GetBool(s *settings.Settings, key string) bool {
	ret, err := s.GetBool(key)
	common.Error(err)
	return ret
}

// GetInt returns the value associated with the key as an integer.
func GetInt(s *settings.Settings, key string) int {
	ret, err := s.GetInt(key)
	common.Error(err)
	return ret
}

// GetString returns the value associated with the key as a string.
func GetString(s *settings.Settings, key string) string {
	ret, err := s.GetString(key)
	common.Error(err)
	return ret
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(s *settings.Settings, key string) []string {
	ret, err := s.GetStringSlice(key)
	common.Error(err)
	return ret
}
