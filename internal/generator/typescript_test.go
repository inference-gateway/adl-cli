package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	for _, want := range []string{
		"import { loadConfig, type Config } from './config.js'",
		"const config = loadConfig();",
		"newLogger(config.server.debug)",
		"loadAgentCard(config)",
		"server.listen(port, host)",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("index.ts missing %q\n%s", want, got)
		}
	}
	for _, banned := range []string{
		"process.env['A2A_SERVER_PORT']",
		"process.env['A2A_AGENT_SYSTEM_PROMPT']",
		"process.env['A2A_AGENT_CLIENT_PROVIDER']",
	} {
		if strings.Contains(got, banned) {
			t.Fatalf("index.ts should no longer read %s directly\n%s", banned, got)
		}
	}
}

// makeTSConfigADL builds a TypeScript ADL with a fully specified agent
// (provider/model/systemPrompt) plus the given custom spec.config sections, for
// exercising the generated config module.
func makeTSConfigADL(sections schema.SpecConfig) *schema.ADL {
	return &schema.ADL{
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
			Agent: &schema.Agent{
				Provider:     "openai",
				Model:        "gpt-4o",
				SystemPrompt: "You are a test bot.",
			},
			Config: sections,
			Language: schema.Language{
				TypeScript: &schema.TypeScriptConfig{
					PackageName: "@example/ts-agent",
					NodeVersion: "24",
				},
			},
		},
	}
}

func TestGenerator_TypeScriptConfig(t *testing.T) {
	t.Run("ADK options come from A2A_-prefixed env vars with manifest defaults", func(t *testing.T) {
		got := renderTS(t, "config.ts", makeTSConfigADL(nil))
		for _, want := range []string{
			"export interface ServerConfig",
			"export interface AgentConfig",
			"export interface LLMConfig",
			"export interface Config",
			"export function loadConfig(): Config",
			`envString('A2A_SERVER_HOST', '0.0.0.0')`,
			`envNumber('A2A_SERVER_PORT', 9090)`,
			`envBool('A2A_SERVER_DEBUG', false)`,
			`envString('A2A_AGENT_NAME', "ts-agent")`,
			`envString('A2A_AGENT_VERSION', "2.1.0")`,
			`envString('A2A_AGENT_SYSTEM_PROMPT', "You are a test bot.")`,
			`envString('A2A_AGENT_CARD_PATH', '.well-known/agent-card.json')`,
			`envString('A2A_SKILLS_DIR', 'skills')`,
			`envString('A2A_AGENT_CLIENT_PROVIDER', "openai")`,
			`envString('A2A_AGENT_CLIENT_MODEL', "gpt-4o")`,
			"process.env['A2A_AGENT_CLIENT_BASE_URL']",
			"${provider.toUpperCase()}_API_KEY",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("config.ts missing %q\n%s", want, got)
			}
		}
		if c := strings.Count(got, "function load"); c != 1 {
			t.Fatalf("expected only loadConfig (1 loader), got %d\n%s", c, got)
		}
		if strings.Contains(got, "function required(") {
			t.Fatalf("required() must be omitted when provider/model have defaults\n%s", got)
		}
	})

	t.Run("a single custom section yields a typed interface + loader under its own prefix", func(t *testing.T) {
		adl := makeTSConfigADL(schema.SpecConfig{
			"database": {
				"connectionString": "postgres://localhost",
				"maxConnections":   10,
				"verbose":          true,
				"ratio":            0.5,
			},
		})
		got := renderTS(t, "config.ts", adl)
		for _, want := range []string{
			"export interface DatabaseConfig {",
			"connectionString: string;",
			"maxConnections: number;",
			"verbose: boolean;",
			"ratio: number;",
			"function loadDatabaseConfig(): DatabaseConfig {",
			`connectionString: envString('DATABASE_CONNECTION_STRING', "postgres://localhost")`,
			`maxConnections: envNumber('DATABASE_MAX_CONNECTIONS', 10)`,
			`verbose: envBool('DATABASE_VERBOSE', true)`,
			`ratio: envNumber('DATABASE_RATIO', 0.5)`,
			"database: DatabaseConfig;",
			"database: loadDatabaseConfig(),",
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("config.ts missing %q\n%s", want, got)
			}
		}
		if c := strings.Count(got, "function load"); c != 2 {
			t.Fatalf("expected loadConfig + loadDatabaseConfig (2 loaders), got %d\n%s", c, got)
		}
	})

	t.Run("multiple sections each get an interface + loader and the reserved tools namespace is skipped", func(t *testing.T) {
		adl := makeTSConfigADL(schema.SpecConfig{
			"tools": {
				"read": map[string]any{"enabled": true},
			},
			"database": {
				"connectionString": "postgres://localhost",
			},
			"notifications": {
				"retryAttempts": 3,
			},
		})
		got := renderTS(t, "config.ts", adl)
		for _, want := range []string{
			"export interface DatabaseConfig {",
			"export interface NotificationsConfig {",
			"function loadDatabaseConfig(): DatabaseConfig {",
			"function loadNotificationsConfig(): NotificationsConfig {",
			"database: DatabaseConfig;",
			"notifications: NotificationsConfig;",
			`retryAttempts: envNumber('NOTIFICATIONS_RETRY_ATTEMPTS', 3)`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("config.ts missing %q\n%s", want, got)
			}
		}
		if strings.Contains(got, "ToolsConfig") || strings.Contains(got, "loadToolsConfig") {
			t.Fatalf("the reserved tools namespace must be skipped, got:\n%s", got)
		}
		if c := strings.Count(got, "function load"); c != 3 {
			t.Fatalf("expected loadConfig + 2 section loaders (3 loaders), got %d\n%s", c, got)
		}
	})

	t.Run("declared defaults and primitive types are honored; quoted numbers stay strings", func(t *testing.T) {
		adl := makeTSConfigADL(schema.SpecConfig{
			"reporting": {
				"outputPath":  "/tmp/reports",
				"maxItems":    "50",
				"maxFileSize": 50,
				"compress":    false,
			},
		})
		got := renderTS(t, "config.ts", adl)
		for _, want := range []string{
			"outputPath: string;",
			"maxItems: string;",
			"maxFileSize: number;",
			"compress: boolean;",
			`outputPath: envString('REPORTING_OUTPUT_PATH', "/tmp/reports")`,
			`maxItems: envString('REPORTING_MAX_ITEMS', "50")`,
			`maxFileSize: envNumber('REPORTING_MAX_FILE_SIZE', 50)`,
			`compress: envBool('REPORTING_COMPRESS', false)`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("config.ts missing %q\n%s", want, got)
			}
		}
	})

	t.Run("required() guards provider/model and the default prompt is baked when the manifest declares no agent", func(t *testing.T) {
		got := renderTS(t, "config.ts", makeTypeScriptADL(nil, ""))
		for _, want := range []string{
			"function required(",
			"const provider = required('A2A_AGENT_CLIENT_PROVIDER');",
			"required('A2A_AGENT_CLIENT_MODEL')",
			`envString('A2A_AGENT_SYSTEM_PROMPT', "You are a helpful AI assistant.")`,
		} {
			if !strings.Contains(got, want) {
				t.Fatalf("config.ts missing %q\n%s", want, got)
			}
		}
	})
}

