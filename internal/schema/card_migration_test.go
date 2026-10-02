package schema

import (
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

const legacyCardManifest = `
apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: legacy-agent
  description: pre-A2A-v1.0.1 card
  version: 0.1.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    stateTransitionHistory: true
  card:
    protocolVersion: "0.3.0"
    preferredTransport: GRPC
    url: "https://legacy.example.com"
    supportsExtendedAgentCard: true
    security:
      - bearer: []
`

// TestMigrateDeprecatedCardFields_TranslatesLegacyCard pins the translation of
// every card field the A2A v1.0.1 AgentCard replaced, plus the warning raised for
// the capability it dropped outright.
func TestMigrateDeprecatedCardFields_TranslatesLegacyCard(t *testing.T) {
	adl, warnings := migrate(t, legacyCardManifest)

	interfaces := adl.Spec.Card.SupportedInterfaces
	if len(interfaces) != 1 {
		t.Fatalf("expected one derived interface, got %d", len(interfaces))
	}
	if interfaces[0].URL != "https://legacy.example.com" {
		t.Errorf("url not translated, got %q", interfaces[0].URL)
	}
	if interfaces[0].ProtocolBinding != "GRPC" {
		t.Errorf("preferredTransport not translated, got %q", interfaces[0].ProtocolBinding)
	}
	if interfaces[0].ProtocolVersion != "0.3.0" {
		t.Errorf("protocolVersion not translated, got %q", interfaces[0].ProtocolVersion)
	}
	if !adl.Spec.Capabilities.ExtendedAgentCard {
		t.Error("supportsExtendedAgentCard not translated to capabilities.extendedAgentCard")
	}
	if len(adl.Spec.Card.SecurityRequirements) != 1 {
		t.Errorf("security not translated to securityRequirements, got %v", adl.Spec.Card.SecurityRequirements)
	}
	if len(warnings) != 6 {
		t.Fatalf("expected one warning per deprecated field, got %d: %v", len(warnings), warnings)
	}
	if !containsWarning(warnings, "stateTransitionHistory") {
		t.Errorf("expected a stateTransitionHistory warning, got %v", warnings)
	}
}

// TestMigrateDeprecatedCardFields_KeepsExplicitInterfaces makes sure a manifest
// already written in the v1.0.1 shape is never overwritten by a stale mirror field.
func TestMigrateDeprecatedCardFields_KeepsExplicitInterfaces(t *testing.T) {
	adl, warnings := migrate(t, `
apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: mixed-agent
  description: both shapes declared
  version: 0.1.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
  card:
    url: "https://stale.example.com"
    supportedInterfaces:
      - url: "https://current.example.com"
        protocolBinding: JSONRPC
        protocolVersion: "1.0"
`)

	if got := adl.Spec.Card.SupportedInterfaces[0].URL; got != "https://current.example.com" {
		t.Errorf("explicit interface was overwritten, got %q", got)
	}
	if len(warnings) != 1 {
		t.Errorf("expected the deprecated url to still warn, got %v", warnings)
	}
}

// TestMigrateDeprecatedCardFields_ModernManifestIsSilent guards against warning
// noise on manifests that use no removed fields.
func TestMigrateDeprecatedCardFields_ModernManifestIsSilent(t *testing.T) {
	_, warnings := migrate(t, `
apiVersion: adl.inference-gateway.com/v1
kind: Agent
metadata:
  name: modern-agent
  description: v1.0.1 shape only
  version: 0.1.0
spec:
  capabilities:
    streaming: true
    pushNotifications: false
    extendedAgentCard: true
`)

	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func migrate(t *testing.T, manifest string) (*ADL, []string) {
	t.Helper()
	var adl ADL
	if err := yaml.Unmarshal([]byte(manifest), &adl); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return &adl, MigrateDeprecatedCardFields([]byte(manifest), &adl)
}

func containsWarning(warnings []string, substring string) bool {
	for _, warning := range warnings {
		if strings.Contains(warning, substring) {
			return true
		}
	}
	return false
}
