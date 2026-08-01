package templates

import (
	"strconv"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// TelemetryEnvVar is a single KEY=VALUE default emitted into the generated
// .env.example for the OTel-aligned telemetry configuration. Description is the
// human-readable label the README env-var table renders alongside it.
type TelemetryEnvVar struct {
	Key         string
	Value       string
	Description string
}

// telemetryEnvVars maps a manifest's spec.telemetry exporter blocks to the
// OpenTelemetry SDK environment variables. It returns only the exporter
// variables; the master switch (A2A_TELEMETRY_ENABLED) is emitted by the
// template.
//
// Every variable carries the A2A_ prefix regardless of language, so agents
// document one consistent surface:
//   - Go reads the OTel settings through the ADK config, which nests the whole
//     server config under the A2A_ prefix (config.Config.A2A is
//     `env:",prefix=A2A_"`) while its OTelConfig carries no prefix of its own -
//     so every standard OTEL_* name is read as A2A_OTEL_*. The Go ADK only
//     exposes the shared OTEL_EXPORTER_OTLP_{ENDPOINT,PROTOCOL} fields (it has no
//     per-signal variants), so it always emits that shared pair from whichever
//     signal pushes over OTLP.
//   - TypeScript hands the OTLP settings to the OpenTelemetry Node SDK, which
//     reads the bare OTEL_* names from process.env - the generated index.ts
//     mirrors each A2A_OTEL_* var onto its bare name before SDK init. The Node
//     SDK honors the per-signal OTEL_EXPORTER_OTLP_{TRACES,METRICS}_* names, so
//     when traces and metrics push to different collectors it emits those and
//     otherwise collapses to the shared pair.
//
// Signal selection follows the schema's declarative-config model: a present
// traces/metrics exporter sets OTEL_{TRACES,METRICS}_EXPORTER to the chosen key
// (otlp or, for metrics, prometheus); omitting the signal disables it (=none).
// An exporter field that is omitted in the manifest is left out so the
// OTLP/Prometheus SDK default applies.
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

	const prefix = "A2A_"

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

	const metricsDesc = "Metrics exporter (`otlp`, `prometheus`, or `none`)"

	out := []TelemetryEnvVar{
		{Key: prefix + "OTEL_TRACES_EXPORTER", Value: exporterValue(tracesOtlp != nil, "otlp"),
			Description: "Trace exporter (`otlp` or `none`)"},
	}

	switch {
	case metricsOtlp != nil:
		out = append(out, TelemetryEnvVar{Key: prefix + "OTEL_METRICS_EXPORTER", Value: "otlp", Description: metricsDesc})
	case metricsProm != nil:
		out = append(out, TelemetryEnvVar{Key: prefix + "OTEL_METRICS_EXPORTER", Value: "prometheus", Description: metricsDesc})
	default:
		out = append(out, TelemetryEnvVar{Key: prefix + "OTEL_METRICS_EXPORTER", Value: "none", Description: metricsDesc})
	}

	if lang == "go" {
		otlp := tracesOtlp
		if otlp == nil {
			otlp = metricsOtlp
		}
		out = appendOTLPEnv(out, prefix, "", otlp)
	} else {
		shared := tracesOtlp != nil && metricsOtlp != nil &&
			tracesOtlp.Endpoint == metricsOtlp.Endpoint &&
			tracesOtlp.Protocol == metricsOtlp.Protocol
		if shared {
			out = appendOTLPEnv(out, prefix, "", tracesOtlp)
		} else {
			out = appendOTLPEnv(out, prefix, "TRACES_", tracesOtlp)
			out = appendOTLPEnv(out, prefix, "METRICS_", metricsOtlp)
		}
	}

	if metricsProm != nil && lang == "go" {
		if metricsProm.Host != "" {
			out = append(out, TelemetryEnvVar{Key: prefix + "OTEL_EXPORTER_PROMETHEUS_HOST", Value: metricsProm.Host,
				Description: "Prometheus metrics exporter host"})
		}
		if metricsProm.Port != 0 {
			out = append(out, TelemetryEnvVar{Key: prefix + "OTEL_EXPORTER_PROMETHEUS_PORT", Value: strconv.Itoa(metricsProm.Port),
				Description: "Prometheus metrics exporter port"})
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
// the given prefix and infix ("" shared, or "TRACES_"/"METRICS_" per-signal).
// A nil exporter or an omitted field contributes nothing so the SDK default
// applies.
func appendOTLPEnv(out []TelemetryEnvVar, prefix, infix string, otlp *schema.TelemetryOTLPExporter) []TelemetryEnvVar {
	if otlp == nil {
		return out
	}
	label := otlpLabel(infix)
	if otlp.Endpoint != "" {
		out = append(out, TelemetryEnvVar{Key: prefix + "OTEL_EXPORTER_OTLP_" + infix + "ENDPOINT", Value: otlp.Endpoint,
			Description: label + " collector endpoint"})
	}
	if otlp.Protocol != "" {
		out = append(out, TelemetryEnvVar{Key: prefix + "OTEL_EXPORTER_OTLP_" + infix + "PROTOCOL", Value: string(otlp.Protocol),
			Description: label + " protocol (`grpc` or `http/protobuf`)"})
	}
	return out
}

// otlpLabel turns the OTLP env infix ("", "TRACES_", "METRICS_") into the
// human-readable signal name used in the README env-var descriptions.
func otlpLabel(infix string) string {
	switch infix {
	case "TRACES_":
		return "OTLP traces"
	case "METRICS_":
		return "OTLP metrics"
	default:
		return "OTLP"
	}
}