// TestGenerator_TypeScriptTools runs the full pipeline (manifest -> Generate())
// to lock down issue #167: non-reserved tools render one createXTool file each
// with their injected dependencies, services produce a typed factory module,
// the toolbox aggregator wires both together and is consumed by index.ts,
// reserved tool IDs are skipped (built-in TS tools are a tracked follow-up),
// and tool/service files - but not the regenerated aggregator - land in
// .adl-ignore so user implementations survive re-generation.
func TestGenerator_TypeScriptTools(t *testing.T) {
	adl := &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "ts-tools-agent",
			Description: "A TypeScript agent exercising tools + services",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: false},
			Server:       schema.Server{Port: 8080},
			Agent: &schema.Agent{
				Provider:     "openai",
				Model:        "gpt-4o",
				SystemPrompt: "You are a test bot.",
			},
			Config: schema.SpecConfig{
				"notifications": {
					"retryAttempts": 3,
				},
			},
			Services: schema.SpecServices{
				"database": {
					Interface:   "DatabaseService",
					Factory:     "NewDatabaseService",
					Type:        schema.ServiceTypeService,
					Description: "Database access",
				},
				"notifications": {
					Interface:   "NotificationService",
					Factory:     "NewNotificationService",
					Type:        schema.ServiceTypeClient,
					Description: "Notification dispatch",
				},
			},
			Tools: []schema.Tool{
				// Reserved built-in: must be skipped (no src/tools/read.ts).
				{ID: "read"},
				{
					ID:          "query_database",
					Name:        "query_database",
					Description: "Run a read-only SQL query",
					Tags:        []string{"data"},
					Schema: schema.ToolSchema{
						"type": "object",
						"properties": map[string]any{
							"query": map[string]any{"type": "string"},
						},
						"required": []any{"query"},
					},
					Inject: []string{"logger", "database"},
				},
				{
					ID:          "send_notification",
					Name:        "send_notification",
					Description: "Send a notification to a user",
					Tags:        []string{"notify"},
					Schema: schema.ToolSchema{
						"type": "object",
						"properties": map[string]any{
							"message": map[string]any{"type": "string"},
						},
					},
					Inject: []string{"logger", "notifications", "config.notifications"},
				},
			},
			Language: schema.Language{
				TypeScript: &schema.TypeScriptConfig{
					PackageName: "@example/ts-tools-agent",
					NodeVersion: "24",
				},
			},
		},
	}

	tmpDir, err := os.MkdirTemp("", "adl-ts-tools-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	adlPath := filepath.Join(tmpDir, "agent.yaml")
	writeYAML(t, adlPath, adl)

	outDir := filepath.Join(tmpDir, "out")
	gen := New(Config{Template: "minimal", Overwrite: true, Version: "test"})
	if err := gen.Generate(adlPath, outDir); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	read := func(rel string) string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(outDir, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		return string(b)
	}

	// A non-reserved tool renders a factory with its injected service typed
	// from the service module, plus schema-derived parameters and a TODO body.
	queryDatabase := read("src/tools/query_database.ts")
	for _, want := range []string{
		"export function createQueryDatabaseTool(",
		"logger: Logger,",
		"database: DatabaseService,",
		"import type { DatabaseService } from '../services/database.js';",
		"parameters: {",
		`"query"`,
		"// TODO: implement the `query_database` tool.",
	} {
		if !strings.Contains(queryDatabase, want) {
			t.Errorf("src/tools/query_database.ts missing %q\n---\n%s", want, queryDatabase)
		}
	}

	// A tool injecting both a service AND a config section: the two share the
	// base name `notifications` yet must not collide (service -> `notifications`,
	// section -> `notificationsConfig`).
	sendNotification := read("src/tools/send_notification.ts")
	for _, want := range []string{
		"export function createSendNotificationTool(",
		"notifications: NotificationService,",
		"notificationsConfig: NotificationsConfig,",
	} {
		if !strings.Contains(sendNotification, want) {
			t.Errorf("src/tools/send_notification.ts missing %q\n---\n%s", want, sendNotification)
		}
	}

	// Each spec.services entry produces a typed interface + factory module.
	database := read("src/services/database.ts")
	for _, want := range []string{
		"export interface DatabaseService {",
		"export function newDatabaseService(logger: Logger, config: Config): DatabaseService {",
	} {
		if !strings.Contains(database, want) {
			t.Errorf("src/services/database.ts missing %q\n---\n%s", want, database)
		}
	}

	// The toolbox aggregator constructs each service once and registers every
	// non-reserved tool, threading logger/config/config.<section>/service args.
	aggregator := read("src/tools/index.ts")
	for _, want := range []string{
		"export function buildToolBox(logger: Logger, config: Config): ToolBox",
		"const database = newDatabaseService(logger, config);",
		"const notifications = newNotificationService(logger, config);",
		"createQueryDatabaseTool(",
		"createSendNotificationTool(",
		"config.notifications,",
		"import { createQueryDatabaseTool } from './query_database.js';",
	} {
		if !strings.Contains(aggregator, want) {
			t.Errorf("src/tools/index.ts missing %q\n---\n%s", want, aggregator)
		}
	}
	for _, banned := range []string{"createReadTool", "./read.js"} {
		if strings.Contains(aggregator, banned) {
			t.Errorf("src/tools/index.ts must not reference the reserved read tool (%q)\n---\n%s", banned, aggregator)
		}
	}

	// index.ts consumes the aggregator instead of booting an empty toolbox.
	index := read("src/index.ts")
	for _, want := range []string{
		"import { buildToolBox } from './tools/index.js';",
		"const toolBox = buildToolBox(logger, config);",
	} {
		if !strings.Contains(index, want) {
			t.Errorf("src/index.ts missing %q\n---\n%s", want, index)
		}
	}
	if strings.Contains(index, "new DefaultToolBox()") {
		t.Errorf("src/index.ts should not boot an empty DefaultToolBox when user tools exist\n---\n%s", index)
	}

	// Reserved tool IDs are skipped: no src/tools/read.ts is emitted.
	if _, err := os.Stat(filepath.Join(outDir, "src", "tools", "read.ts")); !os.IsNotExist(err) {
		t.Errorf("src/tools/read.ts must NOT exist (reserved tools deferred), stat err = %v", err)
	}

	// .adl-ignore protects user-owned tool + service implementations but lets
	// the deterministically-regenerated aggregator refresh on every run.
	ignore := read(".adl-ignore")
	for _, want := range []string{
		"src/tools/query_database.ts",
		"src/tools/send_notification.ts",
		"src/services/database.ts",
		"src/services/notifications.ts",
	} {
		if !strings.Contains(ignore, want) {
			t.Errorf(".adl-ignore missing %q\n---\n%s", want, ignore)
		}
	}
	if strings.Contains(ignore, "src/tools/index.ts") {
		t.Errorf(".adl-ignore must NOT list the regenerated aggregator src/tools/index.ts\n---\n%s", ignore)
	}
}
