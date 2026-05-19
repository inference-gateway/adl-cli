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
)

// ReservedToolIDs is the lookup set used by validator/generator to detect
// reserved IDs. Keep in sync with the constants above.
var ReservedToolIDs = map[string]struct{}{
	string(ReservedToolRead):  {},
	string(ReservedToolBash):  {},
	string(ReservedToolWrite): {},
	string(ReservedToolEdit):  {},
}

// IsReservedToolID reports whether id maps to a built-in implementation.
func IsReservedToolID(id string) bool {
	_, ok := ReservedToolIDs[id]
	return ok
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

// ResolvedBuiltinConfigs holds the decoded config blocks for every reserved
// tool id present in spec.tools. Entries for absent ids carry default
// (zero) values, which means Enabled=false (the safe default).
type ResolvedBuiltinConfigs struct {
	Read  ReadBuiltinConfig
	Bash  BashBuiltinConfig
	Write WriteBuiltinConfig
	Edit  EditBuiltinConfig
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
