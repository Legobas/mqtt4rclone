package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Legobas/mqtt4rclone/config"
	"github.com/Legobas/mqtt4rclone/mqtt"
	"github.com/Legobas/mqtt4rclone/rclone"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	TIMEOUT        time.Duration = time.Second * 10
	SUBSCRIBE                    = appName + "/#"
	STATUS_TOPIC                 = appName + "/status"
	RESPONSE_TOPIC               = appName + "/response"
)
const (
	appName string = "mqtt4rclone"
)

var (
	//Version import - build with -ldflags "-X main.Version=${VERSION}"
	Version  string = "dev"
	conf     config.Config
	mqttConn *mqtt.MqttConnection
	rc       *rclone.RcloneSettings
)

func init() {
	// Setup logging
	out := zerolog.NewConsoleWriter()
	out.NoColor = true
	out.FormatLevel = func(i any) string {
		return strings.ToUpper(fmt.Sprintf("%-6s", i))
	}
	out.PartsExclude = []string{zerolog.TimestampFieldName, zerolog.CallerFieldName}
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.Output(out)

	switch strings.ToLower(os.Getenv("LOGLEVEL")) {
	case "debug":
		log.Logger = log.Level(zerolog.DebugLevel)
	case "trace":
		log.Logger = log.Level(zerolog.TraceLevel)
	default:
		log.Logger = log.Level(zerolog.TraceLevel)
		// log.Logger = log.Level(zerolog.InfoLevel)
	}

	// Print Application with version
	log.Info().Msgf("%s %s", appName, Version)
}

func receive(topic string, message string) {
	responseTopic := RESPONSE_TOPIC
	if conf.Rclone.ResponseTopic != "" {
		responseTopic = conf.Rclone.ResponseTopic
	}

	if topic != STATUS_TOPIC && topic != responseTopic {
		log.Trace().Msgf("MQTT Topic: %s", topic)
		log.Trace().Msgf("MQTT Message: %s", message)
		log.Trace().Msgf("MQTT Response Topic: %s", responseTopic)

		command := topic[len(appName):]
		json := strings.TrimSpace(message)
		if len(json) == 0 {
			json = "{}"
		}

		if rc != nil {
			response, err := rc.SendToRclone(command, json)
			if err != nil {
				log.Fatal().Err(err)
				return
			}

			mqttConn.Publish(responseTopic, response)
		}
	}
}

func main() {
	var err error

	conf, err = config.GetConfig()
	if err != nil {
		log.Fatal().Err(err)
	}

	log.Debug().Msgf("Config is %v", conf)

	rc, err = rclone.New(conf.Rclone.Username, conf.Rclone.Password)
	if err != nil {
		log.Fatal().Err(err)
	}

	if rc.RcloneOnline() {
		log.Debug().Msg("Rclone is ready")
	} else {
		log.Warn().Msg("Rclone could not be found!")
	}

	var mqttProperties mqtt.MqttProperties
	mqttProperties.ClientId = fmt.Sprintf("%s_%s", appName, fmt.Sprintf("%d", time.Now().UnixMicro()))
	mqttProperties.Url = conf.Mqtt.Url
	mqttProperties.User = conf.Mqtt.Username
	mqttProperties.Password = conf.Mqtt.Password
	mqttProperties.Qos = byte(conf.Mqtt.Qos)
	mqttProperties.Retain = conf.Mqtt.Retain
	mqttProperties.BaseTopic = appName

	mqttConn, err = mqtt.NewMqttConnection(mqttProperties)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize MQTT")
	}

	mqttConn.Subscribe(appName+"/#", receive)

	zoneName, _ := time.Now().Zone()
	log.Debug().Msgf("%s start, Local Time=%s Timezone=%s", appName, time.Now().Local().Format("15:04:05"), zoneName)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	mqttConn.Close()
	log.Info().Msgf("%s stopped", strings.ToUpper(appName))
}
