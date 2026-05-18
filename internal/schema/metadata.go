package schema

import "time"

// GeneratedMetadata describes a single generated-file header — when it was
// produced, which CLI version produced it, which template, and which ADL
// source file. It is not part of the ADL specification and therefore lives
// outside the generated types in types.go.
type GeneratedMetadata struct {
	GeneratedAt time.Time `json:"generatedAt"`
	CLIVersion  string    `json:"cliVersion"`
	Template    string    `json:"template"`
	ADLFile     string    `json:"adlFile"`
}
