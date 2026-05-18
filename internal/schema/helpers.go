package schema

// This file contains hand-written accessor helpers on top of the generated
// types in types.go. Generated optional fields are pointers (e.g. *bool,
// *string, *EnumType); these helpers return zero-value defaults so callers
// don't have to nil-check at every site, and they let Go templates use
// `eq` / `if` without dereferencing.

// ---- accessors on optional fields ----

// GetType returns the deployment type as a string, or "" if unset.
func (d *DeploymentConfig) GetType() string {
	if d == nil || d.Type == nil {
		return ""
	}
	return string(*d.Type)
}

// GetProvider returns the SCM provider as a string, or "" if unset.
func (s *SCM) GetProvider() string {
	if s == nil || s.Provider == nil {
		return ""
	}
	return string(*s.Provider)
}

// GetURL returns the SCM URL or "" if unset.
func (s *SCM) GetURL() string {
	if s == nil || s.URL == nil {
		return ""
	}
	return *s.URL
}

// GetIssueTemplates reports whether issue templates are enabled.
func (s *SCM) GetIssueTemplates() bool {
	return s != nil && s.IssueTemplates != nil && *s.IssueTemplates
}

// GetGithubApp reports whether the GitHub App integration is enabled.
func (s *SCM) GetGithubApp() bool {
	return s != nil && s.GithubApp != nil && *s.GithubApp
}

// GetProvider returns the agent's AI provider as a string, or "" if unset.
func (a *Agent) GetProvider() string {
	if a == nil || a.Provider == nil {
		return ""
	}
	return string(*a.Provider)
}

// GetModel returns the configured model identifier, or "" if unset.
func (a *Agent) GetModel() string {
	if a == nil || a.Model == nil {
		return ""
	}
	return *a.Model
}

// GetSystemPrompt returns the configured system prompt, or "" if unset.
func (a *Agent) GetSystemPrompt() string {
	if a == nil || a.SystemPrompt == nil {
		return ""
	}
	return *a.SystemPrompt
}

// GetMaxTokens returns the configured max tokens, or 0 if unset.
func (a *Agent) GetMaxTokens() int {
	if a == nil || a.MaxTokens == nil {
		return 0
	}
	return *a.MaxTokens
}

// GetTemperature returns the configured temperature, or 0 if unset.
func (a *Agent) GetTemperature() float64 {
	if a == nil || a.Temperature == nil {
		return 0
	}
	return *a.Temperature
}

// GetScheme returns the server scheme or "" if unset.
func (s *Server) GetScheme() string {
	if s == nil || s.Scheme == nil {
		return ""
	}
	return *s.Scheme
}

// GetDebug reports whether debug mode is enabled.
func (s *Server) GetDebug() bool {
	return s != nil && s.Debug != nil && *s.Debug
}

// GetEnabled reports whether auth is enabled.
func (a *AuthConfig) GetEnabled() bool {
	return a != nil && a.Enabled != nil && *a.Enabled
}

// GetURL returns the agent-card URL or "" if unset.
func (c *Card) GetURL() string {
	if c == nil || c.URL == nil {
		return ""
	}
	return *c.URL
}

// GetDocumentationURL returns the documentation URL or "" if unset.
func (c *Card) GetDocumentationURL() string {
	if c == nil || c.DocumentationURL == nil {
		return ""
	}
	return *c.DocumentationURL
}

// GetIconURL returns the icon URL or "" if unset.
func (c *Card) GetIconURL() string {
	if c == nil || c.IconURL == nil {
		return ""
	}
	return *c.IconURL
}

// GetProtocolVersion returns the protocol version or "" if unset.
func (c *Card) GetProtocolVersion() string {
	if c == nil || c.ProtocolVersion == nil {
		return ""
	}
	return *c.ProtocolVersion
}

// GetPreferredTransport returns the preferred transport or "" if unset.
func (c *Card) GetPreferredTransport() string {
	if c == nil || c.PreferredTransport == nil {
		return ""
	}
	return *c.PreferredTransport
}

