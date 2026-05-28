# syntax=docker/dockerfile:1

# --- Build (CGO + sqlite on musl) ---
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/neferpitool \
    ./cmd

# --- Minimal runtime ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates sqlite-libs tzdata \
    && addgroup -S -g 10001 neferpitool \
    && adduser -S -u 10001 -G neferpitool -h /app -s /sbin/nologin neferpitool

WORKDIR /app

COPY --from=builder /out/neferpitool /app/neferpitool

# Baked-in config (not replaced unless you mount over /app/config or individual files).
COPY cmd/config/config.json /app/config/config.json
COPY cmd/config/subdomains.txt /app/config/subdomains.txt
COPY cmd/config/listTLD.json /app/config/listTLD.json
COPY cmd/config/emailTemplates/ /app/config/emailTemplates/

RUN sed -i \
      -e 's|"PATHRESOLVER": "./config/resolv.conf"|"PATHRESOLVER": "/etc/resolv.conf"|' \
      -e 's|"SUBDOMAIN_WORDLIST_PATH": "./config/subdomains.txt"|"SUBDOMAIN_WORDLIST_PATH": "/app/config/subdomains.txt"|' \
      /app/config/config.json \
    && mkdir -p /app/config/database \
    && chown -R neferpitool:neferpitool /app

USER neferpitool

ENV LOG_PLAIN=true \
    EVENTS_ENABLED=true \
    OTEL_SERVICE_NAME=neferpitool \
    OTEL_EXPORTER_OTLP_ENDPOINT=http://signoz-otel-collector:4318 \
    OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf \
    OTEL_EXPORTER_OTLP_INSECURE=true \
    OTEL_RESOURCE_ATTRIBUTES=deployment.environment=production

# SQLite only — config.json and subdomains.txt stay in the image layer.
VOLUME ["/app/config/database"]

ENTRYPOINT ["/app/neferpitool"]
CMD ["-bg"]
