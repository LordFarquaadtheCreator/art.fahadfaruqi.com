package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	R2 struct {
		AccountID       string `yaml:"account_id"`
		Bucket          string `yaml:"bucket"`
		APIToken        string `yaml:"api_token"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
		S3APIEndpoint   string `yaml:"s3_api_endpoint"`
		CDNBase         string `yaml:"cdn_base"`
	} `yaml:"r2"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if base := os.Getenv("CDN_BASE"); base != "" {
		cfg.R2.CDNBase = base
	}

	return &cfg, nil
}
