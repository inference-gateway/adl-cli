package schema

import (
	"cmp"
	"fmt"

	yaml "gopkg.in/yaml.v3"
)

const (
	// DefaultProtocolBinding is the A2A protocol binding assumed for a derived
	// AgentInterface when the manifest does not name one.
	DefaultProtocolBinding = "JSONRPC"

	// DefaultProtocolVersion is the A2A protocol version advertised on a derived
	// AgentInterface when the manifest does not name one.
	DefaultProtocolVersion = "1.0"
)

// legacyCard mirrors the pre-A2A-v1.0.1 spec.card fields that ADL v0.28.0
// removed. They are still read from the raw manifest so existing agents keep
// generating, then folded onto their v1.0.1 replacements.
type legacyCard struct {
	ProtocolVersion           string                         `yaml:"protocolVersion"`
	URL                       string                         `yaml:"url"`
	PreferredTransport        string                         `yaml:"preferredTransport"`
	SupportsExtendedAgentCard bool                           `yaml:"supportsExtendedAgentCard"`
	Security                  []CardSecurityRequirementsElem `yaml:"security"`
}

// legacyManifest is the sliver of a manifest holding the removed card and
// capability fields.
type legacyManifest struct {
	Spec struct {
		Card         legacyCard `yaml:"card"`
		Capabilities struct {
			StateTransitionHistory *bool `yaml:"stateTransitionHistory"`
		} `yaml:"capabilities"`
	} `yaml:"spec"`
}

// MigrateDeprecatedCardFields translates the card and capability fields ADL
// v0.28.0 removed onto their A2A v1.0.1 replacements so pre-v1.0.1 manifests
// still emit a conformant AgentCard. Fields already written in the v1.0.1 shape
// win; every translation returns a deprecation warning for the caller to surface.
func MigrateDeprecatedCardFields(data []byte, adl *ADL) []string {
	var legacy legacyManifest
	if err := yaml.Unmarshal(data, &legacy); err != nil {
		return nil
	}

	warnings := migrateSupportedInterfaces(legacy.Spec.Card, adl)
	warnings = append(warnings, migrateSecurityRequirements(legacy.Spec.Card, adl)...)
	warnings = append(warnings, migrateExtendedAgentCard(legacy.Spec.Card, adl)...)
	if legacy.Spec.Capabilities.StateTransitionHistory != nil {
		warnings = append(warnings, "spec.capabilities.stateTransitionHistory was removed in ADL v0.28.0 (the A2A v1.0.1 AgentCard dropped it); it is ignored when generating the agent card.")
	}
	return warnings
}

// migrateSupportedInterfaces folds spec.card.{url,preferredTransport,protocolVersion}
// into a single preferred spec.card.supportedInterfaces entry.
func migrateSupportedInterfaces(legacy legacyCard, adl *ADL) []string {
	var warnings []string
	if legacy.URL != "" {
		warnings = append(warnings, deprecationWarning("spec.card.url", "spec.card.supportedInterfaces[0].url"))
	}
	if legacy.PreferredTransport != "" {
		warnings = append(warnings, deprecationWarning("spec.card.preferredTransport", "spec.card.supportedInterfaces[0].protocolBinding"))
	}
	if legacy.ProtocolVersion != "" {
		warnings = append(warnings, deprecationWarning("spec.card.protocolVersion", "spec.card.supportedInterfaces[0].protocolVersion"))
	}
	if len(warnings) == 0 || len(declaredInterfaces(adl)) > 0 {
		return warnings
	}

	ensureCard(adl).SupportedInterfaces = []AgentInterface{{
		URL:             legacy.URL,
		ProtocolBinding: cmp.Or(legacy.PreferredTransport, DefaultProtocolBinding),
		ProtocolVersion: cmp.Or(legacy.ProtocolVersion, DefaultProtocolVersion),
	}}
	return warnings
}

// migrateSecurityRequirements renames spec.card.security, which A2A v1.0.1 calls
// securityRequirements.
func migrateSecurityRequirements(legacy legacyCard, adl *ADL) []string {
	if len(legacy.Security) == 0 {
		return nil
	}
	if card := ensureCard(adl); len(card.SecurityRequirements) == 0 {
		card.SecurityRequirements = legacy.Security
	}
	return []string{deprecationWarning("spec.card.security", "spec.card.securityRequirements")}
}

// migrateExtendedAgentCard moves the extended-card flag off the card, where A2A
// v1.0.1 keeps it on the capability set.
func migrateExtendedAgentCard(legacy legacyCard, adl *ADL) []string {
	if !legacy.SupportsExtendedAgentCard {
		return nil
	}
	adl.Spec.Capabilities.ExtendedAgentCard = true
	return []string{deprecationWarning("spec.card.supportsExtendedAgentCard", "spec.capabilities.extendedAgentCard")}
}

func declaredInterfaces(adl *ADL) []AgentInterface {
	if adl.Spec.Card == nil {
		return nil
	}
	return adl.Spec.Card.SupportedInterfaces
}

func deprecationWarning(field, replacement string) string {
	return fmt.Sprintf("%s was removed in ADL v0.28.0 (A2A v1.0.1); it was translated to %s - update the manifest to silence this warning.", field, replacement)
}

func ensureCard(adl *ADL) *Card {
	if adl.Spec.Card == nil {
		adl.Spec.Card = &Card{}
	}
	return adl.Spec.Card
}
