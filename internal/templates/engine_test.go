package templates

import (
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
					Language: &schema.Language{},
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
					Language: &schema.Language{},
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
