package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/viper"

	"github.com/inference-gateway/adl-cli/internal/tui"
)

// runInitInteractive drives the branded huh wizard. It is only reached on a real
// TTY (see runInit); CI, piped, and --defaults invocations use the plain flow.
func runInitInteractive(args []string) error {
	tui.PrintBanner()

	projectDir := wizardProjectDir()

	projectName := resolveProjectName(args, projectDir)

	if projectDir != "." && projectDir != "" {
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
	}

	adlFile := filepath.Join(projectDir, "agent.yaml")

	adl, err := wizardExistingADL()
	if err != nil {
		return err
	}
	if adl == nil {
		ans := collectAnswersWizard(projectName)
		adl = buildADL(ans)
	}

	if err := writeADLFile(adl, adlFile); err != nil {
		return fmt.Errorf("failed to write ADL file: %w", err)
	}

	printInitSummary(adl, adlFile, projectDir)
	return nil
}

// resolveProjectName mirrors the project-name resolution used by the plain flow:
// an explicit arg wins, then the git remote, then the directory name.
func resolveProjectName(args []string, projectDir string) string {
	if len(args) > 0 {
		return args[0]
	}
	if name := getProjectNameFromGit(); name != "" {
		return name
	}
	if projectDir == "." {
		cwd, _ := os.Getwd()
		return filepath.Base(cwd)
	}
	return filepath.Base(projectDir)
}

func wizardProjectDir() string {
	if v, locked := wzString("path", "."); locked {
		return v
	}
	projectDir := "."
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Project directory").
			Description("Where to write agent.yaml. Use \".\" for the current directory.").
			Placeholder(".").
			Value(&projectDir),
	))
	if err := tui.RunForm(form); err != nil {
		return "."
	}
	projectDir = strings.TrimSpace(projectDir)
	if projectDir == "" {
		return "."
	}
	return projectDir
}

// wizardExistingADL offers to import a pre-written agent.yaml. It returns a
// non-nil *adlData when the user chose that path, or (nil, nil) to fall through
// to the wizard.
func wizardExistingADL() (*adlData, error) {
	var useExisting bool
	if err := tui.RunForm(huh.NewForm(huh.NewGroup(
		leftConfirm().
			Title("Start from an existing ADL file?").
			Description("Import a pre-written agent.yaml instead of answering the wizard.").
			Value(&useExisting),
	))); err != nil {
		return nil, err
	}
	if !useExisting {
		return nil, nil
	}

	var path string
	if err := tui.RunForm(huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Path to existing ADL file").
			Description("Relative or absolute path to a valid agent.yaml.").
			Value(&path).
			Validate(func(s string) error {
				s = strings.TrimSpace(s)
				if s == "" {
					return errors.New("a path is required")
				}
				if _, err := os.Stat(absPath(s)); os.IsNotExist(err) {
					return errors.New("file does not exist")
				}
				if _, err := readADLFile(absPath(s)); err != nil {
					return fmt.Errorf("not a valid ADL file: %w", err)
				}
				return nil
			}),
	))); err != nil {
		return nil, err
	}

	adl, err := readADLFile(absPath(strings.TrimSpace(path)))
	if err != nil {
		return nil, fmt.Errorf("failed to read ADL file: %w", err)
	}
	fmt.Printf("\n%s\n", tui.SummaryBox("Imported existing ADL schema", []tui.Field{
		{Label: "Source", Value: strings.TrimSpace(path)},
	}))
	return adl, nil
}

