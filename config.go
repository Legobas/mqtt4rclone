package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	CONFIG_FILE = "mqtt4rclone.yml"
	CONFIG_DIR  = ".config"
	CONFIG_ROOT = "/config"
)

type Mqtt struct {
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Qos      int    `yaml:"qos"`
}

type Rclone struct {
	ResponseTopic string `yaml:"response_topic"`
}

type Config struct {
	Mqtt   Mqtt   `yaml:"mqtt"`
	Rclone Rclone `yaml:"rclone"`
}

func getConfig() Config {
	var config Config

	configFile := filepath.Join(CONFIG_ROOT, CONFIG_FILE)
	msg := configFile
	data, err := os.ReadFile(configFile)
	if err != nil {
		homedir, _ := os.UserHomeDir()
		configFile := filepath.Join(homedir, CONFIG_DIR, CONFIG_FILE)
		msg += ", " + configFile
		data, err = os.ReadFile(configFile)
	}
	if err != nil {
		workingdir, _ := os.Getwd()
		configFile := filepath.Join(workingdir, CONFIG_FILE)
		msg += ", " + configFile
		data, err = os.ReadFile(configFile)
	}
	if err != nil {
		msg = "Configuration file could not be found: " + msg
		log.Fatal().Msg(msg)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal().Err(err).Msg("unmarshal")
	}

	err = validate(config)
	if err != nil {
		log.Fatal().Err(err).Msg("validate")
	}

	log.Trace().Msgf("Config: %+v\n", config)
	return config
}

func validate(config Config) error {
	if config.Mqtt.Url == "" {
		return errors.New("Config error: MQTT Server URL is mandatory")
	}

	return nil
}
