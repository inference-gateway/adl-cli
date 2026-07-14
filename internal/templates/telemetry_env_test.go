package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// otlpExporter is a small constructor for a fully-specified OTLP exporter block.
func otlpExporter(endpoint string, proto schema.TelemetryOTLPExporterProtocol) *schema.TelemetryOTLPExporter {
	return &schema.TelemetryOTLPExporter{Endpoint: endpoint, Protocol: proto}
}

// envMap flattens the emitted telemetry vars into a name -> value lookup.
func envMap(vars []TelemetryEnvVar) map[string]string {
	m := make(map[string]string, len(vars))
	for _, v := range vars {
		m[v.Key] = v.Value
	}
	return m
}

// TestTelemetryEnvVars_GoUsesA2APrefix asserts the Go ADK contract: the OTel SDK
// variables are emitted as A2A_OTEL_* because the generated config reads them
// through the ADK's A2A_-prefixed serverConfig. It also pins the shared OTLP
// pair for the traces-otlp + metrics-prometheus example: the Go ADK exposes only
// the shared OTEL_EXPORTER_OTLP_{ENDPOINT,PROTOCOL} fields, so no per-signal
// A2A_OTEL_EXPORTER_OTLP_TRACES_* variant (which nothing would read) is emitted.
func TestTelemetryEnvVars_GoUsesA2APrefix(t *testing.T) {
	adl := minimalGoADL()
	adl.Spec.Telemetry = &schema.TelemetryConfig{
		Enabled: true,
		Traces: &schema.TelemetryTracesConfig{
			Exporter: &schema.TelemetryTracesExporter{
				Otlp: otlpExporter("http://localhost:4318", schema.TelemetryOTLPExporterProtocolHttpProtobuf),
			},
		},
		Metrics: &schema.TelemetryMetricsConfig{
			Exporter: &schema.TelemetryMetricsExporter{
				Prometheus: &schema.TelemetryPrometheusExporter{Port: 9464},
			},
		},
	}

	got := envMap(telemetryEnvVars(adl))

	want := map[string]string{
		"A2A_OTEL_TRACES_EXPORTER":          "otlp",
		"A2A_OTEL_METRICS_EXPORTER":         "prometheus",
		"A2A_OTEL_EXPORTER_OTLP_ENDPOINT":   "http://localhost:4318",
		"A2A_OTEL_EXPORTER_OTLP_PROTOCOL":   "http/protobuf",
		"A2A_OTEL_EXPORTER_PROMETHEUS_PORT": "9464",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("Go env %s = %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
	for k := range got {
		if !strings.HasPrefix(k, "A2A_OTEL_") {
			t.Errorf("Go telemetry var %q must carry the A2A_ prefix", k)
		}
		// The Go ADK has no per-signal OTLP endpoint fields, so the shared pair
		// must be used - never OTEL_EXPORTER_OTLP_{TRACES,METRICS}_* (which the
		// exporter-selection vars OTEL_{TRACES,METRICS}_EXPORTER don't match).
		if strings.Contains(k, "OTEL_EXPORTER_OTLP_TRACES_") || strings.Contains(k, "OTEL_EXPORTER_OTLP_METRICS_") {
			t.Errorf("Go must emit the shared OTLP pair, not per-signal var %q", k)
		}
	}
}

// TestTelemetryEnvVars_TypeScriptStaysBare asserts the TypeScript contract: the
// OTLP variables stay bare OTEL_* because the OpenTelemetry Node SDK reads them
// from process.env directly, so prefixing them with A2A_ would disable telemetry.
func TestTelemetryEnvVars_TypeScriptStaysBare(t *testing.T) {
	adl := minimalTypeScriptADL()
	adl.Spec.Telemetry = &schema.TelemetryConfig{
		Enabled: true,
		Traces: &schema.TelemetryTracesConfig{
			Exporter: &schema.TelemetryTracesExporter{
				Otlp: otlpExporter("http://localhost:4318", schema.TelemetryOTLPExporterProtocolHttpProtobuf),
			},
		},
		Metrics: &schema.TelemetryMetricsConfig{
			Exporter: &schema.TelemetryMetricsExporter{
				Otlp: otlpExporter("http://localhost:4318", schema.TelemetryOTLPExporterProtocolHttpProtobuf),
			},
		},
	}

	got := envMap(telemetryEnvVars(adl))

	want := map[string]string{
		"OTEL_TRACES_EXPORTER":        "otlp",
		"OTEL_METRICS_EXPORTER":       "otlp",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318",
		"OTEL_EXPORTER_OTLP_PROTOCOL": "http/protobuf",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("TS env %s = %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
	for k := range got {
		if strings.HasPrefix(k, "A2A_") {
			t.Errorf("TypeScript telemetry var %q must not carry the A2A_ prefix", k)
		}
	}
}

// TestTelemetryEnvVars_DisabledAndRust confirms the two early-outs: telemetry
// off yields nothing, and Rust ignores spec.telemetry entirely.
func TestTelemetryEnvVars_DisabledAndRust(t *testing.T) {
	off := minimalGoADL()
	off.Spec.Telemetry = &schema.TelemetryConfig{Enabled: false}
	if got := telemetryEnvVars(off); got != nil {
		t.Errorf("disabled telemetry should emit no vars, got %v", got)
	}

	rust := minimalRustADL()
	rust.Spec.Telemetry = &schema.TelemetryConfig{Enabled: true}
	if got := telemetryEnvVars(rust); got != nil {
		t.Errorf("Rust telemetry should emit no vars, got %v", got)
	}
}
