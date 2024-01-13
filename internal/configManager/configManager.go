package configManager

import (
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/pelletier/go-toml/v2"
)

const (
	config_fp     = "./config.toml"
	latestVersion = "1.0.1"
)

type Config struct {
	Server struct {
		Host string
		Port int16
	}
	Download struct {
		Directory      string
		ConcurrentJobs int8
	}
	WebUi struct {
		UpdateInterval float64
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

	settings.WebUi.UpdateInterval = 5.0

	settings.Dev.PlaywrightDebug = false
	settings.Dev.ServerLogging = false

	settings.Version = latestVersion

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

	err = file.Close()
	if err != nil {
		return cfg, err
	}

	actual, _ := semver.NewVersion(cfg.Version)
	latest, _ := semver.NewVersion(latestVersion)

	// if saved config is older than the latest one, update it
	if actual.Compare(latest) == -1 {
		err = os.Remove(config_fp)
		if err != nil {
			return cfg, err
		}

		newConfig := initSettings()
		newConfig.Update(cfg)
		err = newConfig.Save()
		if err != nil {
			panic(err)
		}

		return newConfig, nil
	}

	return cfg, nil
}

func (cfg *Config) Update(oldcfg *Config) {
	cfg.Server.Host = oldcfg.Server.Host
	cfg.Server.Port = oldcfg.Server.Port

	cfg.Download.Directory = oldcfg.Download.Directory
	cfg.Download.ConcurrentJobs = oldcfg.Download.ConcurrentJobs

	cfg.Dev.PlaywrightDebug = oldcfg.Dev.PlaywrightDebug
	cfg.Dev.ServerLogging = oldcfg.Dev.ServerLogging
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
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}
