# MQTT4Rclone

Control Rclone with MQTT messages

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg?style=flat-square)](https://github.com/mqtt-smarthome/mqtt-smarthome)
[![Build/Test](https://github.com/Legobas/mqtt4rclone/actions/workflows/release.yml/badge.svg)](https://github.com/Legobas/mqtt4rclone/actions/workflows/release.yml)
[![CodeQL](https://github.com/Legobas/mqtt4rclone/actions/workflows/codeql.yml/badge.svg)](https://github.com/Legobas/mqtt4rclone/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt4rclone)](https://goreportcard.com/report/github.com/legobas/mqtt4rclone)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt4rclone?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt4rclone)
[![Docker Image Size](https://badgen.net/docker/size/legobas/mqtt4rclone?icon=docker&label=image%20size)](https://hub.docker.com/r/legobas/mqtt4rclone)

MQTT4Rclone is a service that enables [Rclone](https://rclone.org) to be controlled and monitored using MQTT commands.
It makes it possible to use Rclone with MQTT-enabled devices or services to create automated workflows, such as triggering backups based on specific events or schedules.
Because it is based on the MQTT protocol it can be easily integrated with home automation systems, like Home Assistant, Domoticz or Node-RED.
A use case is to upload the images or videos captured by a security camera to cloud-based storage services.

## How it works

MQTT4Rclone runs **next to the official [rclone/rclone](https://hub.docker.com/r/rclone/rclone) Docker image**, which is started in Remote Control mode so it can be controlled by its HTTP API.
MQTT4Rclone only communicates with the Rclone API over HTTP; it never touches your files itself.

Since MQTT4Rclone talks to the official image, it always works with the latest Rclone releases — just pull `rclone/rclone:latest` and you are never left behind.

The mqtt4rclone service sends the MQTT JSON message it receives to the Rclone API, using the MQTT topic to create the Rclone URL.
Rclone will then perform the action, like synchronizing with a cloud storage.
Actions which take a long time, like sync or copy, are handled as a job and return an HTTP response immediately.
After the HTTP call is completed, the response message is sent back to an MQTT topic.
This is safely done using only the local network environment of the containers.

## Installation

MQTT4Rclone is available as a Docker container on [DockerHub](https://hub.docker.com/r/legobas/mqtt4rclone).

Two containers are deployed together:

1. **rclone** — the official `rclone/rclone` image, running in Remote Control mode
2. **mqtt4rclone** — the MQTT bridge that talks to the Rclone API

```bash
docker compose up -d
```

Updating to a new Rclone version is now as simple as:

```bash
docker compose pull && docker compose up -d
```

## Configuration

MQTT4Rclone can be configured with the `mqtt4rclone.yml` yaml configuration file.

Keep the RC API private; only publish port 5533 if you also want to access the Rclone Web GUI from outside the Docker network.

### MQTT4Rclone Configuration options

| Config item               | Description                              | Mandatory |
| ------------------------- | ---------------------------------------- | --------- |
| **mqtt**                  |                                          |           |
| url                       | MQTT Server URL                          | Yes       |
| username/password         | MQTT Server Credentials                  | No        |
| qos                       | MQTT Server Quality Of Service           | No        |
| **rclone**                |                                          |           |
| rc_username/rc_password   | RClone Remote Control API Credentials    | No        |
| response_topic            | MQTT Topic for Rclone response message   | No        |

Example mqtt4rclone.yml:

```yml
mqtt:
  url: "tcp://<MQTT SERVER>:1883"
  username: <MQTT USERNAME>
  password: <MQTT PASSWORD>
  qos: 0

rclone:
  rc_username: <Rclone Remote Control USERNAME>
  rc_password: <Rclone Remote Control PASSWORD>
  response_topic: <MQTT Topic>
```

### Rclone configuration

The Rclone configuration file must be present in the config directory of the rclone container (`/config/rclone`).
Make sure the credentials match those passed to rclone via `--user`/`--pass`, so MQTT4Rclone can authenticate against the Remote Control API.

## Commands

MQTT4Rclone sends commands to Rclone based on the MQTT topic and message.
The topic is the same as the Rclone rc URL path and the message is the JSON as described by:
[Rclone Commands](https://rclone.org/rc/#supported-commands)

All the `sync/*` commands **will be started as jobs**, MQTT4Rclone will add `"_async":true` to the json message.

The local path inside the **rclone** container is `/data`.
This path has to be used in the commands sent to rclone (see examples). MQTT4Rclone itself does not need access to this data; it only passes the paths on to the Rclone API.

### Examples

```
Order:
MQTT topic
MQTT message

mqtt4rclone/core/stats
<empty message> or {}

mqtt4rclone/options/get
<empty message> or {}

mqtt4rclone/options/set
{"main":{"LogLevel":"DEBUG"}}

mqtt4rclone/config/listremotes
{}

mqtt4rclone/operations/fsinfo
{"fs":"dropbox:"}

mqtt4rclone/sync/sync
{"srcFs":"/data/mydropbox","dstFs":"dropbox:"}

mqtt4rclone/sync/copy
{"srcFs":"/data/mydropbox","dstFs":"dropbox:","_filter":{"MaxAge":"1d"}}
```

## Rclone Response

The default MQTT topic where the Rclone response is sent to is: `mqtt4rclone/response`.
This topic can be changed by the configuration option `response_topic`,
so any MQTT client can receive the Rclone response and process it.

## Autosync

To automatically sync every day/hour/minute you can use [MQTT-Timer](https://github.com/Legobas/mqtt-timer) with a configuration like this:

```yml
    timers:
    - id: rclone_dropbox
      time: 01:00:00
      description: RClone Sync Local to Dropbox every day at 1.00
      topic: mqtt4rclone/sync/sync
      message: '{"srcFs":"/data/mydropbox","dstFs":"dropbox:"}'
    - id: rclone_mega
      cron: 00 * * * *
      description: RClone Mega to Local every hour
      topic: mqtt4rclone/sync/copy
      message: '{"srcFs":"mega:","dstFs":"/data/mymega"}'
```

## Docker

Docker compose example, running MQTT4Rclone next to the official rclone image:

```yml
services:
  rclone:
    image: rclone/rclone:latest
    container_name: rclone
    ports:
      - 5533:5533
    volumes:
      - /home/legobas/data/rclone:/config/rclone
      - /storage/rclone:/data
    command:
      - --api-addr=0.0.0.0:5533
      - --user=xxxxxxxx
      - --pass=xxxxxxxx
    restart: unless-stopped

  mqtt4rclone:
    image: legobas/mqtt4rclone:latest
    container_name: mqtt4rclone
    environment:
      - PUID=1000              # User id for access to config directory with user rights
      - PGID=1000              # User group id
      - LOGLEVEL=debug         # MQTT4Rclone log level: ERROR/INFO (default)/DEBUG/TRACE
      - RCLONE_LOGLEVEL=INFO   # Log level for rclone API responses
      - TZ=America/New_York    # Timezone
    volumes:
      - /home/legobas/mqtt4rclone:/config
    restart: unless-stopped
```

The environment variables are optional; if omitted, the default values will be used.

Only the rclone container needs a data volume: MQTT4Rclone communicates solely with the Rclone API, so all `/data` paths in the MQTT commands are resolved inside the rclone container.

## Logging

Rclone log levels: [DEBUG|INFO|NOTICE|ERROR (default NOTICE)](https://rclone.org/flags/#logging).
To temporarily set the rclone logging to debug, send the command:

```
mqtt4rclone/options/set
{"main":{"LogLevel":"DEBUG"}}
```

After debugging, set it back by restarting the service or by sending:

```
mqtt4rclone/options/set
{"main":{"LogLevel":"NOTICE"}}
```

The logs of mqtt4rclone and rclone are written to stdout, which is the standard docker log.

To view the logging:
`docker compose logs` or `docker logs mqtt4rclone`

## Credits

* [Rclone](https://rclone.org)
* [Paho Mqtt Client](https://github.com/eclipse/paho.mqtt.golang)
* [ZeroLog](https://github.com/rs/zerolog)
