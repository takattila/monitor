package config

import (
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
)

func GetBool(s *settings.Settings, key string) bool {
	ret, err := s.GetBool(key)
	common.Error(err)
	return ret
}

func GetInt(s *settings.Settings, key string) int {
	ret, err := s.GetInt(key)
	common.Error(err)
	return ret
}

func GetString(s *settings.Settings, key string) string {
	ret, err := s.GetString(key)
	common.Error(err)
	return ret
}

func GetStringSlice(s *settings.Settings, key string) []string {
	ret, err := s.GetStringSlice(key)
	common.Error(err)
	return ret
}
