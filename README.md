# MQTT4Rclone

Control Rclone by MQTT

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg?style=flat-square)](https://github.com/mqtt-smarthome/mqtt-smarthome)
[![Build/Test](https://github.com/Legobas/mqtt4rclone/actions/workflows/release.yml/badge.svg)](https://github.com/Legobas/mqtt4rclone/actions/workflows/release.yml)
[![CodeQL](https://github.com/Legobas/mqtt4rclone/actions/workflows/codeql.yml/badge.svg)](https://github.com/Legobas/mqtt4rclone/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt4rclone)](https://goreportcard.com/report/github.com/legobas/mqtt4rclone)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt4rclone?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt4rclone)
[![Docker Image Size](https://badgen.net/docker/size/legobas/mqtt4rclone?icon=docker&label=image%20size)](https://hub.docker.com/r/legobas/mqtt4rclone)

MQTT4Rclone is a service that enables [Rclone](https://rclone.org) to be controlled and monitored using MQTT commands.
It makes it possible to use Rclone with MQTT-enabled devices or services to create automated workflows, such as triggering backups based on specific events or schedules.
Because it is based on the MQTT protocol can it be easily be integrated with home automation systems, like Home Assistant, Domoticz or Node-RED.

## Installation

MQTT4Rclone can be used in a [Go](https://go.dev) environment or as a [Docker container](#docker):

```bash
$ go get -u github.com/Legobas/mqtt4rclone
```

## Environment variables

Supported environment variables:

```
LOGLEVEL = INFO/DEBUG/ERROR
```

# Configuration

MQTT4Rclone can be configured with the `mqtt4rclone.yml` yaml configuration file.
The `mqtt4rclone.yml` file has to exist in one of the following locations:

 * A `config` directory in de filesystem root: `/config/mqtt4rclone.yml`
 * A `.config` directory in the user home directory `~/.config/mqtt4rclone.yml`
 * The current working directory

## Configuration options

| Config item               | Description                                                              |
| ------------------------- | ------------------------------------------------------------------------ |
| **mqtt**                  |                                                                          |
| url                       | MQTT Server URL                                                          |
| username/password         | MQTT Server Credentials                                                  |
| qos                       | MQTT Server Quality Of Service                                           |
| **rclone**                |                                                                          |
| response_topic            | MQTT Topic for Rclone response message                                   |

Example mqtt4rclone.yml:

```yml
    mqtt:
      url: "tcp://<MQTT SERVER>:1883"
      username: <MQTT USERNAME>
      password: <MQTT PASSWORD>
      qos: 0

rclone:
  response_topic: <MQTT Topic>

```      

## Commands

MQTT4Rclone send commands to Rclone based on the MQTT topic and message.
The topic is the same as the Rclone rc url path and the message is the JSON as described by: 
[Rclone Commands](https://rclone.org/rc/#supported-commands)

All the `sync/*` commands **will be started as jobs**, MQTT4Rclone will add `"_async":true` to the json message.

The local path in the docker container is `/data`.
This path has to be used in the commands sent to rclone (see examples).

## Examples

```
Order:
MQTT topic
MQTT message

mqtt4rclone/options/get
<empty message> or {}

mqtt4rclone/options/set
{"main":{"LogLevel":"DEBUG"}}

mqtt4rclone/config/listremotes
{}

mqtt4rclone/operations/fsinfo
{"fs":"dropbox:"}

mqtt4rclone/sync/sync
{"srcFs":"/data/mydropbox","dstFs":"dropbox:","_filter":{"MaxAge":"1d"}}

```

## Rclone Response

The default MQTT topic where the Rclone response is sent to is: `mqtt4rclone/response`.
This topic can be changed by the setting configuration option `response_topic`,
so any MQTT client, like Node-RED, can process the Rclone response for other actions.

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

Docker run example:

```bash
$ docker run -d -v /home/legobas/mqtt4rclone:/config -v /home/legobas/rclone_data:/data legobas/mqtt4rclone
```

Docker compose example:

```yml
services:
  MQTT4Rclone:
    image: legobas/mqtt4rclone:latest
    container_name: mqtt4rclone
    environment:
      - PUID=1000              # User id for access to config or data directories with user rights
      - PGID=1000              # User group id
      - LOGLEVEL=debug         # MQTT4Rclone log level: ERROR/INFO (default)/DEBUG/TRACE
      - RCLONE_LOGLEVEL=INFO   # Rclone log level: ERROR/NOTICE (default)/INFO/DEBUG
      - TZ=America/New_York    # Timezone
    volumes:
      - /home/legobas/mqtt4rclone:/config
      - /cloud_data:/data
    restart: unless-stopped
```

The environment variables are not necessary, if omitted the default values will be used.

## Logging

To temporarily set the rclone logging to debug, send the command:

```
mqtt4rclone/options/set
{"main":{"LogLevel":"DEBUG"}}
```

after debugging set it back by restarting the service or by sending:

```
mqtt4rclone/options/set
{"main":{"LogLevel":"NOTICE"}}
```

The logs of mqtt4rclone and rclone are written to stdout, with docker is this the standard docker log.

To view the logging:
`docker compose logs` or `docker logs mqtt4rclone`


## Credits

* [Rclone](https://rclone.org)
* [Paho Mqtt Client](https://github.com/eclipse/paho.mqtt.golang)
* [ZeroLog](https://github.com/rs/zerolog)
* [MultiRun](https://nicolas-van.github.io/multirun)
