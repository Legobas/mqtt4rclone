package main

import (
	"os"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	TIMEOUT        time.Duration = time.Second * 10
	SUBSCRIBE                    = APPNAME + "/#"
	STATUS_TOPIC                 = APPNAME + "/status"
	RESPONSE_TOPIC               = APPNAME + "/response"
)

var mqttClient MQTT.Client

func sendToMtt(topic string, message string, retain bool) {
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), retain, message)
}

func receive(client MQTT.Client, msg MQTT.Message) {
	topic := msg.Topic()
	responseTopic := RESPONSE_TOPIC
	if config.Rclone.ResponseTopic != "" {
		responseTopic = config.Rclone.ResponseTopic
	}

	if topic != STATUS_TOPIC && topic != responseTopic {
		message := string(msg.Payload()[:])
		log.Trace().Msgf("MQTT Topic: %s", topic)
		log.Trace().Msgf("MQTT Message: %s", message)
		log.Trace().Msgf("MQTT Response Topic: %s", responseTopic)

		command := topic[len(APPNAME):]
		json := strings.TrimSpace(message)
		if len(json) == 0 {
			json = "{}"
		}

		response, err := sendToRclone(command, json)
		if err != nil {
			log.Fatal().Err(err)
			return
		}

		sendToMtt(responseTopic, response, false)
	}
}

func GetClientId() string {
	hostname, _ := os.Hostname()
	return APPNAME + "_" + hostname
}

func startMqttClient() {
	opts := MQTT.NewClientOptions().AddBroker(config.Mqtt.Url)
	if config.Mqtt.Username != "" && config.Mqtt.Password != "" {
		opts.SetUsername(config.Mqtt.Username)
		opts.SetPassword(config.Mqtt.Password)
	}
	opts.SetClientID(GetClientId())
	opts.SetCleanSession(true)
	opts.SetBinaryWill(STATUS_TOPIC, []byte("Offline"), 0, true)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(connLostHandler)
	opts.SetOnConnectHandler(onConnectHandler)

	mqttClient = MQTT.NewClient(opts)
	token := mqttClient.Connect()
	if token.WaitTimeout(TIMEOUT) && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msg("MQTT connection")
	}

	token = mqttClient.Publish(STATUS_TOPIC, 2, true, "Online")
	token.Wait()
}

func connLostHandler(c MQTT.Client, err error) {
	log.Fatal().Err(err).Msg("MQTT connection lost")
}

func onConnectHandler(c MQTT.Client) {
	log.Debug().Msg("MQTT Client connected")
	token := mqttClient.Subscribe(SUBSCRIBE, 0, receive)
	if token.Wait() && token.Error() != nil {
		log.Fatal().Err(token.Error()).Msgf("Could not subscribe to %s", SUBSCRIBE)
	}
}
