package configManager

import (
	"encoding/json"
	"os"

	"github.com/pelletier/go-toml/v2"
)

const (
	config_fp     = "./config.toml"
	latestVersion = "0.3.0"
)

type Config struct {
	Server struct {
		Host string
		Port uint16
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

/*
Creates a new config.

This function DOES NOT save it to file.
To save it, call cfg.Save()
*/
func NewConfig() *Config {
	cfg := &Config{}

	cfg.Server.Host = "127.0.0.1"
	cfg.Server.Port = 5522

	cfg.Download.Directory = "./Downloads"
	cfg.Download.ConcurrentJobs = 2

	cfg.Dev.PlaywrightDebug = false
	cfg.Dev.ServerLogging = false

	cfg.Version = latestVersion

	return cfg
}

/*
Loads the config from file

This function DOES NOT save it to file.
To save it, call cfg.Save()
*/
func (cfg *Config) Load() error {
	file, err := os.Open(config_fp)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := toml.NewDecoder(file)
	err = decoder.Decode(cfg)
	if err != nil {
		return err
	}

	if cfg.Version != latestVersion {
		updateCfg(cfg, NewConfig())
	}

	return nil
}

// Updates the old config by merging the old cfg's values
// into the latest cfg struct
//
// This is a bad hack tbh, but if it works it works
func updateCfg(old *Config, latest *Config) {
	var oldCfg map[string]interface{}

	data, _ := json.Marshal(old)
	json.Unmarshal(data, &oldCfg)

	serverCfg, ok := oldCfg["Server"].(map[string]interface{})
	if ok {
		_, ok = serverCfg["Host"]
		if ok {
			latest.Server.Host = old.Server.Host
		}

		_, ok = serverCfg["Port"]
		if ok {
			latest.Server.Port = old.Server.Port
		}
	}

	downloadCfg, ok := oldCfg["Download"].(map[string]interface{})
	if ok {
		_, ok = downloadCfg["Directory"]
		if ok {
			latest.Download.Directory = old.Download.Directory
		}

		_, ok = downloadCfg["ConcurrentJobs"]
		if ok {
			latest.Download.ConcurrentJobs = old.Download.ConcurrentJobs
		}
	}

	devCfg, ok := oldCfg["Dev"].(map[string]interface{})
	if ok {
		_, ok = devCfg["PlaywrightDebug"]
		if ok {
			latest.Dev.PlaywrightDebug = old.Dev.PlaywrightDebug
		}

		_, ok = devCfg["ServerLogging"]
		if ok {
			latest.Dev.ServerLogging = old.Dev.ServerLogging
		}
	}
}

// Saves the config to file
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
