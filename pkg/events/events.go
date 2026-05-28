package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/moorada/neferpitool/pkg/configuration"
	"github.com/moorada/neferpitool/pkg/log"
	"github.com/moorada/neferpitool/pkg/telemetry"
)

const (
	TypoDiscovered = "neferpitool.typo.discovered"
	TypoActivated  = "neferpitool.typo.activated"
	DNSChanged     = "neferpitool.dns.changed"
	HostDiscovered = "neferpitool.host.discovered"
)

// Emit writes a structured JSON log line and an OTEL trace span (when OTLP is configured).
func Emit(name string, attrs map[string]string) {
	EmitContext(context.Background(), name, attrs)
}

// EmitContext attaches the event to an existing trace context.
func EmitContext(ctx context.Context, name string, attrs map[string]string) {
	if attrs == nil {
		attrs = map[string]string{}
	}
	attrs["event.name"] = name

	if configuration.GetConf().EVENTS_ENABLED {
		payload := map[string]interface{}{
			"event.name": name,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}
		for k, v := range attrs {
			if k == "event.name" {
				continue
			}
			payload[k] = v
		}
		b, err := json.Marshal(payload)
		if err != nil {
			log.Error("event marshal: %s", err.Error())
		} else {
			log.Info("%s", string(b))
		}
	}

	telemetry.RecordEvent(ctx, name, attrs)
}

func EmitTypoDiscovered(zone, host, algorithm, status string) {
	Emit(TypoDiscovered, map[string]string{
		"zone":      zone,
		"host.name": host,
		"algorithm": algorithm,
		"status":    status,
		"source":    "typo",
	})
}

func EmitTypoActivated(zone, host, fromStatus, toStatus string) {
	Emit(TypoActivated, map[string]string{
		"zone":        zone,
		"host.name":   host,
		"status.from": fromStatus,
		"status.to":   toStatus,
		"source":      "typo",
	})
}

func EmitHostDiscovered(zone, host, source, status string) {
	Emit(HostDiscovered, map[string]string{
		"zone":      zone,
		"host.name": host,
		"source":    source,
		"status":    status,
	})
}
