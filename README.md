# MQTT4Rclone

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

Example mqtt4rclone.yml:

```yml
    mqtt:
      url: "tcp://<MQTT SERVER>:1883"
      username: <MQTT USERNAME>
      password: <MQTT PASSWORD>
      qos: 0
```      

## Commands

MQTT4Rclone send commands to Rclone based on the MQTT topic and message.
The topic is the same as the Rclone rc url path and the message is the JSON as described by: 
[Rclone Commands](https://rclone.org/rc/#supported-commands)

## Examples

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
      - PUID=1000
      - PGID=1000
      - LOGLEVEL=debug
      - RCLONE_LOGLEVEL=INFO
      - TZ=Europe/Amsterdam
    volumes:
      - /home/legobas/mqtt4rclone:/config
      - /home/legobas/rclone_data:/data
    restart: unless-stopped
```

## Credits

* [Rclone](https://rclone.org)
* [Paho Mqtt Client](https://github.com/eclipse/paho.mqtt.golang)
* [ZeroLog](https://github.com/rs/zerolog)
