package configManager

import (
	"encoding/json"
	"os"

	"github.com/pelletier/go-toml/v2"
)

const (
	config_fp     = "./config.toml"
	latestVersion = "0.4.0-b3"
)

type Config struct {
	Server struct {
		Host string
		Port uint16
		SSL  struct {
			Active bool
			Cert   string
			Key    string
		}
	}
	Download struct {
		ConcurrentJobs int8
		Directory      string
		Tempdir        string
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

	cfg.Server.Host = "auto"
	cfg.Server.Port = 5522

	cfg.Server.SSL.Active = false
	cfg.Server.SSL.Cert = ""
	cfg.Server.SSL.Key = ""

	cfg.Download.ConcurrentJobs = 2
	cfg.Download.Directory = "./Downloads"
	cfg.Download.Tempdir = "./Downloads/.tmp"

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
		*cfg = *updateCfg(cfg, NewConfig())
	}

	return nil
}

// Updates the old config by merging the old cfg's values
// into the latest cfg struct
//
// This is a bad hack tbh, but if it works it works
func updateCfg(old *Config, latest *Config) *Config {
	var oldCfg map[string]any

	data, _ := json.Marshal(old)
	json.Unmarshal(data, &oldCfg)

	serverCfg, ok := oldCfg["Server"].(map[string]any)
	if ok {
		_, ok = serverCfg["Host"]
		if ok {
			latest.Server.Host = old.Server.Host
		}

		_, ok = serverCfg["Port"]
		if ok {
			latest.Server.Port = old.Server.Port
		}

		sslCfg, ok := oldCfg["SSL"].(map[string]any)
		if ok {
			_, ok = sslCfg["Active"]
			if ok {
				latest.Server.SSL.Active = old.Server.SSL.Active
			}

			_, ok = sslCfg["Cert"]
			if ok {
				latest.Server.SSL.Cert = old.Server.SSL.Cert
			}

			_, ok = sslCfg["Key"]
			if ok {
				latest.Server.SSL.Key = old.Server.SSL.Key
			}
		}
	}

	downloadCfg, ok := oldCfg["Download"].(map[string]any)
	if ok {
		_, ok = downloadCfg["ConcurrentJobs"]
		if ok {
			latest.Download.ConcurrentJobs = old.Download.ConcurrentJobs
		}

		_, ok = downloadCfg["Directory"]
		if ok {
			latest.Download.Directory = old.Download.Directory
		}

		_, ok = downloadCfg["Tempdir"]
		if ok {
			latest.Download.Tempdir = old.Download.Tempdir
		}
	}

	devCfg, ok := oldCfg["Dev"].(map[string]any)
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

	latest.Version = latestVersion

	return latest
}

// Saves the config to file
func (cfg *Config) Save() error {
	file, err := os.OpenFile(config_fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	err = encoder.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}
