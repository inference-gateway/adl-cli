package templates

import (
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// TestTelemetryEnabled covers the three states of spec.telemetry: absent
// (nil block), explicitly disabled, and enabled. Telemetry is off by default,
// so only the last case reports true.
func TestTelemetryEnabled(t *testing.T) {
	cases := []struct {
		name string
		tel  *schema.TelemetryConfig
		want bool
	}{
		{"nil block", nil, false},
		{"disabled", &schema.TelemetryConfig{Enabled: false}, false},
		{"enabled", &schema.TelemetryConfig{Enabled: true}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adl := minimalGoADL()
			adl.Spec.Telemetry = tc.tel
			if got := telemetryEnabled(adl); got != tc.want {
				t.Errorf("telemetryEnabled = %v, want %v", got, tc.want)
			}
		})
	}
}
