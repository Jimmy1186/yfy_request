package config

import (
	"log"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	CONVEYOR_LOCATION_NAME      string `yaml:"CONVEYOR_LOCATION_NAME"`
	CONVEYOR_LOCATION_ID        string `yaml:"CONVEYOR_LOCATION_ID"`
	DEFAULT_CONTAINER_FORMAT_ID string `yaml:"DEFAULT_CONTAINER_FORMAT_ID"`
	DEFAULT_TRAY_FORMAT_ID      string `yaml:"DEFAULT_TRAY_FORMAT_ID"`
}

var Cfg Config

func init() {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Fatal("Cannot read config.yaml:", err)
	}

	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		log.Fatal("YAML parse error:", err)
	}
}
