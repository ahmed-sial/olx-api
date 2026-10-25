package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type HttpServer struct {
	Address string `env:"ADDRESS" yaml:"address" env-required:"true"`
	Port    string `env:"PORT"    yaml:"port"    env-required:"true"`
}

type AppConfig struct {
	Env        string     `env:"ENV"         yaml:"env"         env-required:"true"`
	HttpServer HttpServer `env:"HTTP_SERVER" yaml:"http_server" env-required:"true"`
}

func MustLoad() *AppConfig {
	var configPath string

	configPath = os.Getenv("CONFIG_PATH")

	if configPath == "" {
		flags := flag.String("config_path", "", "path to configuration file for app")
		flag.Parse()
		configPath = *flags
		if configPath == "" {
			log.Fatal("config path is not set")
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist on: %s", configPath)
	}

	var config AppConfig
	err := cleanenv.ReadConfig(configPath, &config)
	if err != nil {
		log.Fatal("failed to read configuration file:", err.Error())
	}

	return &config
}
