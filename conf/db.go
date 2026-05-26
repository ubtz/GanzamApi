package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type DBConfig struct {
	Server   string
	Port     int
	User     string
	Password string
	Database string
}

func getConfig(env string) DBConfig {
	if env == EnvProd {
		return DBConfig{
			Server:   "192.168.4.123",
			Port:     1433,
			User:     "lognorm",
			Password: "UBjsc@norm.nrp",
			Database: "Ganzam",
		}
	}

	return DBConfig{
		Server:   "172.30.30.30",
		Port:     1433,
		User:     "sa",
		Password: "test",
		Database: "Ganzam",
	}
}

func GetDBConfig() DBConfig {
	cfg := getConfig(GetAppEnv())

	if value := strings.TrimSpace(os.Getenv("DB_SERVER")); value != "" {
		cfg.Server = value
	}
	if value := strings.TrimSpace(os.Getenv("DB_PORT")); value != "" {
		if port, err := strconv.Atoi(value); err == nil {
			cfg.Port = port
		}
	}
	if value := strings.TrimSpace(os.Getenv("DB_USER")); value != "" {
		cfg.User = value
	}
	if value := os.Getenv("DB_PASSWORD"); value != "" {
		cfg.Password = value
	}
	if value := strings.TrimSpace(os.Getenv("DB_NAME")); value != "" {
		cfg.Database = value
	}

	return cfg
}

func GetDBConnectionString() string {
	cfg := GetDBConfig()
	return fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s",
		cfg.User,
		cfg.Password,
		cfg.Server,
		cfg.Port,
		cfg.Database,
	)
}

func GetDBTargetSummary() string {
	cfg := GetDBConfig()
	return fmt.Sprintf("env=%s server=%s port=%d database=%s user=%s", GetAppEnv(), cfg.Server, cfg.Port, cfg.Database, cfg.User)
}
