package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	BASE_URL                 = "http://localhost:5572"
	RC_USER                  = "cmd"
	RC_PASS                  = "f6Lhi09wfbxkd8Ok2l4H"
	RC_TIMEOUT time.Duration = time.Second * 10
)

func sendToRclone(command string, payload string) (string, error) {
	client := http.Client{Timeout: RC_TIMEOUT}

	url := BASE_URL + command
	async := false

	// start job for sync (add _async:true to json)
	if strings.HasPrefix(command, "/sync") {
		async = true
		json_data := map[string]interface{}{}
		err := json.Unmarshal([]byte(payload), &json_data)
		if err != nil {
			log.Fatal().Err(err)
		}
		json_data["_async"] = true
		msg, _ := json.Marshal(json_data)
		payload = string(msg)
	}

	log.Debug().Msgf("Rclone url: %s", url)
	log.Debug().Msgf("Rclone json: %s", payload)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		log.Fatal().Err(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(RC_USER, RC_PASS)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal().Err(err)
		return "", err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal().Err(err)
		return "", err
	}

	responseMessage := string(response)

	log.Trace().Msg(responseMessage)

	if async {
		log.Debug().Msgf("Job: %s", responseMessage)
	}

	return responseMessage, nil
}

func rcloneOnline() bool {
	online := false
	for i := 1; i < 10; i++ {
		response, err := sendToRclone("/core/version", "{}")
		if err != nil {
			time.Sleep(time.Second * 1)
		} else {
			online = true
			data := map[string]interface{}{}
			err := json.Unmarshal([]byte(response), &data)
			if err != nil {
				log.Fatal().Err(err)
			}
			log.Info().Msgf("RClone version: %s", data["version"])
			break
		}
	}
	return online
}
