package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Server struct {
	Port           int    `json:"port"`
	PublicHost     string `json:"publicHost"`
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

type Twitter struct {
	Handle string `json:"handle"`
	Token  string `json:"token"`
}

type AccountDetails struct {
	Name      string    `json:"name"`
	Summary   string    `json:"summary"`
	Twitter   *Twitter  `json:"twitter,omitempty"`
	StartDate time.Time `json:"startDate"`
}

type Settings struct {
	Server   *Server                   `json:"server,omitempty"`
	Cron     *Cron                     `json:"cron,omitempty"`
	SQLite   *SQLite                   `json:"sqlite,omitempty"`
	Accounts map[string]AccountDetails `json:"accounts,omitempty"`
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