// collectAnswersWizard gathers every manifest field through grouped huh forms.
// Flags/env values set via viper take precedence and are not re-prompted, so the
// documented "flag/env > prompt > default" order is preserved.
func collectAnswersWizard(projectName string) answers {
	var ans answers

	// ---- metadata + agent type ----
	name, nameLocked := wzString("name", projectName)
	description, descLocked := wzString("description", "A helpful AI agent")
	version, versionLocked := wzString("version", "0.1.0")
	agentType, typeLocked := wzString("type", "ai-powered")

	var metaFields []huh.Field
	if !nameLocked {
		metaFields = append(metaFields, huh.NewInput().
			Title("Agent name").
			Description("A short, memorable identifier for your agent.").
			Value(&name).
			Validate(requireNonEmpty))
	}
	if !descLocked {
		metaFields = append(metaFields, huh.NewInput().
			Title("Description").
			Description("One line describing what your agent does.").
			Value(&description))
	}
	if !versionLocked {
		metaFields = append(metaFields, huh.NewInput().
			Title("Version").
			Value(&version).
			Validate(requireNonEmpty))
	}
	if !typeLocked {
		metaFields = append(metaFields, huh.NewSelect[string]().
			Title("Agent type").
			Description("AI-powered agents call an LLM; minimal agents are plain handlers.").
			Options(
				huh.NewOption("AI-powered", "ai-powered"),
				huh.NewOption("Minimal", "minimal"),
			).
			Value(&agentType))
	}
	runFields(metaFields)

	ans.Name = strings.TrimSpace(name)
	ans.Description = description
	ans.Version = version
	ans.AgentType = agentType

	// ---- AI configuration ----
	if ans.AgentType == "ai-powered" {
		provider, providerLocked := wzString("provider", "")
		model, modelLocked := wzString("model", "")
		systemPrompt, promptLocked := wzString("system-prompt", "You are a helpful AI assistant.")
		var maxTokensStr, temperatureStr string

		var aiFields []huh.Field
		if !providerLocked {
			aiFields = append(aiFields, huh.NewSelect[string]().
				Title("LLM provider").
				Description("You can leave the model blank to stay vendor-neutral.").
				Options(huh.NewOptions(aiProviders...)...).
				Value(&provider))
		}
		if !modelLocked {
			aiFields = append(aiFields, huh.NewInput().
				Title("Model").
				Description("e.g. gpt-5.5, claude-opus-4-8 - optional.").
				Value(&model))
		}
		if !promptLocked {
			aiFields = append(aiFields, huh.NewText().
				Title("System prompt").
				Value(&systemPrompt).
				Lines(3))
		}
		aiFields = append(aiFields,
			huh.NewInput().
				Title("Max tokens").
				Description("Optional - leave blank to use the provider default.").
				Value(&maxTokensStr).
				Validate(validateOptionalInt),
			huh.NewInput().
				Title("Temperature").
				Description("Optional - 0.0 to 2.0.").
				Value(&temperatureStr).
				Validate(validateOptionalTemperature),
		)
		runFields(aiFields)

		ans.Provider = provider
		ans.Model = model
		ans.SystemPrompt = systemPrompt
		if v := strings.TrimSpace(maxTokensStr); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				ans.MaxTokens = n
			}
		}
		if v := strings.TrimSpace(temperatureStr); v != "" {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				ans.Temperature = f
			}
		}
	}

	// ---- capabilities + artifacts ----
	streaming, _ := wzBool("streaming", true)
	push, _ := wzBool("notifications", false)
	history, _ := wzBool("history", false)
	caps := []string{}
	if streaming {
		caps = append(caps, "streaming")
	}
	if push {
		caps = append(caps, "pushNotifications")
	}
	if history {
		caps = append(caps, "stateTransitionHistory")
	}
	runFields([]huh.Field{
		huh.NewMultiSelect[string]().
			Title("Capabilities").
			Description("Space to toggle, enter to confirm.").
			Height(multiSelectHeight(3)).
			Options(
				huh.NewOption("Streaming responses", "streaming").Selected(streaming),
				huh.NewOption("Push notifications", "pushNotifications").Selected(push),
				huh.NewOption("State transition history", "stateTransitionHistory").Selected(history),
			).
			Value(&caps),
	})
	var artifacts bool
	runFields([]huh.Field{
		leftConfirm().
			Title("Enable artifacts support?").
			Description("Filesystem / MinIO storage, configured via A2A_ARTIFACT_* env vars.").
			Value(&artifacts),
	})
	ans.Streaming = slices.Contains(caps, "streaming")
	ans.PushNotifications = slices.Contains(caps, "pushNotifications")
	ans.StateTransitionHistory = slices.Contains(caps, "stateTransitionHistory")
	ans.ArtifactsEnabled = artifacts

	// ---- services / tools / skills ----
	ans.Services = wizardServices()
	ans.Tools = wizardTools(ans.Services)
	ans.Skills = wizardSkills()

	// ---- server ----
	portStr := "8080"
	if v, locked := wzString("port", "8080"); locked {
		portStr = v
	}
	scheme := "http"
	debug, _ := wzBool("debug", false)
	var auth bool
	serverFields := []huh.Field{}
	if _, locked := wzString("port", "8080"); !locked {
		serverFields = append(serverFields, huh.NewInput().
			Title("Server port").
			Value(&portStr).
			Validate(validatePort))
	}
	serverFields = append(serverFields,
		huh.NewSelect[string]().
			Title("Scheme").
			Options(huh.NewOptions("http", "https")...).
			Value(&scheme),
	)
	if _, locked := wzBool("debug", false); !locked {
		serverFields = append(serverFields, leftConfirm().Title("Enable debug mode?").Value(&debug))
	}
	serverFields = append(serverFields, leftConfirm().Title("Enable server authentication?").Value(&auth))
	runFields(serverFields)

	if n, err := strconv.Atoi(strings.TrimSpace(portStr)); err == nil {
		ans.Port = n
	} else {
		ans.Port = 8080
	}
	ans.Scheme = scheme
	ans.Debug = debug
	ans.AuthEnabled = auth

	// ---- agent card ----
	collectCard(&ans)

	// ---- language ----
	collectLanguage(&ans)

	// ---- sandbox ----
	flox, _ := wzBool("flox", false)
	devcontainer, _ := wzBool("devcontainer", false)
	dockerCompose, _ := wzBool("docker-compose", false)
	sandbox := []string{}
	if flox {
		sandbox = append(sandbox, "flox")
	}
	if devcontainer {
		sandbox = append(sandbox, "devcontainer")
	}
	if dockerCompose {
		sandbox = append(sandbox, "dockerCompose")
	}
	runFields([]huh.Field{
		huh.NewMultiSelect[string]().
			Title("Development sandboxes").
			Description("Reproducible dev environments to scaffold.").
			Height(multiSelectHeight(3)).
			Options(
				huh.NewOption("Flox", "flox").Selected(flox),
				huh.NewOption("Dev Container", "devcontainer").Selected(devcontainer),
				huh.NewOption("Docker Compose", "dockerCompose").Selected(dockerCompose),
			).
			Value(&sandbox),
	})
	ans.FloxEnabled = slices.Contains(sandbox, "flox")
	ans.DevcontainerEnabled = slices.Contains(sandbox, "devcontainer")
	ans.DockerComposeEnabled = slices.Contains(sandbox, "dockerCompose")

	// ---- deployment + SCM + AI docs ----
	collectDeploymentSCM(&ans)

	return ans
}

