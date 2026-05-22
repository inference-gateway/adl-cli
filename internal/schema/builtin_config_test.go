package schema

import (
	"strings"
	"testing"
)

func TestIsReservedToolID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		id   string
		want bool
	}{
		{"read", true},
		{"bash", true},
		{"write", true},
		{"edit", true},
		{"fetch", true},
		{"http", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := IsReservedToolID(tc.id); got != tc.want {
			t.Errorf("IsReservedToolID(%q) = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestBuiltinToolMetaFor_Fetch(t *testing.T) {
	t.Parallel()

	meta := BuiltinToolMetaFor(string(ReservedToolFetch))
	if meta.Name != "Fetch" {
		t.Fatalf("expected fetch meta name 'Fetch', got %q", meta.Name)
	}
	if meta.Description == "" {
		t.Fatalf("expected non-empty fetch description")
	}
	wantParams := []string{"url", "method", "save_path", "headers"}
	if len(meta.Parameters) != len(wantParams) {
		t.Fatalf("expected %d parameters, got %d (%v)", len(wantParams), len(meta.Parameters), meta.Parameters)
	}
	for i, p := range wantParams {
		if meta.Parameters[i] != p {
			t.Errorf("parameter[%d] = %q, want %q", i, meta.Parameters[i], p)
		}
	}
}

func TestDecodeBuiltinToolConfig_Fetch(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"enabled": true,
		"allowed_domains": []any{
			"example.com",
			".api.dev",
		},
		"max_bytes":       1048576,
		"timeout_seconds": 15,
		"download_dir":    "/var/tmp/downloads",
		"allow_downloads": true,
	}

	decoded, err := DecodeBuiltinToolConfig(string(ReservedToolFetch), raw)
	if err != nil {
		t.Fatalf("DecodeBuiltinToolConfig fetch returned error: %v", err)
	}
	cfg, ok := decoded.(*FetchBuiltinConfig)
	if !ok {
		t.Fatalf("expected *FetchBuiltinConfig, got %T", decoded)
	}
	if !cfg.Enabled {
		t.Error("expected Enabled=true")
	}
	if cfg.MaxBytes != 1048576 {
		t.Errorf("MaxBytes = %d, want 1048576", cfg.MaxBytes)
	}
	if cfg.TimeoutSeconds != 15 {
		t.Errorf("TimeoutSeconds = %d, want 15", cfg.TimeoutSeconds)
	}
	if cfg.DownloadDir != "/var/tmp/downloads" {
		t.Errorf("DownloadDir = %q, want /var/tmp/downloads", cfg.DownloadDir)
	}
	if !cfg.AllowDownloads {
		t.Error("expected AllowDownloads=true")
	}
	if len(cfg.AllowedDomains) != 2 || cfg.AllowedDomains[0] != "example.com" || cfg.AllowedDomains[1] != ".api.dev" {
		t.Errorf("AllowedDomains = %v", cfg.AllowedDomains)
	}
}

func TestDecodeBuiltinToolConfig_FetchUnknownKeyRejected(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"enabled":  true,
		"max_byts": 10, // typo
	}
	_, err := DecodeBuiltinToolConfig(string(ReservedToolFetch), raw)
	if err == nil {
		t.Fatal("expected error for unknown fetch config key, got nil")
	}
	if !strings.Contains(err.Error(), "spec.config.tools.fetch") {
		t.Errorf("error %q does not mention spec.config.tools.fetch", err.Error())
	}
}

func TestResolveBuiltinConfigs_FetchPresent(t *testing.T) {
	t.Parallel()

	adl := &ADL{}
	adl.Spec.Config = map[string]map[string]any{
		"tools": {
			"fetch": map[string]any{
				"enabled":         true,
				"max_bytes":       2048,
				"timeout_seconds": 10,
				"download_dir":    "/tmp/fetch",
				"allow_downloads": true,
				"allowed_domains": []any{"foo.com"},
			},
		},
	}

	out, err := ResolveBuiltinConfigs(adl)
	if err != nil {
		t.Fatalf("ResolveBuiltinConfigs returned error: %v", err)
	}
	if !out.Fetch.Enabled || out.Fetch.MaxBytes != 2048 || out.Fetch.DownloadDir != "/tmp/fetch" {
		t.Errorf("Fetch config not decoded correctly: %+v", out.Fetch)
	}
}

func TestResolveBuiltinConfigs_FetchAbsent(t *testing.T) {
	t.Parallel()

	adl := &ADL{}
	out, err := ResolveBuiltinConfigs(adl)
	if err != nil {
		t.Fatalf("ResolveBuiltinConfigs returned error: %v", err)
	}
	if out.Fetch.Enabled {
		t.Error("expected Fetch.Enabled to default to false when absent")
	}
}
