package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	dataDir        = "data"
	topicBase      = "mqtt4rclone"
	configFilename = "mqtt4rclone.yml"
)

type Mqtt struct {
	Url       string `yaml:"url"`
	Username  string `yaml:"username"`   // default = ""
	Password  string `yaml:"password"`   // default = ""
	Qos       int    `yaml:"qos"`        // default = 0 (AtMostOnce)
	Retain    bool   `yaml:"retain"`     // default = false
	BaseTopic string `yaml:"base_topic"` // default = ""
}

type Rclone struct {
	//	Port          int    `yaml:"rc_port"`
	Username      string `yaml:"rc_username"`
	Password      string `yaml:"rc_password"`
	ResponseTopic string `yaml:"response_topic"`
}

type Config struct {
	Mqtt   Mqtt   `yaml:"mqtt"`
	Rclone Rclone `yaml:"rclone"`
}

func GetConfig() (Config, error) {
	var conf Config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return conf, errors.New("could not get Home Dir")
	}
	workDir, err := os.Getwd()
	if err != nil {
		return conf, errors.New("could not get Work Dir")
	}

	paths := []string{
		filepath.Join("/", dataDir, configFilename),
		filepath.Join(homeDir, dataDir, configFilename),
		filepath.Join(workDir, configFilename),
		filepath.Join(workDir, dataDir, configFilename),
	}

	var configPath string
	var dataDir string
	var msg string
	for _, path := range paths {
		msg += path + " "
		if data, err := os.ReadFile(path); err == nil {
			err = yaml.Unmarshal(data, &conf)
			if err != nil {
				return conf, fmt.Errorf("failed to unmarshal config file: %s: %w", path, err)
			}
			configPath = path
			dataDir = filepath.Dir(path)
			break
		}
	}

	log.Trace().Msgf("Config: %+v", conf)

	if len(configPath) == 0 {
		return conf, fmt.Errorf("configuration file not found in paths: %s", msg)
	}

	if len(dataDir) == 0 {
		return conf, errors.New("data dir is empty")
	}

	if err := conf.validateConfig(); err != nil {
		return conf, fmt.Errorf("invalid config from %s: %w", configPath, err)
	}

	if len(conf.Mqtt.BaseTopic) == 0 {
		conf.Mqtt.BaseTopic = topicBase
	}

	log.Debug().Msgf("Configuration loaded: %s", configPath)
	// Trace logging is disabled for security reasons
	log.Trace().Msgf("MQTT Config: %+v", conf.Mqtt)
	log.Trace().Msgf("Rclone Config: %+v", conf.Rclone)

	return conf, nil
}

func (cfg Config) validateConfig() error {
	if cfg.Mqtt.Url == "" {
		return errors.New("MQTT Server URL is mandatory")
	}
	if cfg.Rclone.ResponseTopic == "" {
		return errors.New("Rclone ResponseTopic is mandatory")
	}
	return nil
}
