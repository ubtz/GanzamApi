package conf

import (
	"os"

	"github.com/joho/godotenv"
)

var Env string

func init() {
	_ = godotenv.Load(".env")
	Env = os.Getenv("APP_ENV")
	if Env == "" {
		Env = EnvTest
	}
}