func collectCard(ans *answers) {
	var enabled bool
	protocol := "0.3.0"
	transport := "JSONRPC"
	inputModes := []string{"text", "voice"}
	outputModes := []string{"text", "audio"}
	scheme := ans.Scheme
	if scheme == "" {
		scheme = "http"
	}
	cardURL := fmt.Sprintf("%s://%s.example.com:%d", scheme, ans.Name, ans.Port)

	form := huh.NewForm(
		huh.NewGroup(
			leftConfirm().
				Title("Configure the agent card?").
				Description("Protocol version, transport, and input/output modes.").
				Value(&enabled),
		),
		huh.NewGroup(
			huh.NewInput().Title("Protocol version").Value(&protocol),
			huh.NewInput().Title("Preferred transport").Value(&transport),
			huh.NewMultiSelect[string]().
				Title("Default input modes").
				Height(multiSelectHeight(3)).
				Options(
					huh.NewOption("text", "text").Selected(true),
					huh.NewOption("voice", "voice").Selected(true),
					huh.NewOption("audio", "audio"),
				).
				Value(&inputModes),
			huh.NewMultiSelect[string]().
				Title("Default output modes").
				Height(multiSelectHeight(3)).
				Options(
					huh.NewOption("text", "text").Selected(true),
					huh.NewOption("audio", "audio").Selected(true),
					huh.NewOption("voice", "voice"),
				).
				Value(&outputModes),
			huh.NewInput().Title("Agent service URL").Value(&cardURL),
		).WithHideFunc(func() bool { return !enabled }),
	)
	if err := tui.RunForm(form); err != nil {
		return
	}

	ans.CardEnabled = enabled
	if !enabled {
		return
	}
	ans.ProtocolVersion = protocol
	ans.PreferredTransport = transport
	ans.InputModes = inputModes
	ans.OutputModes = outputModes
	ans.CardURL = cardURL
}

