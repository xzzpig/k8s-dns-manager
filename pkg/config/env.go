package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"

	_ "embed"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Environment string `envconfig:"GO_ENV"`
	Default     struct {
		Record struct {
			TTL int `envconfig:"NATM_DEFAULT_RECORD_TTL"`
		}
		Generator struct {
			Type string `envconfig:"NATM_DEFAULT_GENERATOR_TYPE"`
			DDNS struct {
				Timeout         time.Duration `envconfig:"NATM_DEFAULT_GENERATOR_DDNS_TIMEOUT"`
				ExtraApis       []string      `envconfig:"NATM_DEFAULT_GENERATOR_DDNS_EXTRA_APIS"`
				CacheExpire     time.Duration `envconfig:"NATM_DEFAULT_GENERATOR_DDNS_CACHE_EXPIRE"`
				CleanInterval   time.Duration `envconfig:"NATM_DEFAULT_GENERATOR_DDNS_CLEAN_INTERVAL"`
				RefreshInternal time.Duration `envconfig:"NATM_DEFAULT_GENERATOR_DDNS_REFRESH_INTERNAL"`
			}
			CName struct {
				Value string `envconfig:"NATM_DEFAULT_GENERATOR_CNAME_VALUE"`
			}
		}
	}
	Bind struct {
		Metrics     string `envconfig:"NATM_BIND_METRICS"`
		HealthProbe string `envconfig:"NATM_BIND_HEALTH_PROBE"`
	}
}

var config Config

type EnvKey = string

const (
	GO_ENV EnvKey = "GO_ENV"
)

//go:embed .env
var defaultEnv string

func init() {
	env := os.Getenv(GO_ENV)
	if env == "" {
		env = "development"
	}

	godotenv.Load(".env." + env + ".local")
	godotenv.Load(".env." + env)
	godotenv.Load()
	envMap, err := godotenv.Unmarshal(defaultEnv)
	if err != nil {
		panic(err)
	}
	for k, v := range envMap {
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}

	err = envconfig.Process("dnsm", &config)
	if err != nil {
		panic(err)
	}
}

func GetConfig() *Config {
	return &config
}
