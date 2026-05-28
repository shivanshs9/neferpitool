# SigNoz alerting guide for neferpitool

This guide lists every structured event, recommended alerts, and common pitfalls so you can configure SigNoz once and not miss signals.

## Prerequisites

Configure the neferpitool deployment **before** creating alerts:

| Setting | Value | Why |
|---------|-------|-----|
| `EVENTS_ENABLED` | `true` | JSON `event.name` lines on stdout (best for alerts) |
| `LOG_PLAIN` | `true` | Parseable logs in Kubernetes / `kubectl logs` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | e.g. `http://<host>:4318` | Traces in SigNoz (HTTP OTLP) |
| `OTEL_EXPORTER_OTLP_INSECURE` | `true` (if no TLS) | Node/host collector without TLS |
| `OTEL_RESOURCE_ATTRIBUTES` | `service.name=neferpitool-app,...` | Service name in Traces UI |
| `MONITOR_ASSET_DNS_CHANGES` | `true` | Asset DNS + apex WHOIS change checks |
| Volume mount | `/app/config/database` only | Do not mount over `/app/config` (hides baked config) |

On startup you should see:

- `OTEL trace export enabled` — OTLP traces active
- Per cycle: `neferpitool.monitor.cycle` span (if OTLP works)

**Two pipelines:** logs (stdout JSON) and traces (OTLP). You can alert on either; **logs are simpler** for `event.name` filters.

---

## Event catalog

All security-relevant events use span/log name = `event.name`. The monitored hostname is **`monitored.domain`** (not `host.name`, which SigNoz reserves for the node/EC2).

| `event.name` | Severity | When it fires | Key attributes |
|--------------|----------|---------------|----------------|
| `neferpitool.dns.changed` | **High** | Public DNS record changed (asset host, verified twice) | `zone`, `monitored.domain`, `field`, `before`, `after` |
| `neferpitool.whois.changed` | **High** | Apex WHOIS changed (registrar view, verified twice) | `zone`, `monitored.domain` (= apex), `field`, `before`, `after` |
| `neferpitool.typo.activated` | **Medium** | Typosquat host became resolvable / active | `zone`, `monitored.domain`, `status.from`, `status.to` |
| `neferpitool.host.discovered` | **Low** | New subdomain asset added to DB | `zone`, `monitored.domain`, `source` |
| `neferpitool.typo.discovered` | **Info** | New typo host generated (deferred mode) | `zone`, `monitored.domain`, `algorithm` |
| `neferpitool.monitor.cycle` | **Heartbeat** | Background cycle completed | (cycle wrapper only) |

### WHOIS `field` values

| `field` | Meaning |
|---------|---------|
| `RegistrantName Registrant` | Registrant name changed |
| `Organization` | Registrant org changed |
| `WHOIS Name Servers` | Registrar nameserver list changed (order ignored) |
| `Expiration Date` | Registry expiration changed |

WHOIS runs **only on the zone apex** (e.g. `avantisfinance.net`), not subdomains.

### DNS `field` values

| `field` | Examples |
|---------|----------|
| `DNS A`, `DNS AAAA`, `DNS CNAME`, `DNS NS`, `DNS MX`, `DNS SOA` | Public resolver answers |

**DNS vs WHOIS nameservers:** `DNS NS` = what recursive DNS returns (may be Cloudflare proxy). `WHOIS Name Servers` = delegation in registry/WHOIS (registrar change).

---

## Recommended alerts (priority order)

### 1. DNS record changed (primary)

**Channel:** Log alert (preferred) or Trace alert

**Log query (example):**

```
body contains "neferpitool.dns.changed"
```

Or if JSON parsed:

```
event.name = neferpitool.dns.changed
```

**Optional filters:**

- `zone = avantisfinance.net`
- `field` contains `CNAME` (CNAME-specific alert)

**Trace alert:**

- Span name = `neferpitool.dns.changed`
- Optional: `monitored.domain` exists

**Notify:** Slack/PagerDuty immediately.

---

### 2. WHOIS / registrar change (apex)

**Channel:** Log or Trace

**Log query:**

```
body contains "neferpitool.whois.changed"
```

**High-signal filters:**

```
field = "WHOIS Name Servers"
```

```
field = "RegistrantName Registrant"
```

**Trace alert:** span name = `neferpitool.whois.changed`

**Notify:** Security + domain owners (possible takeover or transfer).

---

### 3. Typosquat activated

**Log query:**

```
body contains "neferpitool.typo.activated"
```

**Notify:** Brand protection channel; tune per zone.

---

### 4. New monitored subdomain (optional)

**Log query:**

```
body contains "neferpitool.host.discovered"
```

