package generator

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/inference-gateway/adl-cli/internal/schema"
	"github.com/inference-gateway/adl-cli/internal/templates"
	"github.com/inference-gateway/adl-cli/internal/vendor"
)

// makeTypeScriptADL builds a minimal but representative TypeScript ADL for
// exercising the core scaffolding templates. Callers tweak the vendor block
// and system prompt via the closure arguments.
func makeTypeScriptADL(v *schema.VendorConfig, systemPrompt string) *schema.ADL {
	adl := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "ts-agent",
			Description: "A TypeScript test agent",
			Version:     "2.1.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: false},
			Server:       schema.Server{Port: 9090},
			Language: schema.Language{
				TypeScript: &schema.TypeScriptConfig{
					PackageName: "@example/ts-agent",
					NodeVersion: "24",
					Vendor:      v,
				},
			},
		},
	}
	if systemPrompt != "" {
		adl.Spec.Agent = &schema.Agent{SystemPrompt: systemPrompt}
	}
	return adl
}

// renderTS renders a single TypeScript template through the registry with a
// fully resolved vendor view, mirroring the production generation path for
// the language-level templates (index.ts/package.json/tsconfig.json/logger.ts).
func renderTS(t *testing.T, tmplKey string, adl *schema.ADL) string {
	t.Helper()
	registry, err := templates.NewRegistry("typescript")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	engine := templates.NewWithRegistry("", registry)
	view, err := vendor.ResolveADL(adl)
	if err != nil {
		t.Fatalf("vendor.ResolveADL: %v", err)
	}
	out, err := engine.ExecuteTemplate(tmplKey, templates.Context{
		ADL:      adl,
		Language: "typescript",
		Vendor:   view,
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate %s: %v", tmplKey, err)
	}
	return out
}

func TestGenerator_TypeScriptPackageJSON(t *testing.T) {
	t.Run("name, version, engines, and pinned ADK dependency are wired from the manifest", func(t *testing.T) {
		got := renderTS(t, "package.json", makeTypeScriptADL(nil, ""))

		var pkg map[string]any
		if err := json.Unmarshal([]byte(got), &pkg); err != nil {
			t.Fatalf("package.json is not valid JSON: %v\n%s", err, got)
		}
		if pkg["name"] != "@example/ts-agent" {
			t.Fatalf("expected name from packageName, got %q\n%s", pkg["name"], got)
		}
		if pkg["version"] != "2.1.0" {
			t.Fatalf("expected version from metadata.version, got %q", pkg["version"])
		}
		if pkg["type"] != "module" {
			t.Fatalf("expected ESM type=module, got %q", pkg["type"])
		}
		engines, _ := pkg["engines"].(map[string]any)
		if engines["node"] != ">=24" {
			t.Fatalf("expected engines.node from nodeVersion, got %q", engines["node"])
		}
		deps, _ := pkg["dependencies"].(map[string]any)
		if deps["@inference-gateway/adk"] != "0.11.0" {
			t.Fatalf("expected pinned ADK dependency, got %q\n%s", deps["@inference-gateway/adk"], got)
		}
		scripts, _ := pkg["scripts"].(map[string]any)
		for _, s := range []string{"build", "start", "typecheck", "test"} {
			if _, ok := scripts[s]; !ok {
				t.Fatalf("expected %q script, got %v", s, scripts)
			}
		}
	})

	t.Run("vendor deps and devdeps merge into the right sections and stay valid JSON", func(t *testing.T) {
		got := renderTS(t, "package.json", makeTypeScriptADL(&schema.VendorConfig{
			Deps:    []string{"zod@^3.23.0"},
			Devdeps: []string{"vitest@^2.0.0", "@types/ws@^8.5.0"},
		}, ""))

		var pkg map[string]any
		if err := json.Unmarshal([]byte(got), &pkg); err != nil {
			t.Fatalf("package.json with vendor entries is not valid JSON: %v\n%s", err, got)
		}
		deps, _ := pkg["dependencies"].(map[string]any)
		if deps["zod"] != "^3.23.0" {
			t.Fatalf("expected zod in dependencies, got %v\n%s", deps, got)
		}
		if _, ok := deps["vitest"]; ok {
			t.Fatalf("devdep vitest leaked into dependencies\n%s", got)
		}
		devDeps, _ := pkg["devDependencies"].(map[string]any)
		if devDeps["vitest"] != "^2.0.0" || devDeps["@types/ws"] != "^8.5.0" {
			t.Fatalf("expected vendor devdeps in devDependencies, got %v\n%s", devDeps, got)
		}
		if devDeps["typescript"] == nil || devDeps["tsx"] == nil || devDeps["@types/node"] == nil {
			t.Fatalf("expected built-in devDeps preserved alongside vendor, got %v", devDeps)
		}
	})
}

func TestGenerator_TypeScriptTSConfig(t *testing.T) {
	got := renderTS(t, "tsconfig.json", makeTypeScriptADL(nil, ""))

	var cfg map[string]any
	if err := json.Unmarshal([]byte(got), &cfg); err != nil {
		t.Fatalf("tsconfig.json is not valid JSON: %v\n%s", err, got)
	}
	opts, _ := cfg["compilerOptions"].(map[string]any)
	if opts == nil {
		t.Fatalf("expected compilerOptions, got\n%s", got)
	}
	boolFlags := []string{
		"strict",
		"verbatimModuleSyntax",
		"noUncheckedIndexedAccess",
		"exactOptionalPropertyTypes",
		"isolatedModules",
	}
	for _, f := range boolFlags {
		if v, ok := opts[f].(bool); !ok || !v {
			t.Fatalf("expected compilerOptions.%s=true, got %v\n%s", f, opts[f], got)
		}
	}
	if !strings.EqualFold(opts["target"].(string), "es2024") {
		t.Fatalf("expected target es2024, got %q", opts["target"])
	}
	if !strings.EqualFold(opts["module"].(string), "nodenext") {
		t.Fatalf("expected module nodenext, got %q", opts["module"])
	}
}

func TestGenerator_TypeScriptLogger(t *testing.T) {
	got := renderTS(t, "logger.ts", makeTypeScriptADL(nil, ""))

	for _, want := range []string{
		"from '@inference-gateway/adk'",
		"createLogger",
		"export function newLogger(debug: boolean)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("logger.ts missing %q\n%s", want, got)
		}
	}
}

