package config

import "testing"

func TestGetConfigUsesProdValues(t *testing.T) {
	cfg := getConfig(EnvProd)

	if cfg.Server != "192.168.4.123" {
		t.Fatalf("expected prod server, got %q", cfg.Server)
	}
	if cfg.Port != 1433 {
		t.Fatalf("expected prod port 1433, got %d", cfg.Port)
	}
	if cfg.User != "lognorm" {
		t.Fatalf("expected prod user, got %q", cfg.User)
	}
	if cfg.Password != "UBjsc@norm.nrp" {
		t.Fatalf("expected prod password, got %q", cfg.Password)
	}
	if cfg.Database != "norm" {
		t.Fatalf("expected prod database, got %q", cfg.Database)
	}
}

func TestGetConfigDefaultsToTestValues(t *testing.T) {
	cfg := getConfig("anything-else")

	if cfg.Server != "172.30.30.30" {
		t.Fatalf("expected test server, got %q", cfg.Server)
	}
	if cfg.Port != 1433 {
		t.Fatalf("expected test port 1433, got %d", cfg.Port)
	}
	if cfg.User != "sa" {
		t.Fatalf("expected test user, got %q", cfg.User)
	}
	if cfg.Password != "test" {
		t.Fatalf("expected test password, got %q", cfg.Password)
	}
	if cfg.Database != "test" {
		t.Fatalf("expected test database, got %q", cfg.Database)
	}
}

func TestGetDBConfigUsesAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", EnvProd)

	cfg := GetDBConfig()

	if cfg.Database != "norm" {
		t.Fatalf("expected prod database from app env, got %q", cfg.Database)
	}
}
