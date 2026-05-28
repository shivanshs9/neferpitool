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
COPY cmd/config/ /app/config/

RUN sed -i 's|"PATHRESOLVER": "./config/resolv.conf"|"PATHRESOLVER": "/etc/resolv.conf"|' /app/config/config.json \
    && chown -R neferpitool:neferpitool /app

USER neferpitool

ENV LOG_PLAIN=true \
    EVENTS_ENABLED=true \
    OTEL_SERVICE_NAME=neferpitool \
    OTEL_EXPORTER_OTLP_ENDPOINT=http://signoz-otel-collector:4318 \
    OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf \
    OTEL_EXPORTER_OTLP_INSECURE=true \
    OTEL_RESOURCE_ATTRIBUTES=deployment.environment=production

VOLUME ["/app/config"]

ENTRYPOINT ["/app/neferpitool"]
CMD ["-bg"]
