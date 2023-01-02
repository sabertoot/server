package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sabertoot/server/internal/uid"
)

type Server struct {
	Port           int    `json:"port"`
	Domain         string `json:"domain"`
	PublicBaseURL  string `json:"publicBaseURL"`
	MaxHeaderBytes int    `json:"maxHeaderBytes"`
}

type Cron struct {
	IntervalSeconds int `json:"intervalSeconds"`
}

func (c *Cron) Interval() time.Duration {
	return time.Duration(c.IntervalSeconds) * time.Second
}

type SQLite struct {
	DSN string `json:"dsn"`
}

type Storage struct {
	Path string `json:"path"`
}

func (s *Storage) ProfileImageDirectory() string {
	return fmt.Sprintf("%s/profile_images", s.Path)
}

func (s *Storage) ProfileImageFullFilePath(userID uid.UserID, ext string) string {
	return fmt.Sprintf(
		"%s/%d%s",
		s.ProfileImageDirectory(),
		userID.Int(),
		ext)
}

func (s *Storage) ProfileImageRelativeURLPath(userID uid.UserID) string {
	return fmt.Sprintf("/profile_images/%d", userID)
}

type Twitter struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	FullName  string    `json:"fullName"`
	Summary   string    `json:"summary"`
	Twitter   *Twitter  `json:"twitter,omitempty"`
	StartDate time.Time `json:"startDate"`
}

func (u *User) UserID() uid.UserID {
	return uid.UserID(u.ID)
}

type Settings struct {
	Server  *Server  `json:"server,omitempty"`
	Cron    *Cron    `json:"cron,omitempty"`
	SQLite  *SQLite  `json:"sqlite,omitempty"`
	Storage *Storage `json:"storage,omitempty"`
	Users   []User   `json:"users,omitempty"`
}

func (s *Settings) Validate() (bool, error) {
	if s.Server == nil {
		return false, fmt.Errorf("server settings are missing")
	}
	if s.Cron == nil {
		return false, fmt.Errorf("cron settings are missing")
	}
	if s.SQLite == nil {
		return false, fmt.Errorf("sqlite settings are missing")
	}
	if len(s.Users) == 0 {
		return false, fmt.Errorf("no users are configured")
	}
	for _, user := range s.Users {
		if user.Twitter == nil {
			return false, fmt.Errorf("user %s is missing twitter settings", user.Username)
		}
	}

	// ToDo finish validation

	return true, nil
}

func Load() (*Settings, error) {
	filepath := os.Getenv("SETTINGS_PATH")
	if filepath == "" {
		filepath = "settings.json"
	}
	return LoadFrom(filepath)
}

func LoadFrom(filepath string) (*Settings, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading settings.json file: %w", err)
	}
	var settings Settings
	if err = json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("error deserializing settings.json file: %w", err)
	}
	return &settings, nil
}
