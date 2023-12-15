package configManager

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

const config_fp = "./config.toml"

type Config struct {
	Server struct {
		Host string
		Port int16
	}
	Download struct {
		Directory      string
		ConcurrentJobs int8
	}
	Dev struct {
		PlaywrightDebug bool
		ServerLogging   bool
	}
	Version string
}

func initSettings() *Config {
	settings := new(Config)

	settings.Server.Host = "127.0.0.1"
	settings.Server.Port = 5522

	settings.Download.Directory = "./Downloads"
	settings.Download.ConcurrentJobs = 2

	settings.Dev.PlaywrightDebug = false
	settings.Dev.ServerLogging = false

	settings.Version = "1.0"

	return settings
}

func NewConfig() (*Config, error) {
	cfg := &Config{}

	file, err := os.Open(config_fp)
	if err != nil {
		cfg = initSettings()
		_ = cfg.Save()

		return initSettings(), nil
	}

	decoder := toml.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) Save() error {
	file, err := os.Open(config_fp)
	if err != nil {
		file, err = os.Create(config_fp)
		if err != nil {
			return err
		}
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(initSettings())
	if err != nil {
		return err
	}

	return nil
}