func TestGenerator_TypeScriptIndex(t *testing.T) {
	got := renderTS(t, "index.ts", makeTypeScriptADL(nil, "You are a test bot."))

	// A2A server wiring and the three JSON-RPC handlers the issue requires.
	mustContain := []string{
		"createA2AServer({ card })",
		"createMessageSendHandler({ storage })",
		"createTaskGetHandler({ storage })",
		"createTaskListHandler({ storage })",
		"new InMemoryTaskStorage()",
		"new DefaultBackgroundTaskHandler(",
		"new OpenAICompatibleLLMClient(",
		"new AgentBuilder()",
		".withSystemPrompt(systemPrompt)",
		// background worker loop + dead-lettering
		"async function runWorker(",
		"storage.dequeue(signal)",
		"storage.storeDeadLetter(",
		// skills manifest loading (frontmatter-only, Read-on-demand)
		"function loadSkillsManifest(",
		"AVAILABLE SKILLS:",
		// agent card loaded from the generated file
		".well-known/agent-card.json",
		// graceful shutdown
		"process.once('SIGINT', shutdown)",
		"process.once('SIGTERM', shutdown)",
		// the LLM client bridge carried from the example
		"function adaptLLMClient(",
		"import { newLogger } from './logger.js'",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("index.ts missing %q\n%s", want, got)
		}
	}

	// Manifest-driven values land as runtime-overridable defaults.
	if !strings.Contains(got, `'9090'`) {
		t.Fatalf("expected server port 9090 default, got:\n%s", got)
	}
	if !strings.Contains(got, `"ts-agent"`) {
		t.Fatalf("expected agent name default from metadata, got:\n%s", got)
	}
	if !strings.Contains(got, "You are a test bot.") {
		t.Fatalf("expected system prompt from spec.agent, got:\n%s", got)
	}
}

func TestGenerator_TypeScriptIndexDefaultSystemPrompt(t *testing.T) {
	got := renderTS(t, "index.ts", makeTypeScriptADL(nil, ""))
	if !strings.Contains(got, "You are a helpful AI assistant.") {
		t.Fatalf("expected fallback system prompt when spec.agent is absent, got:\n%s", got)
	}
}