func collectLanguage(ans *answers) {
	language, languageLocked := wzString("language", "typescript")
	if !languageLocked {
		runFields([]huh.Field{
			huh.NewSelect[string]().
				Title("Programming language").
				Description("The language the generated agent is scaffolded in.").
				Options(
					huh.NewOption("TypeScript", "typescript"),
					huh.NewOption("Go", "go"),
					huh.NewOption("Rust", "rust"),
				).
				Value(&language),
		})
	}
	ans.Language = language

	switch language {
	case "rust":
		pkg, pkgLocked := wzString("rust-package-name", ans.Name)
		ver, verLocked := wzString("rust-version", "1.89.0")
		edition, editionLocked := wzString("rust-edition", "2024")
		var fields []huh.Field
		if !pkgLocked {
			fields = append(fields, huh.NewInput().Title("Rust package name").Value(&pkg))
		}
		if !verLocked {
			fields = append(fields, huh.NewInput().Title("Rust version").Value(&ver))
		}
		if !editionLocked {
			fields = append(fields, huh.NewInput().Title("Rust edition").Value(&edition))
		}
		runFields(fields)
		ans.RustPackageName = pkg
		ans.RustVersion = ver
		ans.RustEdition = edition
	case "typescript":
		pkg, pkgLocked := wzString("typescript-name", ans.Name)
		if !pkgLocked {
			runFields([]huh.Field{huh.NewInput().Title("TypeScript package name").Value(&pkg)})
		}
		ans.TSPackageName = pkg
	default:
		module, moduleLocked := wzString("go-module", getDefaultGoModule(ans.Name))
		ver, verLocked := wzString("go-version", "1.26.2")
		var fields []huh.Field
		if !moduleLocked {
			fields = append(fields, huh.NewInput().
				Title("Go module path").
				Description("e.g. github.com/you/your-agent").
				Value(&module))
		}
		if !verLocked {
			fields = append(fields, huh.NewInput().Title("Go version").Value(&ver))
		}
		runFields(fields)
		ans.GoModule = module
		ans.GoVersion = ver
	}
}