Use for awareness, not paging, unless you expect a fixed asset list.

---

### 5. Monitor heartbeat missing (optional)

**Trace alert only:**

- No span `neferpitool.monitor.cycle` for **> 15–30 minutes** (set above `MINUTESLEEPBACKGROUNDMONITORING` + scan time)

**Do not page** on single missed cycle if scans are slow.

---

## Alerts to avoid (noise)

| Event | Why skip |
|-------|----------|
| `neferpitool.typo.discovered` | Thousands on first deferred typo run |
| `neferpitool.monitor.cycle` | Fires every minute; not a security event |
| All `neferpitool.dns.changed` on CF-proxied hosts without tuning | May miss origin changes; see below |

---

## SigNoz setup checklist

### Log ingestion

- [ ] Container stdout collected (DaemonSet / k8s log collector → SigNoz)
- [ ] Parser extracts JSON or search works on raw `body`
- [ ] Test: run one cycle, search logs for `event.name`

### Trace ingestion

- [ ] `OTEL_EXPORTER_OTLP_ENDPOINT` reachable from pod (`HOST_IP:4318` or cluster collector)
- [ ] Service appears as `neferpitool-app` (or your `service.name`)
- [ ] Test: Traces → filter `name = neferpitool.monitor.cycle`

### Create alerts

- [ ] **DNS changed** — log or trace, severity critical
- [ ] **WHOIS changed** — log or trace, severity critical
- [ ] **Typo activated** — log, severity warning
- [ ] (Optional) heartbeat on `monitor.cycle`
- [ ] Notification channel wired (Slack/email)

### Validate

- [ ] Trigger DNS test on a **non-proxied** host or grey-cloud label
- [ ] Trigger WHOIS test (see below)
- [ ] Confirm alert fires with `monitored.domain` in payload

---

## Testing alerts

### DNS

Use a host that is in the asset list and **not** orange-cloud proxied (or accept only edge IP stability). Change a TXT or grey-cloud CNAME, wait one full cycle (~1 min + scan time).

### WHOIS (apex only)

1. Let neferpitool store apex WHOIS in SQLite (`config/database`).
2. Stop the pod.
3. Edit the apex row’s WHOIS JSON in the DB (e.g. change `name_servers` in stored registrar data).
4. Start pod; after `change-check` + `change-verify` you should see:
   - Log: `{"event.name":"neferpitool.whois.changed",...,"field":"WHOIS Name Servers",...}`
   - Trace: span `neferpitool.whois.changed`

Or run unit tests: `go test ./pkg/changes/... -run WHOIS -v`

**Requirements for live WHOIS scans:**

- Apex status **Inactive** or **Active** (not Alias-only)
- Same status on old vs new snapshot (status change suppresses WHOIS diff rows)

---

## Cloudflare proxy caveat

For **orange-cloud** hosts, `neferpitool.dns.changed` reflects **public DNS** (often stable Cloudflare anycast IPs), not the origin CNAME you set in the dashboard.

- Use **DNS-only** (grey cloud) for names you need in DNS alerts, or
- Alert on `neferpitool.whois.changed` for registrar/delegation changes, or
- Monitor origin out-of-band

---

## Example JSON payloads

**DNS:**

```json
{
  "event.name": "neferpitool.dns.changed",
  "timestamp": "2026-05-28T12:00:00Z",
  "zone": "avantisfinance.net",
  "monitored.domain": "api.avantisfinance.net",
  "field": "DNS CNAME",
  "before": "api.avantisfinance.net.\t300\tIN\tCNAME\told.example.com.",
  "after": "api.avantisfinance.net.\t300\tIN\tCNAME\tnew.example.com.",
  "change.type": "dns"
}
```

**WHOIS:**

```json
{
  "event.name": "neferpitool.whois.changed",
  "timestamp": "2026-05-28T12:00:00Z",
  "zone": "avantisfinance.net",
  "monitored.domain": "avantisfinance.net",
  "field": "WHOIS Name Servers",
  "before": "ns1.cloudflare.com, ns2.cloudflare.com",
  "after": "ns1.otherdns.com, ns2.otherdns.com",
  "change.type": "whois"
}
```

---

## Quick reference: trace span names

| Span name | Alert? |
|-----------|--------|
| `neferpitool.dns.changed` | Yes |
| `neferpitool.whois.changed` | Yes |
| `neferpitool.typo.activated` | Yes |
| `neferpitool.host.discovered` | Optional |
| `neferpitool.typo.discovered` | No (noise) |
| `neferpitool.monitor.cycle` | Heartbeat only |

---

## Related docs

- [docker.md](./docker.md) — image, env vars, volumes
- [README.md](../README.md) — monitoring modes and config