// GetRegistry returns the CloudRun image registry or "" if unset.
func (i *ImageConfig) GetRegistry() string {
	if i == nil || i.Registry == nil {
		return ""
	}
	return *i.Registry
}

// GetRepository returns the CloudRun image repository or "" if unset.
func (i *ImageConfig) GetRepository() string {
	if i == nil || i.Repository == nil {
		return ""
	}
	return *i.Repository
}

// GetTag returns the CloudRun image tag or "" if unset.
func (i *ImageConfig) GetTag() string {
	if i == nil || i.Tag == nil {
		return ""
	}
	return *i.Tag
}

// GetUseCloudBuild reports whether the image is built with Cloud Build.
func (i *ImageConfig) GetUseCloudBuild() bool {
	return i != nil && i.UseCloudBuild != nil && *i.UseCloudBuild
}

// GetCPU returns the CloudRun CPU allocation or "" if unset.
func (r *ResourcesConfig) GetCPU() string {
	if r == nil || r.CPU == nil {
		return ""
	}
	return *r.CPU
}

// GetMemory returns the CloudRun memory allocation or "" if unset.
func (r *ResourcesConfig) GetMemory() string {
	if r == nil || r.Memory == nil {
		return ""
	}
	return *r.Memory
}

// GetMinInstances returns the configured minimum instances or 0 if unset.
func (s *ScalingConfig) GetMinInstances() int {
	if s == nil || s.MinInstances == nil {
		return 0
	}
	return *s.MinInstances
}

// GetMaxInstances returns the configured maximum instances or 0 if unset.
func (s *ScalingConfig) GetMaxInstances() int {
	if s == nil || s.MaxInstances == nil {
		return 0
	}
	return *s.MaxInstances
}

// GetConcurrency returns the configured concurrency or 0 if unset.
func (s *ScalingConfig) GetConcurrency() int {
	if s == nil || s.Concurrency == nil {
		return 0
	}
	return *s.Concurrency
}

// GetTimeout returns the configured CloudRun service timeout or 0 if unset.
func (s *ServiceConfig) GetTimeout() int {
	if s == nil || s.Timeout == nil {
		return 0
	}
	return *s.Timeout
}

// GetAllowUnauthenticated reports whether unauthenticated invocations are allowed.
func (s *ServiceConfig) GetAllowUnauthenticated() bool {
	return s != nil && s.AllowUnauthenticated != nil && *s.AllowUnauthenticated
}

// GetServiceAccount returns the CloudRun service account or "" if unset.
func (s *ServiceConfig) GetServiceAccount() string {
	if s == nil || s.ServiceAccount == nil {
		return ""
	}
	return *s.ServiceAccount
}

// GetExecutionEnvironment returns the CloudRun execution environment or "" if unset.
func (s *ServiceConfig) GetExecutionEnvironment() string {
	if s == nil || s.ExecutionEnvironment == nil {
		return ""
	}
	return *s.ExecutionEnvironment
}

// GetImplementation returns the skill implementation path or "" if unset.
func (s *Skill) GetImplementation() string {
	if s == nil || s.Implementation == nil {
		return ""
	}
	return *s.Implementation
}

// ---- constructor helpers ----

// StrPtr returns a pointer to the given string.
func StrPtr(s string) *string { return &s }

// BoolPtr returns a pointer to the given bool.
func BoolPtr(b bool) *bool { return &b }

// IntPtr returns a pointer to the given int.
func IntPtr(i int) *int { return &i }

// Float64Ptr returns a pointer to the given float64.
func Float64Ptr(f float64) *float64 { return &f }

// DeploymentTypePtr returns a pointer to a DeploymentConfigType.
func DeploymentTypePtr(t DeploymentConfigType) *DeploymentConfigType { return &t }

// SCMProviderPtr returns a pointer to an SCMProvider.
func SCMProviderPtr(p SCMProvider) *SCMProvider { return &p }

// AgentProviderPtr returns a pointer to an AgentProvider.
func AgentProviderPtr(p AgentProvider) *AgentProvider { return &p }