func collectDeploymentSCM(ans *answers) {
	deployment := ""
	if v, locked := wzString("deployment", ""); locked {
		deployment = v
	} else {
		runFields([]huh.Field{
			huh.NewSelect[string]().
				Title("Deployment target").
				Description("Generate deployment manifests for this platform.").
				Options(
					huh.NewOption("None", ""),
					huh.NewOption("Kubernetes", "kubernetes"),
					huh.NewOption("Cloud Run", "cloudrun"),
				).
				Value(&deployment),
		})
	}
	ans.DeploymentType = deployment

	provider := "github"
	runFields([]huh.Field{
		huh.NewSelect[string]().
			Title("Source control provider").
			Options(
				huh.NewOption("GitHub", "github"),
				huh.NewOption("GitLab", "gitlab"),
				huh.NewOption("Bitbucket", "bitbucket"),
				huh.NewOption("None", ""),
			).
			Value(&provider),
	})
	ans.ScmProvider = provider

	if provider == "github" {
		owner, repo := parseGitRemote()
		defaultURL := fmt.Sprintf("https://github.com/example/%s", ans.Name)
		if owner != "" && repo != "" {
			defaultURL = fmt.Sprintf("https://github.com/%s/%s", owner, repo)
		}
		url := defaultURL
		githubApp := true
		var issueTemplates, dependabot bool
		runFields([]huh.Field{
			huh.NewInput().Title("Repository URL").Value(&url),
			leftConfirm().Title("Enable GitHub App integration?").Value(&githubApp),
			leftConfirm().Title("Generate issue templates?").Value(&issueTemplates),
			leftConfirm().Title("Enable Dependabot?").Value(&dependabot),
		})
		ans.ScmURL = url
		ans.GithubApp = githubApp
		ans.IssueTemplates = issueTemplates
		ans.Dependabot = dependabot
	}

	if provider != "" {
		ci, _ := wzBool("ci", false)
		cd, _ := wzBool("cd", false)
		ciCdFields := []huh.Field{}
		if _, locked := wzBool("ci", false); !locked {
			ciCdFields = append(ciCdFields, leftConfirm().Title("Generate a CI workflow?").Value(&ci))
		}
		if _, locked := wzBool("cd", false); !locked {
			ciCdFields = append(ciCdFields, leftConfirm().Title("Generate a CD pipeline?").Value(&cd))
		}
		runFields(ciCdFields)
		ans.CI = ci
		ans.CD = cd
	}

	// AI coding assistants. Only claudecode is flag-backed (--ai); the rest are
	// wizard/manifest-only, so the multi-select is always shown with claudecode
	// pre-selected when --ai was passed.
	claudecode, _ := wzBool("ai", false)
	orchestrators := []string{}
	if claudecode {
		orchestrators = append(orchestrators, "claudecode")
	}
	runFields([]huh.Field{
		huh.NewMultiSelect[string]().
			Title("AI coding assistants").
			Description("Scaffold docs and workflows for these coding agents.").
			Height(multiSelectHeight(5)).
			Options(
				huh.NewOption("Claude Code", "claudecode").Selected(claudecode),
				huh.NewOption("Codex", "codex"),
				huh.NewOption("Gemini", "gemini"),
				huh.NewOption("OpenCode", "opencode"),
				huh.NewOption("Infer", "infer"),
			).
			Value(&orchestrators),
	})
	ans.Claudecode = slices.Contains(orchestrators, "claudecode")
	ans.Codex = slices.Contains(orchestrators, "codex")
	ans.Gemini = slices.Contains(orchestrators, "gemini")
	ans.Opencode = slices.Contains(orchestrators, "opencode")
	ans.Infer = slices.Contains(orchestrators, "infer")
}

// wizardServices runs the "add another service" loop.
func wizardServices() []string {
	var add bool
	if err := tui.RunForm(huh.NewForm(huh.NewGroup(
		leftConfirm().
			Title("Add service dependencies?").
			Description("Injected singletons such as a logger, database, or cache.").
			Value(&add),
	))); err != nil {
		return nil
	}
	if !add {
		return nil
	}

	var services []string
	for {
		var (
			name string
			more bool
		)
		captured := services
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Service name").
				Placeholder("logger").
				Value(&name).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return errors.New("a name is required")
					}
					if !isValidIdentifier(s) {
						return errors.New("use letters, numbers, and underscores, starting with a letter or underscore")
					}
					if slices.Contains(captured, s) {
						return errors.New("already added")
					}
					return nil
				}),
			leftConfirm().Title("Add another service?").Value(&more),
		))
		if err := tui.RunForm(form); err != nil {
			return services
		}
		services = append(services, strings.TrimSpace(name))
		if !more {
			break
		}
	}
	return services
}

