package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string     `yaml:"env" env-default:"local"`
	StoragePath string     `yaml:"storage_path" env-required:"true"`
	DBPath      string     `yaml:"database_path"`
	GRPC        GRPCConfig `yaml:"grpc"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty2: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()
	cfgpath := firstNonEmptyString(res, os.Getenv("CONFIG_PATH"), "../../internal/config/local.yaml")

	return cfgpath
}

func firstNonEmptyString(strings ...string) string {
	for _, str := range strings {
		if str != "" {
			return str
		}
	}

	return ""
}
