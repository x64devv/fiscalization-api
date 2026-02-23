package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Crypto   CryptoConfig   `yaml:"crypto"`
	Redis    RedisConfig    `yaml:"redis"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	SMS      SMSConfig      `yaml:"sms"`
}

type ServerConfig struct {
	Port         int    `yaml:"port"`
	Mode         string `yaml:"mode"` // development, production
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	DBName          string `yaml:"dbname"`
	SSLMode         string `yaml:"sslmode"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MigrationsPath  string `yaml:"migrations_path"`
}

type CryptoConfig struct {
	CertificatePath     string `yaml:"certificate_path"`
	PrivateKeyPath      string `yaml:"private_key_path"`
	CACertificatePath   string `yaml:"ca_certificate_path"`
	CertificateValidityDays int `yaml:"certificate_validity_days"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
}

type SMSConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	Sender   string `yaml:"sender"`
}

func Load() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if present
	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%d", &config.Server.Port)
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}

	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		config.Database.Password = dbPass
	}

	return &config, nil
}