// wizardTools runs the "add another tool" loop, optionally injecting services.
func wizardTools(services []string) []toolAnswer {
	var add bool
	if err := tui.RunForm(huh.NewForm(huh.NewGroup(
		leftConfirm().
			Title("Add tools?").
			Description("Function-call entry points the agent can invoke.").
			Value(&add),
	))); err != nil {
		return nil
	}
	if !add {
		return nil
	}

	var tools []toolAnswer
	for {
		var (
			name, description, tags string
			inject                  []string
			more                    bool
		)
		fields := []huh.Field{
			huh.NewInput().
				Title("Tool name").
				Placeholder("get_weather").
				Value(&name).
				Validate(requireNonEmpty),
			huh.NewInput().Title("Description").Value(&description),
			huh.NewInput().
				Title("Tags").
				Description("Comma-separated; defaults to \"general\".").
				Placeholder("weather,api").
				Value(&tags),
		}
		if len(services) > 0 {
			opts := make([]huh.Option[string], 0, len(services))
			for _, svc := range services {
				opts = append(opts, huh.NewOption(svc, svc))
			}
			fields = append(fields, huh.NewMultiSelect[string]().
				Title("Inject services").
				Description("Services to make available inside this tool.").
				Height(multiSelectHeight(len(services))).
				Options(opts...).
				Value(&inject))
		}
		fields = append(fields, leftConfirm().Title("Add another tool?").Value(&more))

		if err := tui.RunForm(huh.NewForm(huh.NewGroup(fields...))); err != nil {
			return tools
		}

		tool := toolAnswer{
			ID:          strings.TrimSpace(name),
			Name:        strings.TrimSpace(name),
			Description: description,
			Inject:      inject,
		}
		if strings.TrimSpace(tags) != "" {
			tool.Tags = splitAndTrim(tags)
		} else {
			tool.Tags = []string{"general"}
		}
		tools = append(tools, tool)

		if !more {
			break
		}
	}
	return tools
}

// wizardSkills runs the "add another skill" loop.
func wizardSkills() []skillAnswer {
	var add bool
	if err := tui.RunForm(huh.NewForm(huh.NewGroup(
		leftConfirm().
			Title("Add markdown skills?").
			Description("Playbooks loaded into the system prompt at startup.").
			Value(&add),
	))); err != nil {
		return nil
	}
	if !add {
		return nil
	}

	var skills []skillAnswer
	for {
		var (
			id, source, name, description, tags, version string
			more                                         bool
		)
		source = "registry"
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Skill id").
					Description("kebab-case, e.g. data-analysis.").
					Value(&id).
					Validate(requireNonEmpty),
				huh.NewSelect[string]().
					Title("Source").
					Options(
						huh.NewOption("Registry", "registry"),
						huh.NewOption("Bare (scaffold a blank skill)", "bare"),
					).
					Value(&source),
			),
			huh.NewGroup(
				huh.NewInput().Title("Skill name").Value(&name),
				huh.NewInput().Title("Description").Value(&description),
				huh.NewInput().Title("Tags").Description("Comma-separated, optional.").Value(&tags),
			).WithHideFunc(func() bool { return source != "bare" }),
			huh.NewGroup(
				huh.NewInput().
					Title("Pin to version").
					Description("Optional, e.g. 0.1.0.").
					Value(&version),
			).WithHideFunc(func() bool { return source != "registry" }),
			huh.NewGroup(
				leftConfirm().Title("Add another skill?").Value(&more),
			),
		)
		if err := tui.RunForm(form); err != nil {
			return skills
		}

		skill := skillAnswer{ID: strings.TrimSpace(id)}
		if source == "bare" {
			skill.Bare = true
			skill.Name = strings.TrimSpace(name)
			if skill.Name == "" {
				skill.Name = skill.ID
			}
			skill.Description = description
			if strings.TrimSpace(tags) != "" {
				skill.Tags = splitAndTrim(tags)
			}
		} else {
			skill.Version = strings.TrimSpace(version)
		}
		skills = append(skills, skill)

		if !more {
			break
		}
	}
	return skills
}

