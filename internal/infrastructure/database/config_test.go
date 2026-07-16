package database

import (
	"strings"
	"testing"
)

func TestLoadConfigFromEnvSuccess(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "1433")
	t.Setenv("DB_USER", "sa")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "card_transaction")
	t.Setenv("DB_ENCRYPT", "disable")
	t.Setenv("DB_TRUST_SERVER_CERTIFICATE", "true")
	t.Setenv("DB_MAX_OPEN_CONNS", "30")
	t.Setenv("DB_MAX_IDLE_CONNS", "15")
	t.Setenv("DB_PING_TIMEOUT", "8")

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("expected host localhost, got %s", cfg.Host)
	}

	if cfg.Port != 1433 {
		t.Fatalf("expected port 1433, got %d", cfg.Port)
	}

	if cfg.MaxOpenConns != 30 {
		t.Fatalf("expected max open conns 30, got %d", cfg.MaxOpenConns)
	}

	if !cfg.TrustServerCertificate {
		t.Fatalf("expected trust server certificate true")
	}

	dsn := cfg.DSN()
	if !strings.Contains(dsn, "sqlserver://sa:secret@localhost:1433") {
		t.Fatalf("expected sqlserver DSN host, got %s", dsn)
	}

	if !strings.Contains(dsn, "database=card_transaction") {
		t.Fatalf("expected sqlserver DSN database, got %s", dsn)
	}
}

func TestLoadConfigFromEnvMissingRequiredEnv(t *testing.T) {
	t.Setenv("DB_PORT", "1433")
	t.Setenv("DB_USER", "sa")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "card_transaction")

	_, err := LoadConfigFromEnv()
	if err == nil {
		t.Fatalf("expected error when DB_HOST is missing")
	}

	if !strings.Contains(err.Error(), "DB_HOST") {
		t.Fatalf("expected DB_HOST error, got %v", err)
	}
}
