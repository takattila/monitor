package auth

import (
	"database/sql"
	"errors"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/suite"
)

type (
	WebSettingsSuite struct {
		suite.Suite
	}
)

// ── SaveUserSetting tests ──────────────────────────────────────────────────

func (s WebSettingsSuite) TestSaveUserSetting() {
	authdb := "test_settings_save.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	err := SaveUserSetting(authdb, "user1", "css", "github_red")
	s.Equal(nil, err)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	defer db.Close()

	var key, value string
	err = db.QueryRow("SELECT key_name, value FROM user_settings WHERE username = ? AND key_name = ?", "user1", "css").Scan(&key, &value)
	s.Require().NoError(err)
	s.Equal("css", key)
	s.Equal("github_red", value)
}

func (s WebSettingsSuite) TestSaveUserSettingOverwrite() {
	authdb := "test_settings_overwrite.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	err := SaveUserSetting(authdb, "user1", "skin", "dark")
	s.Equal(nil, err)

	err = SaveUserSetting(authdb, "user1", "skin", "light")
	s.Equal(nil, err)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	defer db.Close()

	var value string
	err = db.QueryRow("SELECT value FROM user_settings WHERE username = ? AND key_name = ?", "user1", "skin").Scan(&value)
	s.Require().NoError(err)
	s.Equal("light", value)
}

func (s WebSettingsSuite) TestSaveUserSettingBadPath() {
	err := SaveUserSetting("/nonexistent/dir/test.db", "user1", "css", "github_red")
	s.NotEqual(nil, err)
}

func (s WebSettingsSuite) TestSaveUserSettingInsertError() {
	authdb := "insert_error_settings.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	_, err = db.Exec("CREATE TABLE user_settings (username TEXT, key_name TEXT)")
	s.Require().NoError(err)
	db.Close()

	err = SaveUserSetting(authdb, "user1", "css", "github_red")
	s.NotEqual(nil, err)
	s.Contains(err.Error(), "INSERT user_settings")
}

// ── GetUserSettings tests ──────────────────────────────────────────────────

func (s WebSettingsSuite) TestGetUserSettings() {
	authdb := "test_settings_get_all.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	s.Require().NoError(SaveUserSetting(authdb, "user1", "skin", "dark"))
	s.Require().NoError(SaveUserSetting(authdb, "user1", "css", "github_red"))
	s.Require().NoError(SaveUserSetting(authdb, "user1", "logo", "rpi"))
	s.Require().NoError(SaveUserSetting(authdb, "user1", "preset", "block"))

	settings, err := GetUserSettings(authdb, "user1")
	s.Equal(nil, err)
	s.Equal("dark", settings["skin"])
	s.Equal("github_red", settings["css"])
	s.Equal("rpi", settings["logo"])
	s.Equal("block", settings["preset"])
}

func (s WebSettingsSuite) TestGetUserSettingsEmpty() {
	authdb := "test_settings_empty.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	settings, err := GetUserSettings(authdb, "nonexistent_user")
	s.Equal(nil, err)
	s.Equal(0, len(settings))
}

func (s WebSettingsSuite) TestGetUserSettingsBadPath() {
	settings, err := GetUserSettings("/nonexistent/dir/test.db", "user1")
	s.NotEqual(nil, err)
	s.Nil(settings)
}

func (s WebSettingsSuite) TestGetUserSettingsSelectError() {
	authdb := "select_error_settings.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	_, err = db.Exec("CREATE TABLE user_settings (username TEXT, bogus TEXT)")
	s.Require().NoError(err)
	db.Close()

	settings, err := GetUserSettings(authdb, "user1")
	s.NotEqual(nil, err)
	s.Nil(settings)
}

func (s WebSettingsSuite) TestGetUserSettingsScanError() {
	authdb := "scan_error_settings.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	_, err = db.Exec("CREATE TABLE user_settings (username TEXT, key_name TEXT, value TEXT)")
	s.Require().NoError(err)
	_, err = db.Exec("INSERT INTO user_settings (username, key_name, value) VALUES (?, ?, NULL)", "user1", "css")
	s.Require().NoError(err)
	db.Close()

	settings, err := GetUserSettings(authdb, "user1")
	s.NotEqual(nil, err)
	s.Nil(settings)
}

func (s WebSettingsSuite) TestGetUserSettingsRowsError() {
	authdb := "rows_error_settings.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS user_settings (username TEXT, key_name TEXT, value TEXT)")
	s.Require().NoError(err)
	_, err = db.Exec("INSERT INTO user_settings (username, key_name, value) VALUES (?, ?, ?)", "user1", "css", "github_red")
	s.Require().NoError(err)
	db.Close()

	patch := monkey.PatchInstanceMethod(reflect.TypeOf((*sql.Rows)(nil)), "Err", func(rows *sql.Rows) error {
		return errors.New("rows iteration error")
	})
	defer patch.Unpatch()

	settings, err := GetUserSettings(authdb, "user1")
	s.NotEqual(nil, err)
	s.Nil(settings)
	s.Contains(err.Error(), "rows iteration error")
}

// ── GetUserSetting (single) tests ───────────────────────────────────────────

func (s WebSettingsSuite) TestGetUserSetting() {
	authdb := "test_settings_get_one.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	s.Require().NoError(SaveUserSetting(authdb, "user1", "css", "github_red"))

	value, err := GetUserSetting(authdb, "user1", "css")
	s.Equal(nil, err)
	s.Equal("github_red", value)
}

func (s WebSettingsSuite) TestGetUserSettingNotFound() {
	authdb := "test_settings_not_found.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	value, err := GetUserSetting(authdb, "user1", "css")
	s.Equal(nil, err)
	s.Equal("", value)
}

func (s WebSettingsSuite) TestGetUserSettingBadPath() {
	value, err := GetUserSetting("/nonexistent/dir/test.db", "user1", "css")
	s.NotEqual(nil, err)
	s.Equal("", value)
}

func (s WebSettingsSuite) TestGetUserSettingSelectError() {
	authdb := "select_error_getone.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	_, err = db.Exec("CREATE TABLE user_settings (username TEXT, key_name TEXT)")
	s.Require().NoError(err)
	db.Close()

	value, err := GetUserSetting(authdb, "user1", "css")
	s.NotEqual(nil, err)
	s.Equal("", value)
}

func (s WebSettingsSuite) TestGetUserSettingScanError() {
	authdb := "scan_error_getone.db"
	defer os.Remove(authdb)
	_ = os.Remove(authdb)

	db, err := sql.Open("sqlite", authdb)
	s.Require().NoError(err)
	_, err = db.Exec("CREATE TABLE user_settings (username TEXT, key_name TEXT, value TEXT)")
	s.Require().NoError(err)
	_, err = db.Exec("INSERT INTO user_settings (username, key_name, value) VALUES (?, ?, NULL)", "user1", "css")
	s.Require().NoError(err)
	db.Close()

	value, err := GetUserSetting(authdb, "user1", "css")
	s.NotEqual(nil, err)
	s.Equal("", value)
}

func TestWebSettingsSuite(t *testing.T) {
	suite.Run(t, new(WebSettingsSuite))
}
