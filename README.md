# MQTT4Rclone

**Control and monitor Rclone instances via MQTT messages.**

[![mqtt-smarthome](https://img.shields.io/badge/mqtt-smarthome-blue.svg?style=flat-square)](https://github.com/mqtt-smarthome/mqtt-smarthome)
[![Build/Test](https://github.com/Legobas/mqtt4rclone/actions/workflows/release.yml/badge.svg)](https://github.com/Legobas/mqtt4rclone/actions/workflows/release.yml)
[![CodeQL](https://github.com/Legobas/mqtt4rclone/actions/workflows/codeql.yml/badge.svg)](https://github.com/Legobas/mqtt4rclone/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Legobas/mqtt4rclone)](https://goreportcard.com/report/github.com/legobas/mqtt4rclone)
[![Docker Pulls](https://badgen.net/docker/pulls/legobas/mqtt4rclone?icon=docker&label=pulls)](https://hub.docker.com/r/legobas/mqtt4rclone)
[![Docker Image Size](https://badgen.net/docker/size/legobas/mqtt4rclone?icon=docker&label=image%20size)](https://hub.docker.com/r/legobas/mqtt4rclone)

MQTT4Rclone is a lightweight bridge service that enables [Rclone](https://rclone.org) to be controlled and monitored through the MQTT protocol. By decoupling file synchronization from the command line, it allows you to trigger backups and manage cloud storage based on external events or schedules.

Built on standard MQTT conventions, it integrates seamlessly into smart home and automation platforms such as **Home Assistant**, **Domoticz**, and **Node-RED**—making it ideal for automated workflows like uploading security camera footage to cloud storage upon motion detection.

---

## Architecture & How It Works

MQTT4Rclone operates as a companion container alongside the official [`rclone/rclone`](https://docker.com) Docker image running in Remote Control (RC) mode.

* **Decoupled Design:** MQTT4Rclone communicates strictly with the Rclone HTTP API over the internal Docker network. It never accesses your local filesystem or cloud credentials directly.
* **Seamless Upgrades:** Because it interfaces with the official Rclone image, you can update Rclone independently (`rclone/rclone:latest`) without waiting for updates to MQTT4Rclone.
* **Asynchronous Job Handling:** When an MQTT payload arrives, MQTT4Rclone maps the topic to the corresponding Rclone RC endpoint. Long-running operations (such as `sync/*` commands) are automatically executed asynchronously as background jobs, returning an immediate job ID while publishing execution results back to a designated MQTT response topic.

---

## Quick Start (Docker Compose)

MQTT4Rclone is distributed via [Docker Hub](https://hub.docker.com/r/legobas/mqtt4rclone) and is designed to be deployed alongside Rclone using Docker Compose.

### 1. Create `docker-compose.yml`

```yaml
services:
  rclone:
    image: rclone/rclone:latest
    container_name: rclone
    ports:
      # Optional: Expose port 5533 only if external access to the Rclone Web GUI/API is required
      - 5533:5533
    volumes:
      - /path/to/rclone/config:/config/rclone
      - /path/to/local/storage:/data
    command:
      - --api-addr=0.0.0.0:5533
      - --user=<RCLONE_RC_USER>
      - --pass=<RCLONE_RC_PASSWORD>
    restart: unless-stopped

  mqtt4rclone:
    image: legobas/mqtt4rclone:latest
    container_name: mqtt4rclone
    depends_on:
      - rclone
    environment:
      - PUID=1000              # User ID for config directory permissions
      - PGID=1000              # Group ID for config directory permissions
      - LOGLEVEL=INFO          # MQTT4Rclone log level: ERROR, INFO (default), DEBUG, TRACE
      - RCLONE_LOGLEVEL=INFO   # Log level for Rclone API responses
      - TZ=America/New_York    # Container timezone
    volumes:
      - /path/to/mqtt4rclone/config:/config
    restart: unless-stopped
```

> **Note on Volumes & Security:**
> * **Data Volumes:** Only the `rclone` container requires access to your local data volume (`/data`). All file paths specified in MQTT payloads are resolved relative to the `rclone` container.
> * **Network Security:** For optimal security, avoid exposing port `5533` to the host network unless you explicitly require external access to the Rclone Web GUI.

### 2. Deploy and Update

Start both services in detached mode:
```bash
docker compose up -d
```

To update either Rclone or MQTT4Rclone to the latest release:
```bash
docker compose pull && docker compose up -d
```

---

## Configuration

### 1. MQTT4Rclone Configuration (`mqtt4rclone.yml`)

Place a `mqtt4rclone.yml` file inside the directory mounted to `/config` in the `mqtt4rclone` container.

| Section / Key | Type | Required | Default | Description |
| :--- | :--- | :---: | :--- | :--- |
| **`mqtt`** | | | | |
| `url` | String | **Yes** | — | Connection URL of the MQTT broker (e.g., `tcp://192.168.1.10:1883`). |
| `username` | String | No | `""` | Username for MQTT broker authentication. |
| `password` | String | No | `""` | Password for MQTT broker authentication. |
| `qos` | Integer | No | `0` | MQTT Quality of Service level (`0`, `1`, or `2`). |
| **`rclone`** | | | | |
| `rc_username` | String | No | `""` | Username matching the `--user` flag passed to the Rclone container. |
| `rc_password` | String | No | `""` | Password matching the `--pass` flag passed to the Rclone container. |
| `response_topic`| String | No | `mqtt4rclone/response` | MQTT topic where Rclone API responses are published. |

#### Example `mqtt4rclone.yml`

```yaml
mqtt:
  url: "tcp://<MQTT_SERVER>:1883"
  username: "<MQTT_USERNAME>"
  password: "<MQTT_PASSWORD>"
  qos: 0

rclone:
  rc_username: "<RCLONE_RC_USER>"
  rc_password: "<RCLONE_RC_PASSWORD>"
  response_topic: "mqtt4rclone/response"
```

### 2. Rclone Configuration

Ensure a valid `rclone.conf` file exists in the directory mounted to `/config/rclone`. The credentials configured under `rclone.rc_username` and `rclone.rc_password` in `mqtt4rclone.yml` must match the `--user` and `--pass` flags defined in the Rclone container command.

---

## Usage & Command Reference

MQTT4Rclone translates incoming MQTT messages directly into [Rclone Remote Control API](https://rclone.org/rc/#supported-commands) requests:

* **MQTT Topic:** `mqtt4rclone/<rc_endpoint>`
* **MQTT Payload:** Valid JSON matching the parameters expected by the Rclone RC endpoint (use `{}` or an empty payload if no parameters are required).
* **Async Behavior:** All `sync/*` commands automatically have `"_async": true` injected into the payload so long-running transfers do not block the API.

### Command Examples

#### Retrieve Core Statistics
* **Topic:** `mqtt4rclone/core/stats`
* **Payload:** `{}`

#### List Configured Remotes
* **Topic:** `mqtt4rclone/config/listremotes`
* **Payload:** `{}`

#### Check Remote Filesystem Info
* **Topic:** `mqtt4rclone/operations/fsinfo`
* **Payload:**
  ```json
  {
    "fs": "dropbox:"
  }
  ```

#### Sync Local Directory to Remote
* **Topic:** `mqtt4rclone/sync/sync`
* **Payload:**
  ```json
  {
    "srcFs": "/data/mydropbox",
    "dstFs": "dropbox:"
  }
  ```

#### Copy Recent Files (Filtered by Age)
* **Topic:** `mqtt4rclone/sync/copy`
* **Payload:**
  ```json
  {
    "srcFs": "/data/mydropbox",
    "dstFs": "dropbox:",
    "_filter": {
      "MaxAge": "1d"
    }
  }
  ```

### API Responses

After executing a command, MQTT4Rclone publishes the JSON response from the Rclone API to `mqtt4rclone/response` (or your custom `response_topic`). Downstream automation tools can subscribe to this topic to track job IDs, verify completion, or trigger error alerts.

---

## Scheduled Automation

To trigger scheduled synchronization tasks via MQTT, you can pair MQTT4Rclone with [MQTT-Timer](https://github.com/Legobas/mqtt-timer):

```yaml
timers:
  - id: rclone_dropbox
    time: 01:00:00
    description: Daily sync from local storage to Dropbox at 01:00
    topic: mqtt4rclone/sync/sync
    message: '{"srcFs":"/data/mydropbox","dstFs":"dropbox:"}'

  - id: rclone_mega
    cron: 00 * * * *
    description: Hourly copy from Mega to local storage
    topic: mqtt4rclone/sync/copy
    message: '{"srcFs":"mega:","dstFs":"/data/mymega"}'
```

---

## Logging & Troubleshooting

### Viewing Logs
Both containers output logs to `stdout`. You can inspect them using standard Docker commands:

```bash
docker compose logs -f mqtt4rclone
```

### Dynamic Rclone Log Levels
Rclone supports dynamic log-level adjustments (`DEBUG`, `INFO`, `NOTICE`, `ERROR`; default is `NOTICE`) via the RC API.

To temporarily enable `DEBUG` logging without restarting the container, publish the following message:
* **Topic:** `mqtt4rclone/options/set`
* **Payload:**
  ```json
  {
    "main": {
      "LogLevel": "DEBUG"
    }
  }
  ```

To revert to normal logging, restart the Rclone container or publish `"LogLevel": "NOTICE"` to the same topic.

---

## Acknowledgments

* [Rclone](https://rclone.org) — Cloud storage sync engine
* [Eclipse Paho MQTT Go Client](https://github.com) — MQTT connectivity
* [Zerolog](https://github.com) — Structured JSON logging
