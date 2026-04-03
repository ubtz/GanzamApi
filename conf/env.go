package conf

import "os"

const (
	EnvTest = "test"
	EnvProd = "prod"
)

func GetAppEnv() string {
	switch os.Getenv("APP_ENV") {
	case EnvProd:
		return EnvProd
	default:
		return EnvTest
	}
}

func GetTargetURL() string {
	if GetAppEnv() == EnvProd {
		if value := os.Getenv("PROD_API_URL"); value != "" {
			return value
		}
		return "https://prod.example.com"
	}

	if value := os.Getenv("TEST_API_URL"); value != "" {
		return value
	}
	return "https://test.example.com"
}
