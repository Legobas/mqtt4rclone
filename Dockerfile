FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o app .

FROM alpine:latest
RUN echo http://dl-2.alpinelinux.org/alpine/edge/community/ >> /etc/apk/repositories
RUN apk update && apk upgrade
RUN apk add -U shadow
RUN apk add --no-cache tzdata multirun curl unzip bash
RUN curl https://rclone.org/install.sh | bash

ENV PUID=1000
ENV PGID=1000
ENV TZ="Europe/London"
ENV RCLONE_LOGLEVEL="NOTICE"
RUN addgroup -g "${PGID}" -S appusers
RUN adduser -S -D -H -h /config -u "${PUID}" -g "${PGID}" appuser
RUN mkdir /config
COPY --from=builder /build/app /bin
COPY --from=builder /build/start.sh /bin
RUN chmod +x /bin/start.sh
RUN chmod +x /usr/bin/multirun
WORKDIR /config
VOLUME /config
VOLUME /data
CMD ["/bin/start.sh"]
