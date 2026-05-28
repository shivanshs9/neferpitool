# Docker

## Build

From the repository root:

```bash
docker build -t neferpitool:latest .
```

## Run (background monitoring)

Register apex zones and start the monitor loop:

```bash
docker run --rm -it \
  --name neferpitool \
  -v neferpitool-config:/app/config \
  neferpitool:latest \
  -bg avantisfi.com avantisfinance.net
```

Zones passed after `-bg` are added to the database if missing, then monitoring runs for **all** zones in the DB.

One-shot add (no background):

```bash
docker run --rm -v neferpitool-config:/app/config neferpitool:latest avantisfi.com
```

## Configuration

Settings are loaded from `/app/config/config.json`, then **overridden** by environment variables when set.

| Environment variable | Description |
|---------------------|-------------|
| `LOG_PLAIN` | `true` in image — plain `INFO`/`WARN` logs (no ANSI) |
| `EVENTS_ENABLED` | `true` in image — JSON `event.name` lines for SigNoz |
| `TYPO_MODE` | `off`, `immediate`, or `deferred` |
| `DISCOVERY_CT` | Enable Certificate Transparency discovery |
| `DISCOVERY_WORDLIST` | Enable wordlist subdomain brute-force |
| `PATHRESOLVER` | DNS resolver file (default in image: `/etc/resolv.conf`) |
| `MINUTESLEEPBACKGROUNDMONITORING` | Minutes between monitor cycles |

Mount a custom config:

```bash
docker run -v $(pwd)/my-config:/app/config neferpitool:latest -bg example.com
```

## OpenTelemetry and SigNoz

Neferpitool currently exports **structured events as JSON log lines** on stdout when `EVENTS_ENABLED=true`, for example:

```json
{"event.name":"neferpitool.typo.activated","zone":"example.com","host.name":"examp1e.com",...}
```

### SigNoz today (log pipeline)

1. Run the container with `LOG_PLAIN=true` and `EVENTS_ENABLED=true` (defaults in the Dockerfile).
2. Collect container stdout with your platform (Kubernetes → SigNoz log collector, Docker → OTLP log receiver).
3. In SigNoz, query logs: `event.name = neferpitool.typo.activated` or `neferpitool.typo.discovered`.
4. Create alerts on those patterns or on log-based metrics.

### OTEL environment variables (Dockerfile `ENV`)

The image sets standard [OTEL SDK environment variables](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/) so they are ready when native OTLP export is added, and so sidecars or the SigNoz collector can use the same conventions:

| Variable | Default in image | Purpose |
|----------|------------------|---------|
| `OTEL_SERVICE_NAME` | `neferpitool` | Service name in traces/logs/metrics |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://signoz-otel-collector:4318` | SigNoz OTLP HTTP endpoint (adjust for your stack) |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `http/protobuf` | Use `grpc` with port `4317` if needed |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` | Set `false` when using TLS |
| `OTEL_RESOURCE_ATTRIBUTES` | `deployment.environment=production` | Extra resource attributes |

Example override for Kubernetes:

```yaml
env:
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: http://signoz-otel-collector.observability.svc.cluster.local:4318
  - name: OTEL_RESOURCE_ATTRIBUTES
    value: deployment.environment=prod,service.namespace=security
  - name: EVENTS_ENABLED
    value: "true"
  - name: LOG_PLAIN
    value: "true"
  - name: TYPO_MODE
    value: deferred
```

When native OTLP log export is implemented, you may also set:

- `OTEL_LOGS_EXPORTER=otlp`
- `OTEL_TRACES_EXPORTER=otlp`
- `OTEL_METRICS_EXPORTER=otlp`

Until then, rely on stdout JSON events + log ingestion.

## Kubernetes (sketch)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: neferpitool
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: neferpitool
          image: neferpitool:latest
          args: ["-bg", "example.com"]
          envFrom:
            - configMapRef:
                name: neferpitool-config
          volumeMounts:
            - name: data
              mountPath: /app/config
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: neferpitool-config
```
