package config

import (
	"time"

	yaml "gopkg.in/yaml.v3"
)

type AirlinesConfig struct {
	AirAsia         AirlineConfig `yaml:"airasia"`
	BatikAir        AirlineConfig `yaml:"batik_air"`
	GarudaIndonesia AirlineConfig `yaml:"garuda_indonesia"`
	LionAir         AirlineConfig `yaml:"lion_air"`
}

type AirlineConfig struct {
	Enabled   bool            `yaml:"enabled"`
	Timeout   Duration        `yaml:"timeout"`
	Mock      MockConfig      `yaml:"mock"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	Retry     RetryConfig     `yaml:"retry"`
}

type RateLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requests_per_second"`
	Burst             int     `yaml:"burst"`
}

type RetryConfig struct {
	MaxAttempts int      `yaml:"max_attempts"`
	BaseDelay   Duration `yaml:"base_delay"`
}

type MockConfig struct {
	FilePath    string   `yaml:"file_path"`
	MinDelay    Duration `yaml:"min_delay"`
	MaxDelay    Duration `yaml:"max_delay"`
	SuccessRate *float64 `yaml:"success_rate"`
}

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return err
	}

	*d = Duration(parsed)
	return nil
}

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}
