FROM alpine:3 AS builder
ARG natscli_version=0.1.4
ARG natscli_download_url=https://github.com/nats-io/natscli/releases/download/v$natscli_version/nats-$natscli_version-linux-amd64.zip
RUN wget $natscli_download_url
RUN unzip -j nats*

FROM jeffail/benthos:4
COPY --from=builder /nats .
