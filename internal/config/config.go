package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type NotifierConfig struct {
	BaseURL string `yaml:"base_url"`
}

type LoggingConfig struct {
	Level    string `yaml:"level"`
	FilePath string `yaml:"file_path"`
}

type MetricsConfig struct {
	Port int `yaml:"port"`
}

type SchedulerConfig struct {
	PollIntervalSeconds int `yaml:"poll_interval_seconds"`
}

type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	Notifier  NotifierConfig  `yaml:"notifier"`
	Logging   LoggingConfig   `yaml:"logging"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
