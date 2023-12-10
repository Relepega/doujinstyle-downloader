package settingsmanager

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Settings struct {
	Server struct {
		Host string `yaml:"host"`
	}
}

func NewSettings() {
}

func openFile() error {
	f, err := os.Open("config.yml")
	if err != nil {
		return err
	}
	defer f.Close()

	var cfg Settings
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}
