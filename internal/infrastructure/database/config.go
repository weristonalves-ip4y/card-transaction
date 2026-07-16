package database

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

const (
	defaultEncrypt                = "disable"
	defaultTrustServerCertificate = true
	defaultMaxOpenConns           = 20
	defaultMaxIdleConns           = 10
	defaultPingTimeout            = 5
)

type Config struct {
	Host                   string
	Port                   int
	User                   string
	Password               string
	Database               string
	Encrypt                string
	TrustServerCertificate bool
	MaxOpenConns           int
	MaxIdleConns           int
	PingTimeout            int
}

func LoadConfigFromEnv() (Config, error) {
	cfg := Config{}

	host, err := requiredEnv("DB_HOST")
	if err != nil {
		return Config{}, err
	}
	cfg.Host = host

	port, err := requiredEnvInt("DB_PORT")
	if err != nil {
		return Config{}, err
	}
	cfg.Port = port

	user, err := requiredEnv("DB_USER")
	if err != nil {
		return Config{}, err
	}
	cfg.User = user

	password, err := requiredEnv("DB_PASSWORD")
	if err != nil {
		return Config{}, err
	}
	cfg.Password = password

	database, err := requiredEnv("DB_NAME")
	if err != nil {
		return Config{}, err
	}
	cfg.Database = database

	cfg.Encrypt = optionalEnv("DB_ENCRYPT", defaultEncrypt)
	cfg.TrustServerCertificate = optionalEnvBool("DB_TRUST_SERVER_CERTIFICATE", defaultTrustServerCertificate)
	cfg.MaxOpenConns = optionalEnvInt("DB_MAX_OPEN_CONNS", defaultMaxOpenConns)
	cfg.MaxIdleConns = optionalEnvInt("DB_MAX_IDLE_CONNS", defaultMaxIdleConns)
	cfg.PingTimeout = optionalEnvInt("DB_PING_TIMEOUT", defaultPingTimeout)

	return cfg, nil
}

func (c Config) DSN() string {
	query := url.Values{}
	query.Set("database", c.Database)
	query.Set("encrypt", c.Encrypt)
	query.Set("TrustServerCertificate", strconv.FormatBool(c.TrustServerCertificate))

	endpoint := url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(c.User, c.Password),
		Host:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		RawQuery: query.Encode(),
	}

	return endpoint.String()
}

func requiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("environment variable %s is required", key)
	}
	return value, nil
}

func requiredEnvInt(key string) (int, error) {
	value, err := requiredEnv(key)
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", key, err)
	}

	if parsed <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return parsed, nil
}

func optionalEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func optionalEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	if parsed <= 0 {
		return fallback
	}

	return parsed
}

func optionalEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}
