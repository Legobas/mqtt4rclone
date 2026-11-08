package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	appName string = "mqtt4rclone"
)

var (
	//Version import - build with -ldflags "-X main.Version=${VERSION}"
	Version string = "dev"
	config  Config
)

func init() {
	// Setup logging
	out := zerolog.NewConsoleWriter()
	out.NoColor = true
	out.FormatLevel = func(i interface{}) string {
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
		log.Logger = log.Level(zerolog.InfoLevel)
	}

	// Get Config
	config = getConfig()

	// Print Application with version
	log.Info().Msgf("%s %s", appName, Version)
}

func main() {
	zoneName, _ := time.Now().Zone()
	log.Debug().Msgf("%s start, Local Time=%s Timezone=%s", appName, time.Now().Local().Format("15:04:05"), zoneName)

	startMqttClient()

	// Refresh status
	sendToMtt(appName+"/status", "Online", true)

	if rcloneOnline() {
		log.Debug().Msg("Rclone is ready")
	} else {
		log.Warn().Msg("Rclone could not be found!")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	log.Debug().Msgf("%s stop, Local Time=%s Timezone=%s", appName, time.Now().Local().Format("15:04:05"), zoneName)
}
