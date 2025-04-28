package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env     string `yaml:"env" env-default:"local"`
	LogFile `yaml:"logFile"`
	GRPC    `yaml:"grpc"`
	Storage StorageData `yaml:"storage"`
	Cert    Cert        `yaml:"cert"`
}

type LogFile struct {
	Use  bool   `yaml:"use" env-default:"false"`
	Name string `yaml:"name" env-default:"auth-roles-hub.log"`
}

type GRPC struct {
	Port    int           `yaml:"port" env-default:"50000"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type StorageData struct {
	User     string `yaml:"user"`
	Pass     string `yaml:"pass"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port" env-default:"5432"`
	Database string `yaml:"database"`
	Schema   string `yaml:"schema"`
}

type Cert struct {
	Jwt string `yaml:"jwt"`
}

var cfg *Config

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if len(configPath) == 0 {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file \"" + configPath + "\" does not exist")
	}

	cfg = new(Config)
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		panic("cant load config: " + err.Error())
	}

	return cfg
}

func fetchConfigPath() string {
	var path string
	flag.StringVar(&path, "config", "", "path to config file")
	flag.Parse()

	if len(path) == 0 {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
