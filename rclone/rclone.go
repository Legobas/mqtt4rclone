package rclone

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	baseUrl        = "http://rclone:5533"
	requestTimeout = 10 * time.Second
)

type RcloneSettings struct {
	Username string
	Password string
}

func New(username string, password string) (*RcloneSettings, error) {
	var err error

	log.Trace().Msgf("Username: %s", username)

	rc := &RcloneSettings{
		Username: username,
		Password: password,
	}

	if rc.RcloneOnline() {
		err = nil
	} else {
		err = errors.New("could not connect to Rclone")
	}

	return rc, err
}

func (rc RcloneSettings) SendToRclone(command string, jsonMsg string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	log.Debug().Msgf("jsonMsg: %s", jsonMsg)

	url := baseUrl + command
	async := false

	// jsonData, err := json.Marshal(jsonMsg)
	// if err != nil {
	// log.Error().Err(err).Msg("Failed to marshal request")
	// return "", err
	// }

	// start job for sync (add _async:true to json)
	// if strings.HasPrefix(command, "/sync") {
	// async = true
	// json_data := map[string]any{}
	// err := json.Unmarshal([]byte(jsonMsg), &json_data)
	// if err != nil {
	// log.Fatal().Err(err)
	// }
	// json_data["_async"] = true
	// jsonData, err = json.Marshal(jsonMsg)
	// if err != nil {
	// log.Error().Err(err).Msg("Failed to marshal request")
	// return "", err
	// }
	// }

	log.Debug().Msgf("Rclone url: %s", url)
	log.Debug().Msgf("Rclone json: %s", jsonMsg)

	payload := bytes.NewReader([]byte(jsonMsg))
	log.Trace().Msgf("Rclone Request:\n%v", jsonMsg)

	req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		log.Error().Err(err).Msg("HTTP Request error")
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	if rc.Username != "" {
		req.SetBasicAuth(rc.Username, rc.Password)
	}

	httpClient := &http.Client{Timeout: requestTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("HTTP Client error")
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal().Err(err)
		return "", err
	}

	responseMessage := string(response)

	log.Trace().Msgf("Rclone Response:\n%s", responseMessage)

	if async {
		log.Debug().Msgf("Job: %s", responseMessage)
	}

	return responseMessage, nil
}

func (rc RcloneSettings) RcloneOnline() bool {
	online := false
	for i := 1; i < 10; i++ {
		time.Sleep(time.Second * 1)
		response, err := rc.SendToRclone("/core/version", "{}")
		if err == nil {
			online = true
			data := map[string]any{}
			err := json.Unmarshal([]byte(response), &data)
			if err != nil {
				log.Fatal().Err(err)
			}
			log.Info().Msgf("Rclone version: %s", data["version"])
			break
		}
	}
	return online
}