// printInitSummary renders the closing summary panel and next steps. It is
// shared by the interactive wizard and the --defaults/non-interactive flow so
// both finish identically; output is routed through tui.Println, which strips
// color when stdout is not a terminal.
func printInitSummary(adl *adlData, adlFile, projectDir string) {
	lang := "go"
	if adl.Spec.Language != nil {
		switch {
		case adl.Spec.Language.Rust != nil:
			lang = "rust"
		case adl.Spec.Language.TypeScript != nil:
			lang = "typescript"
		}
	}
	agentType := "minimal"
	if adl.Spec.Agent != nil {
		agentType = "ai-powered"
	}

	fields := []tui.Field{
		{Label: "Agent", Value: adl.Metadata.Name},
		{Label: "Type", Value: agentType},
		{Label: "Language", Value: lang},
		{Label: "Manifest", Value: adlFile},
	}

	steps := []string{
		"Run 'adl generate' to generate the project code",
		"Implement the TODO placeholders in the generated files",
		"Run 'task build' to build your agent",
		"Run 'task run' to start your agent server",
	}
	if projectDir != "." && projectDir != "" {
		steps = append([]string{fmt.Sprintf("cd %s", projectDir)}, steps...)
	}

	tui.Println("")
	tui.Println(tui.SummaryBox(fmt.Sprintf("Project '%s' initialized", adl.Metadata.Name), fields))
	tui.Println("")
	tui.Println(tui.NextSteps(steps))
	tui.Println("")
}

// ---- small helpers ----

// runFields runs a single-page form built from the given fields. Empty field
// lists (everything resolved via flags) are skipped.
func runFields(fields []huh.Field) {
	if len(fields) == 0 {
		return
	}
	_ = tui.RunForm(huh.NewForm(huh.NewGroup(fields...)))
}

// leftConfirm builds a confirm whose Yes/No buttons are left-aligned. huh
// defaults the button alignment to center, which indents the buttons relative to
// the rest of the (left-aligned) form.
func leftConfirm() *huh.Confirm {
	return huh.NewConfirm().WithButtonAlignment(lipgloss.Left)
}

// multiSelectHeight returns the field height that keeps every option of a
// static multi-select visible. huh derives the viewport height from the field
// height minus the title and description lines, so without this a small static
// option list collapses to a single visible row.
func multiSelectHeight(options int) int {
	return options + 2
}

// wzString resolves a string answer: a flag/env value (via viper) wins and is
// reported as locked so the wizard does not re-prompt; otherwise the default is
// returned unlocked.
func wzString(key, def string) (value string, locked bool) {
	if viper.IsSet(key) && viper.GetString(key) != "" {
		return viper.GetString(key), true
	}
	return def, false
}

// wzBool resolves a bool answer with the same precedence as wzString.
func wzBool(key string, def bool) (value, locked bool) {
	if viper.IsSet(key) {
		return viper.GetBool(key), true
	}
	return def, false
}

func requireNonEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.New("this field is required")
	}
	return nil
}

func validatePort(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return errors.New("a port is required")
	}
	if _, err := strconv.Atoi(s); err != nil {
		return errors.New("must be a whole number")
	}
	return nil
}

func validateOptionalInt(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if _, err := strconv.Atoi(s); err != nil {
		return errors.New("must be a whole number")
	}
	return nil
}

func validateOptionalTemperature(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return errors.New("must be a number")
	}
	if v < 0 || v > 2 {
		return errors.New("must be between 0.0 and 2.0")
	}
	return nil
}

func absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, p)
}
