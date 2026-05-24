package schema

import (
	"fmt"

	"github.com/go-viper/mapstructure/v2"
)

// ReservedToolID is the set of tool IDs that map to built-in
// implementations supplied by the generator. Users opt in by listing one
// of these under spec.tools, then enable + tune via spec.config.tools.<id>.
type ReservedToolID string

const (
	ReservedToolRead  ReservedToolID = "read"
	ReservedToolBash  ReservedToolID = "bash"
	ReservedToolWrite ReservedToolID = "write"
	ReservedToolEdit  ReservedToolID = "edit"
	ReservedToolFetch ReservedToolID = "fetch"
)

// ReservedToolIDs is the lookup set used by validator/generator to detect
// reserved IDs. Keep in sync with the constants above.
var ReservedToolIDs = map[string]struct{}{
	string(ReservedToolRead):  {},
	string(ReservedToolBash):  {},
	string(ReservedToolWrite): {},
	string(ReservedToolEdit):  {},
	string(ReservedToolFetch): {},
}

// IsReservedToolID reports whether id maps to a built-in implementation.
func IsReservedToolID(id string) bool {
	_, ok := ReservedToolIDs[id]
	return ok
}

// BuiltinToolMeta is the documentation-facing view of a reserved built-in
// tool - the same information the generator bakes into the per-language
// runtime descriptor, surfaced here so templates (README, CLAUDE.md) can
// render a uniform table without each template re-stating it.
type BuiltinToolMeta struct {
	ID          string
	Name        string
	Description string
	Parameters  []string
}

// BuiltinToolMetas maps reserved tool IDs to their documentation metadata.
// Keep the Name/Description/Parameters here aligned with what each
// language's `builtin/<id>.<ext>.tmpl` advertises to the LLM.
var BuiltinToolMetas = map[string]BuiltinToolMeta{
	string(ReservedToolRead): {
		ID:          string(ReservedToolRead),
		Name:        "Read",
		Description: "Read a file from disk. Returns its contents, optionally sliced by line offset/limit. Use this to load SKILL.md bodies on demand.",
		Parameters:  []string{"file_path", "offset", "limit"},
	},
	string(ReservedToolBash): {
		ID:          string(ReservedToolBash),
		Name:        "Bash",
		Description: "Execute a shell command. Subject to the configured whitelist and timeout; honors A2A_BASH_DISABLED as a runtime kill switch.",
		Parameters:  []string{"command"},
	},
	string(ReservedToolWrite): {
		ID:          string(ReservedToolWrite),
		Name:        "Write",
		Description: "Write content to a file, creating intermediate directories as needed. Overwrites the file if it already exists.",
		Parameters:  []string{"file_path", "content"},
	},
	string(ReservedToolEdit): {
		ID:          string(ReservedToolEdit),
		Name:        "Edit",
		Description: "Replace a unique string in a file with a new value. Errors if old_string is not found or appears more than once.",
		Parameters:  []string{"file_path", "old_string", "new_string"},
	},
	string(ReservedToolFetch): {
		ID:          string(ReservedToolFetch),
		Name:        "Fetch",
		Description: "Fetch a URL over HTTP(S). Subject to an allowed-domains whitelist and a max-bytes cap; can optionally save the response body to a file inside the configured download_dir (defaults to /tmp).",
		Parameters:  []string{"url", "method", "save_path", "headers"},
	},
}

// BuiltinToolMetaFor returns the metadata for a reserved tool ID, or an
// empty struct if id is not a reserved built-in. Templates can call this
// unconditionally and check `.Name` to detect presence.
func BuiltinToolMetaFor(id string) BuiltinToolMeta {
	return BuiltinToolMetas[id]
}

// ReadBuiltinConfig is the typed shape of spec.config.tools.read.
type ReadBuiltinConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	MaxLines     int      `mapstructure:"max_lines"`
	AllowedRoots []string `mapstructure:"allowed_roots"`
}

// BashBuiltinConfig is the typed shape of spec.config.tools.bash.
type BashBuiltinConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	Whitelist      []string `mapstructure:"whitelist"`
	TimeoutSeconds int      `mapstructure:"timeout_seconds"`
	WorkingDir     string   `mapstructure:"working_dir"`
}

// WriteBuiltinConfig is the typed shape of spec.config.tools.write.
type WriteBuiltinConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	AllowedRoots []string `mapstructure:"allowed_roots"`
}

// EditBuiltinConfig is the typed shape of spec.config.tools.edit.
type EditBuiltinConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	AllowedRoots []string `mapstructure:"allowed_roots"`
}

