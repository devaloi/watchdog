// Package config handles YAML configuration parsing and validation for watchdog.
package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level watchdog configuration.
type Config struct {
	Global Global `yaml:"global"`
	Rules  []Rule `yaml:"rules"`
}

// Global holds default settings applied to all rules.
type Global struct {
	Debounce Duration `yaml:"debounce"`
	Ignore   []string `yaml:"ignore"`
}

// Rule defines a single watch rule with patterns, event filters, and an action.
type Rule struct {
	Name     string   `yaml:"name"`
	Watch    []string `yaml:"watch"`
	Events   []string `yaml:"events"`
	Debounce Duration `yaml:"debounce"`
	Action   Action   `yaml:"action"`
}

// Action describes what to do when a rule matches.
type Action struct {
	Type    string            `yaml:"type"`
	Command string            `yaml:"command"`
	Dir     string            `yaml:"dir"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Timeout Duration          `yaml:"timeout"`
	Format  string            `yaml:"format"`
}

// Duration wraps time.Duration for YAML unmarshalling.
type Duration struct {
	time.Duration
}

// UnmarshalYAML parses a duration string like "500ms" or "2s".
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return err
	}

	d.Duration = parsed

	return nil
}

// MarshalYAML serializes the duration as a string.
func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// Load reads and parses a YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) //nolint:gosec // config path is user-provided by design
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

// Parse decodes YAML bytes into a validated Config.
func Parse(data []byte) (*Config, error) {
	var cfg Config

	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	err = validate(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func isValidActionType(t string) bool {
	switch t {
	case "command", "webhook", "log":
		return true
	default:
		return false
	}
}

func validate(cfg *Config) error {
	if len(cfg.Rules) == 0 {
		return errors.New("config: at least one rule is required")
	}

	for i, r := range cfg.Rules {
		if r.Name == "" {
			return errors.New("config: rule at index " + itoa(i) + " is missing a name")
		}

		if len(r.Watch) == 0 {
			return errors.New("config: rule " + r.Name + " must have at least one watch pattern")
		}

		if !isValidActionType(r.Action.Type) {
			return errors.New("config: rule " + r.Name + " has invalid action type: " + r.Action.Type)
		}

		err := validateAction(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateAction(r Rule) error {
	switch r.Action.Type {
	case "command":
		if r.Action.Command == "" {
			return errors.New("config: rule " + r.Name + " command action requires a command")
		}
	case "webhook":
		if r.Action.URL == "" {
			return errors.New("config: rule " + r.Name + " webhook action requires a url")
		}
	case "log":
		if r.Action.Format == "" {
			return errors.New("config: rule " + r.Name + " log action requires a format")
		}
	}

	return nil
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	buf := [20]byte{}
	pos := len(buf)

	for i > 0 {
		pos--

		buf[pos] = byte('0' + i%10)
		i /= 10
	}

	return string(buf[pos:])
}
