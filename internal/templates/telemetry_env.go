package templates

import (
	"strconv"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// TelemetryEnvVar is a single KEY=VALUE default emitted into the generated
// .env.example for the OTel-aligned telemetry configuration.
type TelemetryEnvVar struct {
	Key   string
	Value string
}

// telemetryEnvVars maps a manifest's spec.telemetry exporter blocks to the
// standard OpenTelemetry SDK environment variables, following the schema's 1:1
// field -> OTEL_* contract. It returns only the OTEL_* variables; the master
// switch (A2A_TELEMETRY_ENABLE / TELEMETRY_ENABLE) is emitted by the template
// since its name is language-specific.
//
// Signal selection follows the schema's declarative-config model: a present
// traces/metrics exporter sets OTEL_{TRACES,METRICS}_EXPORTER to the chosen key
// (otlp or, for metrics, prometheus); omitting the signal disables it (=none).
// When both traces and metrics push over OTLP with identical endpoint/protocol
// the shared OTEL_EXPORTER_OTLP_* pair is emitted; otherwise the per-signal
// OTEL_EXPORTER_OTLP_{TRACES,METRICS}_* variants are. An exporter field that is
// omitted in the manifest is left out so the OTLP/Prometheus SDK default applies.
//
// Prometheus pull is Go-only: the TypeScript ADK does not support it yet, so its
// OTEL_EXPORTER_PROMETHEUS_* defaults are skipped for TypeScript (the validator
// warns on that combination separately).
func telemetryEnvVars(adl *schema.ADL) []TelemetryEnvVar {
	if adl == nil || adl.Spec.Telemetry == nil || !adl.Spec.Telemetry.Enabled {
		return nil
	}
	lang := DetectLanguageFromADL(adl)
	if lang == "rust" {
		return nil
	}

	tel := adl.Spec.Telemetry

	var tracesOtlp *schema.TelemetryOTLPExporter
	if tel.Traces != nil && tel.Traces.Exporter != nil {
		tracesOtlp = tel.Traces.Exporter.Otlp
	}

	var metricsOtlp *schema.TelemetryOTLPExporter
	var metricsProm *schema.TelemetryPrometheusExporter
	if tel.Metrics != nil && tel.Metrics.Exporter != nil {
		metricsOtlp = tel.Metrics.Exporter.Otlp
		metricsProm = tel.Metrics.Exporter.Prometheus
	}

	out := []TelemetryEnvVar{
		{Key: "OTEL_TRACES_EXPORTER", Value: exporterValue(tracesOtlp != nil, "otlp")},
	}

	switch {
	case metricsOtlp != nil:
		out = append(out, TelemetryEnvVar{Key: "OTEL_METRICS_EXPORTER", Value: "otlp"})
	case metricsProm != nil:
		out = append(out, TelemetryEnvVar{Key: "OTEL_METRICS_EXPORTER", Value: "prometheus"})
	default:
		out = append(out, TelemetryEnvVar{Key: "OTEL_METRICS_EXPORTER", Value: "none"})
	}

	shared := tracesOtlp != nil && metricsOtlp != nil &&
		tracesOtlp.Endpoint == metricsOtlp.Endpoint &&
		tracesOtlp.Protocol == metricsOtlp.Protocol
	if shared {
		out = appendOTLPEnv(out, "", tracesOtlp)
	} else {
		out = appendOTLPEnv(out, "TRACES_", tracesOtlp)
		out = appendOTLPEnv(out, "METRICS_", metricsOtlp)
	}

	if metricsProm != nil && lang == "go" {
		if metricsProm.Host != "" {
			out = append(out, TelemetryEnvVar{Key: "OTEL_EXPORTER_PROMETHEUS_HOST", Value: metricsProm.Host})
		}
		if metricsProm.Port != 0 {
			out = append(out, TelemetryEnvVar{Key: "OTEL_EXPORTER_PROMETHEUS_PORT", Value: strconv.Itoa(metricsProm.Port)})
		}
	}

	return out
}

// exporterValue returns the chosen exporter key when the signal is configured,
// or "none" to disable it.
func exporterValue(configured bool, key string) string {
	if configured {
		return key
	}
	return "none"
}

// appendOTLPEnv appends the endpoint/protocol vars for one OTLP exporter using
// the given infix ("" shared, or "TRACES_"/"METRICS_" per-signal). A nil
// exporter or an omitted field contributes nothing so the SDK default applies.
func appendOTLPEnv(out []TelemetryEnvVar, infix string, otlp *schema.TelemetryOTLPExporter) []TelemetryEnvVar {
	if otlp == nil {
		return out
	}
	if otlp.Endpoint != "" {
		out = append(out, TelemetryEnvVar{Key: "OTEL_EXPORTER_OTLP_" + infix + "ENDPOINT", Value: otlp.Endpoint})
	}
	if otlp.Protocol != "" {
		out = append(out, TelemetryEnvVar{Key: "OTEL_EXPORTER_OTLP_" + infix + "PROTOCOL", Value: string(otlp.Protocol)})
	}
	return out
}
