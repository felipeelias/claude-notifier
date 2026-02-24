package config

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/felipeelias/claude-notifier/internal/notifier"
)

// Global holds top-level configuration.
type Global struct {
	Timeout time.Duration `toml:"timeout"`
}

// Config is the top-level configuration file structure.
type Config struct {
	Global    Global                      `toml:"global"`
	Notifiers map[string][]toml.Primitive `toml:"notifiers"`
	meta      toml.MetaData
}

// Load reads and parses a TOML config file.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	defer func() { _ = f.Close() }()

	cfg := &Config{
		Global: Global{
			Timeout: 10 * time.Second,
		},
	}

	meta, err := toml.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	cfg.meta = meta

	return cfg, nil
}

// Decode unmarshals a plugin's TOML primitive into the given struct.
func (c *Config) Decode(p toml.Primitive, v any) error {
	return c.meta.PrimitiveDecode(p, v)
}

// DefaultPath returns the default config file path.
func DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.ExpandEnv("$HOME/.config")
	}
	return dir + "/claude-notifier/config.toml"
}

// Configurable is implemented by notifiers that provide sample config.
type Configurable interface {
	SampleConfig() string
}

// SampleConfig generates a sample config from all registered plugins.
func SampleConfig(reg *notifier.Registry) string {
	s := "# claude-notifier configuration\n\n"
	s += "[global]\ntimeout = \"10s\"\n\n"

	all := reg.All()
	names := make([]string, 0, len(all))
	for name := range all {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		n := all[name]()
		if c, ok := n.(Configurable); ok {
			s += c.SampleConfig() + "\n"
		} else {
			s += fmt.Sprintf("# [[notifiers.%s]]\n\n", name)
		}
	}
	return s
}
