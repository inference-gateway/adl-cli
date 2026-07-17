package templates

import (
	"strings"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

func TestToPascalCaseWithAcronyms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		acronyms map[string]string
		expected string
	}{
		{
			name:     "no acronyms",
			input:    "hello_world",
			acronyms: make(map[string]string),
			expected: "HelloWorld",
		},
		{
			name:     "single acronym",
			input:    "get_api_data",
			acronyms: map[string]string{"api": "API"},
			expected: "GetAPIData",
		},
		{
			name:     "multiple acronyms",
			input:    "process_json_from_http_api",
			acronyms: map[string]string{"json": "JSON", "http": "HTTP", "api": "API"},
			expected: "ProcessJSONFromHTTPAPI",
		},
		{
			name:     "custom acronym n8n",
			input:    "get_n8n_docs",
			acronyms: map[string]string{"n8n": "N8N"},
			expected: "GetN8NDocs",
		},
		{
			name:     "mixed case input",
			input:    "Get_N8n_Docs",
			acronyms: map[string]string{"n8n": "N8N"},
			expected: "GetN8NDocs",
		},
		{
			name:     "dash separated",
			input:    "get-n8n-data",
			acronyms: map[string]string{"n8n": "N8N"},
			expected: "GetN8NData",
		},
		{
			name:     "empty string",
			input:    "",
			acronyms: map[string]string{"api": "API"},
			expected: "",
		},
		{
			name:     "single word with acronym",
			input:    "api",
			acronyms: map[string]string{"api": "API"},
			expected: "API",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toPascalCaseWithAcronyms(tt.input, tt.acronyms)
			if result != tt.expected {
				t.Errorf("toPascalCaseWithAcronyms(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase_DefaultAcronyms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "id acronym",
			input:    "user_id",
			expected: "UserID",
		},
		{
			name:     "api acronym",
			input:    "get_api_data",
			expected: "GetAPIData",
		},
		{
			name:     "url acronym",
			input:    "base_url",
			expected: "BaseURL",
		},
		{
			name:     "json acronym",
			input:    "parse_json_response",
			expected: "ParseJSONResponse",
		},
		{
			name:     "multiple default acronyms",
			input:    "http_api_json_url",
			expected: "HTTPAPIJSONURL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple case",
			input:    "hello_world",
			expected: "helloWorld",
		},
		{
			name:     "with acronym",
			input:    "get_api_data",
			expected: "getAPIData",
		},
		{
			name:     "single word",
			input:    "test",
			expected: "test",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildAcronymsMap(t *testing.T) {
	tests := []struct {
		name           string
		customAcronyms []string
		expectedKeys   []string
		expectedValues []string
	}{
		{
			name:           "empty custom acronyms",
			customAcronyms: []string{},
			expectedKeys:   []string{"id", "api", "url"},
			expectedValues: []string{"ID", "API", "URL"},
		},
		{
			name:           "single custom acronym",
			customAcronyms: []string{"n8n"},
			expectedKeys:   []string{"id", "api", "n8n"},
			expectedValues: []string{"ID", "API", "N8N"},
		},
		{
			name:           "multiple custom acronyms",
			customAcronyms: []string{"n8n", "xyz", "abc"},
			expectedKeys:   []string{"n8n", "xyz", "abc"},
			expectedValues: []string{"N8N", "XYZ", "ABC"},
		},
		{
			name:           "override default acronym",
			customAcronyms: []string{"id"},
			expectedKeys:   []string{"id"},
			expectedValues: []string{"ID"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAcronymsMap(tt.customAcronyms)

			for i, key := range tt.expectedKeys {
				expectedValue := tt.expectedValues[i]
				if value, exists := result[key]; !exists {
					t.Errorf("buildAcronymsMap() missing key %q", key)
				} else if value != expectedValue {
					t.Errorf("buildAcronymsMap() key %q = %q, want %q", key, value, expectedValue)
				}
			}
		})
	}
}

func TestEngine_PrepareContext(t *testing.T) {
	tests := []struct {
		name     string
		adl      *schema.ADL
		expected map[string]string
	}{
		{
			name: "no custom acronyms",
			adl: &schema.ADL{
				Spec: schema.Spec{
					Language: schema.Language{},
				},
			},
			expected: getDefaultAcronyms(),
		},
		{
			name: "with custom acronyms",
			adl: &schema.ADL{
				Spec: schema.Spec{
					Acronyms: []string{"n8n", "xyz"},
				},
			},
			expected: func() map[string]string {
				result := getDefaultAcronyms()
				result["n8n"] = "N8N"
				result["xyz"] = "XYZ"
				return result
			}(),
		},
		{
			name: "nil language",
			adl: &schema.ADL{
				Spec: schema.Spec{},
			},
			expected: getDefaultAcronyms(),
		},
		{
			name:     "nil adl",
			adl:      nil,
			expected: getDefaultAcronyms(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New("test")
			ctx := Context{ADL: tt.adl}

			preparedCtx := engine.prepareContext(ctx)

			for key, expectedValue := range tt.expected {
				if value, exists := preparedCtx.customAcronyms[key]; !exists {
					t.Errorf("prepareContext() missing acronym %q", key)
				} else if value != expectedValue {
					t.Errorf("prepareContext() acronym %q = %q, want %q", key, value, expectedValue)
				}
			}
		})
	}
}

func TestEngine_Execute_WithCustomAcronyms(t *testing.T) {
	tests := []struct {
		name     string
		template string
		adl      *schema.ADL
		expected string
	}{
		{
			name:     "template with toPascalCase and custom acronym",
			template: `{{ "get_n8n_docs" | toPascalCase }}`,
			adl: &schema.ADL{
				Spec: schema.Spec{
					Acronyms: []string{"n8n"},
				},
			},
			expected: "GetN8NDocs\n",
		},
		{
			name:     "template with toCamelCase and custom acronym",
			template: `{{ "get_n8n_docs" | toCamelCase }}`,
			adl: &schema.ADL{
				Spec: schema.Spec{
					Acronyms: []string{"n8n"},
				},
			},
			expected: "getN8NDocs\n",
		},
		{
			name:     "template with default acronyms only",
			template: `{{ "get_api_data" | toPascalCase }}`,
			adl: &schema.ADL{
				Spec: schema.Spec{
					Language: schema.Language{},
				},
			},
			expected: "GetAPIData\n",
		},
		{
			name:     "template with mixed default and custom acronyms",
			template: `{{ "process_n8n_api_data" | toPascalCase }}`,
			adl: &schema.ADL{
				Spec: schema.Spec{
					Acronyms: []string{"n8n"},
				},
			},
			expected: "ProcessN8NAPIData\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := New("test")
			ctx := Context{ADL: tt.adl}

			result, err := engine.Execute(tt.template, ctx)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Execute() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestToTitleCaseWithAcronyms(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		acronyms []string
		expected string
	}{
		{
			name:     "lowercase brand acronym preserved",
			input:    "n8n-agent",
			acronyms: []string{"n8n"},
			expected: "n8n Agent",
		},
		{
			name:     "uppercase acronym preserved",
			input:    "api-server",
			acronyms: []string{"API"},
			expected: "API Server",
		},
		{
			name:     "mixed-case brand acronym preserved",
			input:    "postgresql-driver",
			acronyms: []string{"PostgreSQL"},
			expected: "PostgreSQL Driver",
		},
		{
			name:     "multiple acronyms in one name",
			input:    "n8n-graphql-bridge",
			acronyms: []string{"n8n", "GraphQL"},
			expected: "n8n GraphQL Bridge",
		},
		{
			name:     "snake_case input",
			input:    "weather_agent",
			acronyms: nil,
			expected: "Weather Agent",
		},
		{
			name:     "no acronyms",
			input:    "weather-agent",
			acronyms: nil,
			expected: "Weather Agent",
		},
		{
			name:     "empty string",
			input:    "",
			acronyms: []string{"n8n"},
			expected: "",
		},
		{
			name:     "single word non-acronym",
			input:    "agent",
			acronyms: []string{"n8n"},
			expected: "Agent",
		},
		{
			name:     "acronym match is case-insensitive against input",
			input:    "N8N-agent",
			acronyms: []string{"n8n"},
			expected: "n8n Agent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toTitleCaseWithAcronyms(tt.input, tt.acronyms)
			if got != tt.expected {
				t.Errorf("toTitleCaseWithAcronyms(%q, %v) = %q, want %q", tt.input, tt.acronyms, got, tt.expected)
			}
		})
	}
}

func TestREADMETemplate_TitleIsHumanized(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	tests := []struct {
		name      string
		metadata  schema.Metadata
		acronyms  []string
		wantTitle string
	}{
		{
			name:      "lowercase brand acronym preserved in title",
			metadata:  schema.Metadata{Name: "n8n-agent", Version: "0.1.0"},
			acronyms:  []string{"n8n"},
			wantTitle: "# n8n Agent",
		},
		{
			name:      "no acronyms renders standard title case",
			metadata:  schema.Metadata{Name: "weather-agent", Version: "0.1.0"},
			wantTitle: "# Weather Agent",
		},
		{
			name:      "mixed-case brand acronym preserved",
			metadata:  schema.Metadata{Name: "postgresql-driver", Version: "0.1.0"},
			acronyms:  []string{"PostgreSQL"},
			wantTitle: "# PostgreSQL Driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adl := &schema.ADL{
				Metadata: tt.metadata,
				Spec: schema.Spec{
					Acronyms: tt.acronyms,
					Language: schema.Language{
						Go: &schema.GoConfig{Module: "example.com/x", Version: "1.26.4"},
					},
				},
			}
			ctx := Context{ADL: adl, Language: "go"}

			out, err := engine.ExecuteTemplate("docs/README.md", ctx)
			if err != nil {
				t.Fatalf("ExecuteTemplate(README.md): %v", err)
			}

			lines := strings.Split(out, "\n")
			var got string
			for _, line := range lines {
				if strings.HasPrefix(line, "# ") {
					got = line
					break
				}
			}
			if got != tt.wantTitle {
				t.Errorf("README title = %q, want %q", got, tt.wantTitle)
			}
		})
	}
}

// TestREADMETemplate_LinksToConfigurations verifies the README delegates the
// env-var reference to CONFIGURATIONS.md instead of inlining the A2A_* table.
func TestREADMETemplate_LinksToConfigurations(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	out, err := engine.ExecuteTemplate("docs/README.md", Context{ADL: minimalGoADL(), Language: "go"})
	if err != nil {
		t.Fatalf("ExecuteTemplate(README.md): %v", err)
	}

	if !strings.Contains(out, "[CONFIGURATIONS.md](CONFIGURATIONS.md)") {
		t.Error("README missing link to CONFIGURATIONS.md")
	}
	if strings.Contains(out, "A2A_AGENT_CLIENT_PROVIDER") {
		t.Error("README still inlines the A2A_* env var table")
	}
}

// TestCONFIGURATIONSTemplate_TelemetryEnvVars guards the regression from
// inference-gateway/mock-agent#64: once telemetry moved from spec.config.telemetry
// (which the "Custom Configuration" loop documented) to top-level spec.telemetry,
// the docs stopped documenting any telemetry env vars even though .env.example
// still emits them. The env-var table (now in CONFIGURATIONS.md) must carry the
// language-specific master switch plus every telemetryEnvVars entry, and stay
// empty when telemetry is off.
func TestCONFIGURATIONSTemplate_TelemetryEnvVars(t *testing.T) {
	enabledGo := &schema.TelemetryConfig{
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

	tests := []struct {
		name       string
		lang       string
		adl        func() *schema.ADL
		wantRows   []string
		wantAbsent []string
	}{
		{
			name: "go enabled documents A2A_-prefixed vars",
			lang: "go",
			adl: func() *schema.ADL {
				adl := minimalGoADL()
				adl.Spec.Telemetry = enabledGo
				return adl
			},
			wantRows: []string{
				"| **Telemetry** | `A2A_TELEMETRY_ENABLE` | Enable OpenTelemetry instrumentation | `true` |",
				"| **Telemetry** | `A2A_OTEL_TRACES_EXPORTER` |",
				"| **Telemetry** | `A2A_OTEL_EXPORTER_OTLP_ENDPOINT` |",
				"| **Telemetry** | `A2A_OTEL_EXPORTER_PROMETHEUS_PORT` |",
			},
		},
		{
			name: "typescript enabled documents A2A_-prefixed vars",
			lang: "typescript",
			adl: func() *schema.ADL {
				adl := minimalTypeScriptADL()
				adl.Spec.Telemetry = &schema.TelemetryConfig{
					Enabled: true,
					Traces: &schema.TelemetryTracesConfig{
						Exporter: &schema.TelemetryTracesExporter{
							Otlp: otlpExporter("http://localhost:4318", schema.TelemetryOTLPExporterProtocolHttpProtobuf),
						},
					},
				}
				return adl
			},
			wantRows: []string{
				"| **Telemetry** | `A2A_TELEMETRY_ENABLE` | Enable OpenTelemetry instrumentation | `true` |",
				"| **Telemetry** | `A2A_OTEL_TRACES_EXPORTER` |",
			},
			wantAbsent: []string{"| `TELEMETRY_ENABLE`", "| `OTEL_TRACES_EXPORTER`"},
		},
		{
			name: "go disabled documents no telemetry rows",
			lang: "go",
			adl: func() *schema.ADL {
				adl := minimalGoADL()
				adl.Spec.Telemetry = &schema.TelemetryConfig{Enabled: false}
				return adl
			},
			wantAbsent: []string{"**Telemetry**", "TELEMETRY_ENABLE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry, err := NewRegistry(tt.lang)
			if err != nil {
				t.Fatalf("NewRegistry: %v", err)
			}
			engine := NewWithRegistry("minimal", registry)

			out, err := engine.ExecuteTemplate("docs/CONFIGURATIONS.md", Context{ADL: tt.adl(), Language: tt.lang})
			if err != nil {
				t.Fatalf("ExecuteTemplate(CONFIGURATIONS.md): %v", err)
			}

			for _, row := range tt.wantRows {
				if !strings.Contains(out, row) {
					t.Errorf("CONFIGURATIONS.md missing telemetry row %q", row)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(out, absent) {
					t.Errorf("CONFIGURATIONS.md unexpectedly contains %q", absent)
				}
			}
		})
	}
}

func TestGetDefaultAcronyms(t *testing.T) {
	acronyms := getDefaultAcronyms()

	expectedDefaults := map[string]string{
		"id":   "ID",
		"api":  "API",
		"url":  "URL",
		"json": "JSON",
		"sql":  "SQL",
		"html": "HTML",
	}

	for key, expectedValue := range expectedDefaults {
		if value, exists := acronyms[key]; !exists {
			t.Errorf("getDefaultAcronyms() missing default acronym %q", key)
		} else if value != expectedValue {
			t.Errorf("getDefaultAcronyms() acronym %q = %q, want %q", key, value, expectedValue)
		}
	}

	if len(acronyms) < 10 {
		t.Errorf("getDefaultAcronyms() returned %d acronyms, expected at least 10", len(acronyms))
	}
}

func TestSkillScaffoldTemplate_License(t *testing.T) {
	registry, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := NewWithRegistry("minimal", registry)

	tests := []struct {
		name        string
		skill       map[string]any
		wantSubstr  string
		notExpected string
	}{
		{
			name: "with SPDX license renders license frontmatter",
			skill: map[string]any{
				"ID":          "company-policy",
				"Name":        "company-policy",
				"Description": "Internal rules",
				"Tags":        []string{"policy"},
				"License":     "Apache-2.0",
			},
			wantSubstr: "\nlicense: Apache-2.0\n",
		},
		{
			name: "without license omits the field entirely",
			skill: map[string]any{
				"ID":          "company-policy",
				"Name":        "company-policy",
				"Description": "Internal rules",
				"Tags":        []string{"policy"},
				"License":     "",
			},
			notExpected: "license:",
		},
		{
			name: "Proprietary license is rendered verbatim",
			skill: map[string]any{
				"ID":          "internal-runbook",
				"Name":        "internal-runbook",
				"Description": "Closed-source runbook",
				"License":     "Proprietary",
			},
			wantSubstr: "\nlicense: Proprietary\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := engine.ExecuteToolTemplateWithContext("skills/skill.md", tc.skill, Context{
				ADL: &schema.ADL{},
			})
			if err != nil {
				t.Fatalf("ExecuteToolTemplateWithContext: %v", err)
			}
			if tc.wantSubstr != "" && !strings.Contains(out, tc.wantSubstr) {
				t.Errorf("rendered output missing %q\n---\n%s\n---", tc.wantSubstr, out)
			}
			if tc.notExpected != "" && strings.Contains(out, tc.notExpected) {
				t.Errorf("rendered output unexpectedly contains %q\n---\n%s\n---", tc.notExpected, out)
			}
		})
	}
}
