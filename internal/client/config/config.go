package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ClientConfig struct {
	QueryTime   time.Duration `yaml:"query_timeout" env-default:"2s"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	GRPCAddress string        `yaml:"grpc_address" env-required:"true"`
	WSURL       string        `yaml:"ws_url" env-required:"true"`
	CaCertFile  string        `yaml:"ca_cert_file" env=required:"true"`
}

func MustLoad() *ClientConfig {
	path := configPath()
	if path == "" {
		fmt.Println("empty config path")
		os.Exit(1)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists: " + path)
	}

	var cfg ClientConfig

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func configPath() string {
	var res string

	flag.StringVar(&res, "c", "./config/client_config.yaml", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CLIENT_CONFIG_PATH")
	}

	return res
}
