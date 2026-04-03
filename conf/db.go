package conf

import "fmt"

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
	return getConfig(GetAppEnv())
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
