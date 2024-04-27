package main

import (
	"os"
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

func sendToMtt(topic string, message string) {
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), config.Mqtt.Retain, message)
}

func sendToMttRetain(topic string, message string) {
	mqttClient.Publish(topic, byte(config.Mqtt.Qos), true, message)
}

func receive(client MQTT.Client, msg MQTT.Message) {
	topic := msg.Topic()
	log.Trace().Msgf("MQTT Topic: %s", topic)
	log.Trace().Msgf("MQTT Message: %s", msg)
	if topic != STATUS_TOPIC && topic != RESPONSE_TOPIC {
		command := topic[len(APPNAME):]
		json := string(msg.Payload()[:])

		response, err := sendToRclone(command, json)
		if err != nil {
			log.Fatal().Err(err)
			return
		}
	
		sendToMtt(RESPONSE_TOPIC, response)
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
