package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DatabaseURL    string        `yaml:"database_url" env-required:"true"`
	TokenTTL       time.Duration `yaml:"token_ttl" env-default:"1h"`
	ConnectTimeout time.Duration `yaml:"connect_timeout" env-default:"2s"`
	CertFile       string        `yaml:"cert_file" env-required:"true"`
	KeyFile        string        `yaml:"key_file" env-required:"true"`
	GRPC           GRPCConfig    `yaml:"grpc"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	path := configPath()
	if path == "" {
		panic("config path is empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func configPath() string {
	var res string
	flag.StringVar(&res, "config", "./config/auth_config.yaml", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("AUTH_CONFIG_PATH")
	}
	return res
}
