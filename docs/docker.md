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
  -v neferpitool-data:/app/config/database \
  neferpitool:latest \
  -bg avantisfi.com avantisfinance.net
```

`config.json` and `subdomains.txt` are **included in the image** at `/app/config/`. Mount only `/app/config/database` for persistent zone/typo state unless you intentionally override config.

Zones passed after `-bg` are added to the database if missing, then monitoring runs for **all** zones in the DB.

One-shot add (no background):

```bash
docker run --rm -v neferpitool-data:/app/config/database neferpitool:latest avantisfi.com
```

## Configuration

Settings are loaded from `/app/config/config.json` (baked into the image from `cmd/config/config.json`), then **overridden** by environment variables when set.

Subdomain wordlist: `/app/config/subdomains.txt` (source: `cmd/config/subdomains.txt`, referenced by `SUBDOMAIN_WORDLIST_PATH` in config).

| Environment variable | Description |
|---------------------|-------------|
| `LOG_PLAIN` | `true` in image — plain `INFO`/`WARN` logs (no ANSI) |
| `EVENTS_ENABLED` | `true` in image — JSON `event.name` lines for SigNoz |
| `TYPO_MODE` | `off`, `immediate`, or `deferred` |
| `DISCOVERY_CT` | Enable Certificate Transparency discovery |
| `DISCOVERY_WORDLIST` | Enable wordlist subdomain brute-force |
| `PATHRESOLVER` | DNS resolver file (default in image: `/etc/resolv.conf`) |
| `MINUTESLEEPBACKGROUNDMONITORING` | Minutes between monitor cycles |

Override config or wordlist (optional):

```bash
docker run \
  -v neferpitool-data:/app/config/database \
  -v $(pwd)/my-config.json:/app/config/config.json:ro \
  -v $(pwd)/my-subdomains.txt:/app/config/subdomains.txt:ro \
  neferpitool:latest -bg example.com
```

Avoid mounting an empty volume on `/app/config` — that hides the baked-in `config.json` and `subdomains.txt`.

## OpenTelemetry and SigNoz

Neferpitool uses **two channels**:

1. **Logs** — JSON lines on stdout when `EVENTS_ENABLED=true` (good for log-based alerts).
2. **Traces** — OTLP HTTP spans when `OTEL_EXPORTER_OTLP_ENDPOINT` is set (shows under **Services → Traces** in SigNoz).

Example log event:

```json
{"event.name":"neferpitool.dns.changed","zone":"example.com","monitored.domain":"api.example.com",...}
```

Trace span names match `event.name` (e.g. `neferpitool.dns.changed`, `neferpitool.monitor.cycle`).

Event attributes use `monitored.domain` for the FQDN under watch (not `host.name`, which SigNoz maps to the collector/node).

### SigNoz alerting

See **[signoz-alerting.md](./signoz-alerting.md)** for the full checklist, every `event.name`, log/trace query examples, WHOIS vs DNS nameserver notes, and test procedures.

Key events: `neferpitool.dns.changed`, `neferpitool.whois.changed`, `neferpitool.typo.activated`, `neferpitool.host.discovered`, `neferpitool.monitor.cycle` (heartbeat).

### SigNoz: logs vs traces

| Goal | What to configure |
|------|-------------------|
| Log queries / log alerts | `EVENTS_ENABLED=true`, `LOG_PLAIN=true`, ingest container stdout |
| Service map & trace view | `OTEL_EXPORTER_OTLP_ENDPOINT` (HTTP port **4318** on node/host collector), `OTEL_EXPORTER_OTLP_INSECURE=true` if needed |
| Service name in UI | `OTEL_SERVICE_NAME` or `service.name=...` inside `OTEL_RESOURCE_ATTRIBUTES` |

JSON log lines alone do **not** create a new service in the Traces tab; you need OTLP export (or a log→trace pipeline).

### OTEL environment variables (Dockerfile `ENV`)

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

On startup you should see `OTEL trace export enabled` in logs when the endpoint is reachable. Each DNS/typo event also creates a short-lived span with the same attributes as the JSON payload.

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
              mountPath: /app/config/database
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: neferpitool-data
```