// FetchBuiltinConfig is the typed shape of spec.config.tools.fetch.
//
// AllowedDomains is the host whitelist; entries are matched against the
// request URL's host (case-insensitive). An entry beginning with "." is
// treated as a suffix match, so ".example.com" allows any subdomain. An
// empty list with Enabled=true means "any host is allowed" - intentional
// for users who want unrestricted HTTP access, but discouraged.
//
// MaxBytes caps how much of the response body the tool reads (0 means
// "use the built-in default of 10 MiB"). DownloadDir is the root the
// tool writes to when the model requests `save_path`; AllowDownloads
// must be true to enable file output at all. TimeoutSeconds caps the
// total request time (0 means "use the built-in default of 30s").
type FetchBuiltinConfig struct {
	Enabled        bool     `mapstructure:"enabled"`
	AllowedDomains []string `mapstructure:"allowed_domains"`
	MaxBytes       int      `mapstructure:"max_bytes"`
	TimeoutSeconds int      `mapstructure:"timeout_seconds"`
	DownloadDir    string   `mapstructure:"download_dir"`
	AllowDownloads bool     `mapstructure:"allow_downloads"`
}

// ResolvedBuiltinConfigs holds the decoded config blocks for every reserved
// tool id present in spec.tools. Entries for absent ids carry default
// (zero) values, which means Enabled=false (the safe default).
type ResolvedBuiltinConfigs struct {
	Read  ReadBuiltinConfig
	Bash  BashBuiltinConfig
	Write WriteBuiltinConfig
	Edit  EditBuiltinConfig
	Fetch FetchBuiltinConfig
}

// ResolveBuiltinConfigs decodes spec.config.tools.<reserved-id> into
// typed structs. Missing sections yield zero-valued (disabled) configs.
// Returns an error only if a present section has a bad shape or unknown
// keys. The validator should have caught these already; this function
// re-validates so codegen can rely on the result.
func ResolveBuiltinConfigs(adl *ADL) (ResolvedBuiltinConfigs, error) {
	var out ResolvedBuiltinConfigs
	if adl == nil {
		return out, nil
	}
	toolsCfg, ok := adl.Spec.Config["tools"]
	if !ok {
		return out, nil
	}

	if raw, present := toolsCfg[string(ReservedToolRead)]; present {
		decoded, err := DecodeBuiltinToolConfig(string(ReservedToolRead), raw)
		if err != nil {
			return out, err
		}
		out.Read = *decoded.(*ReadBuiltinConfig)
	}
	if raw, present := toolsCfg[string(ReservedToolBash)]; present {
		decoded, err := DecodeBuiltinToolConfig(string(ReservedToolBash), raw)
		if err != nil {
			return out, err
		}
		out.Bash = *decoded.(*BashBuiltinConfig)
	}
	if raw, present := toolsCfg[string(ReservedToolWrite)]; present {
		decoded, err := DecodeBuiltinToolConfig(string(ReservedToolWrite), raw)
		if err != nil {
			return out, err
		}
		out.Write = *decoded.(*WriteBuiltinConfig)
	}
	if raw, present := toolsCfg[string(ReservedToolEdit)]; present {
		decoded, err := DecodeBuiltinToolConfig(string(ReservedToolEdit), raw)
		if err != nil {
			return out, err
		}
		out.Edit = *decoded.(*EditBuiltinConfig)
	}
	if raw, present := toolsCfg[string(ReservedToolFetch)]; present {
		decoded, err := DecodeBuiltinToolConfig(string(ReservedToolFetch), raw)
		if err != nil {
			return out, err
		}
		out.Fetch = *decoded.(*FetchBuiltinConfig)
	}
	return out, nil
}

// DecodeBuiltinToolConfig decodes the raw value under
// spec.config.tools.<id> into the matching typed struct, rejecting any
// keys the built-in doesn't know about. Returns a non-nil error with a
// path-prefixed message ("spec.config.tools.<id>.<bad-key>") so callers
// can surface a precise location. Accepts `any` to interoperate with
// untyped YAML decoding.
func DecodeBuiltinToolConfig(id string, raw any) (any, error) {
	var target any
	switch id {
	case string(ReservedToolRead):
		target = &ReadBuiltinConfig{}
	case string(ReservedToolBash):
		target = &BashBuiltinConfig{}
	case string(ReservedToolWrite):
		target = &WriteBuiltinConfig{}
	case string(ReservedToolEdit):
		target = &EditBuiltinConfig{}
	case string(ReservedToolFetch):
		target = &FetchBuiltinConfig{}
	default:
		return nil, fmt.Errorf("not a reserved built-in tool id: %q", id)
	}

	if _, ok := raw.(map[string]any); !ok && raw != nil {
		return nil, fmt.Errorf("spec.config.tools.%s must be a mapping (got %T)", id, raw)
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      target,
	})
	if err != nil {
		return nil, fmt.Errorf("build decoder for spec.config.tools.%s: %w", id, err)
	}
	if err := decoder.Decode(raw); err != nil {
		return nil, fmt.Errorf("spec.config.tools.%s: %w", id, err)
	}
	return target, nil
}
