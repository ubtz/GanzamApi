package conf

import "os"

var Env string

func init() {
	Env = os.Getenv("APP_ENV")
	if Env == "" {
		Env = EnvTest
	}
}
