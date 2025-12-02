package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Scraper Scraper `yaml:"scraper"`
	Server  Server  `yaml:"server"`
}

type Scraper struct {
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

type Server struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Scraper.Interval != "" {
		if _, err := time.ParseDuration(c.Scraper.Interval); err != nil {
			return fmt.Errorf("scraper.interval: invalid duration format: %w", err)
		}
	}

	if c.Scraper.Timeout != "" {
		if _, err := time.ParseDuration(c.Scraper.Timeout); err != nil {
			return fmt.Errorf("scraper.timeout: invalid duration format: %w", err)
		}
	}

	return nil
}

func (c *Config) GetScraperInterval() time.Duration {
	if c.Scraper.Interval == "" {
		return 6 * time.Hour
	}
	d, _ := time.ParseDuration(c.Scraper.Interval)
	return d
}

func (c *Config) GetScraperTimeout() time.Duration {
	if c.Scraper.Timeout == "" {
		return 30 * time.Second
	}
	d, _ := time.ParseDuration(c.Scraper.Timeout)
	return d
}

func (c *Config) GetServerAddress() string {
	host := c.Server.Host
	if host == "" {
		host = "0.0.0.0"
	}
	port := c.Server.Port
	if port == "" {
		port = "8080"
	}
	return host + ":" + port
}
