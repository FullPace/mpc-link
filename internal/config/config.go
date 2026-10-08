// Package config holds MPC Link's settings: which MPC, which folders it shares, and how they sync.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Folder is a computer folder shared with the MPC: it shows up as <mount>/<Name> in the MPC's browser.
type Folder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Direction of a sync pair.
type Direction string

const (
	ToMPC      Direction = "to-mpc"      // computer -> MPC
	ToComputer Direction = "to-computer" // MPC -> computer
	BothWays   Direction = "both"        // keep both in step
)

// Pair keeps a computer folder and an MPC folder in sync.
type Pair struct {
	Local     string    `json:"local"`
	Remote    string    `json:"remote"` // path on the MPC, e.g. /media/MPC SSD/Splice
	Direction Direction `json:"direction"`
}

type Config struct {
	Host         string   `json:"host"`          // the MPC's address, e.g. 192.168.178.58
	ComputerName string   `json:"computer_name"` // the folder name on the MPC; empty: this computer's name
	Folders      []Folder `json:"folders"`
	Pairs        []Pair   `json:"pairs"`
	HostKey      string   `json:"host_key,omitempty"` // the MPC's SSH host key, remembered on first connect
}

// Dir is where the settings live: ~/Library/Application Support/MPC Link, %AppData%\MPC Link. Settings from the
// app's first name, "Padbangers Link", move there once.
func Dir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		base = "."
	}
	dir := filepath.Join(base, "MPC Link")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Rename(filepath.Join(base, "Padbangers Link"), dir)
	}
	return dir
}

func path() string { return filepath.Join(Dir(), "config.json") }

func Load() (*Config, error) {
	c := &Config{}
	data, err := os.ReadFile(path())
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	return c, json.Unmarshal(data, c)
}

func (c *Config) Save() error {
	if err := os.MkdirAll(Dir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := path() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path())
}

// SafeName makes a name usable as a folder on the MPC and in a shell command: letters, digits, space, - and _ only
// ("Mike's Mac Mini" -> "Mikes Mac Mini").
func SafeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	out := strings.Join(strings.Fields(b.String()), " ")
	if out == "" {
		return "Computer"
	}
	return out
}
