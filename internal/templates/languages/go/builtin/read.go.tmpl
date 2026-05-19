package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	server "github.com/inference-gateway/adk/server"
	zap "go.uber.org/zap"
)

// ReadConfig is the compile-time config for the Read built-in. Values
// flow in from spec.config.tools.read at generation time.
type ReadConfig struct {
	Enabled      bool
	MaxLines     int
	AllowedRoots []string
}

// ReadTool exposes a Read built-in. Disabled by default; flip
// spec.config.tools.read.enabled: true in your ADL to activate.
type ReadTool struct {
	logger *zap.Logger
	cfg    ReadConfig
}

// NewReadTool builds a Read tool with the resolved config baked in.
func NewReadTool(logger *zap.Logger, cfg ReadConfig) server.Tool {
	if cfg.MaxLines <= 0 {
		cfg.MaxLines = 2000
	}
	t := &ReadTool{logger: logger, cfg: cfg}
	return server.NewBasicTool(
		"Read",
		"Read a file from disk. Returns its contents, optionally sliced by line offset/limit. Use this to load SKILL.md bodies on demand.",
		map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "Path to the file (absolute or relative to the agent's working directory).",
				},
				"offset": map[string]any{
					"type":        "integer",
					"description": "1-based line number to start reading from.",
					"minimum":     1,
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of lines to return.",
					"minimum":     1,
				},
			},
			"required": []string{"file_path"},
		},
		t.Handler,
	)
}

// Handler executes the Read tool.
func (t *ReadTool) Handler(ctx context.Context, args map[string]any) (string, error) {
	if !t.cfg.Enabled {
		return "", errors.New("Read tool is disabled; set spec.config.tools.read.enabled: true in the ADL and regenerate")
	}

	rawPath, _ := args["file_path"].(string)
	if rawPath == "" {
		return "", errors.New("file_path is required")
	}

	if err := t.validatePath(rawPath); err != nil {
		return "", err
	}

	offset := 1
	if v, ok := args["offset"]; ok {
		if iv, ok := toInt(v); ok && iv > 0 {
			offset = iv
		}
	}
	limit := t.cfg.MaxLines
	if v, ok := args["limit"]; ok {
		if iv, ok := toInt(v); ok && iv > 0 {
			limit = iv
		}
	}

	cleaned := filepath.Clean(rawPath)

	if isImagePath(cleaned) {
		return "", fmt.Errorf("Read does not support image files (%s); use a multimodal API instead", filepath.Ext(cleaned))
	}

	file, err := os.Open(cleaned)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", cleaned, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.logger.Warn("failed to close file", zap.String("path", cleaned), zap.Error(closeErr))
		}
	}()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	var b strings.Builder
	currentLine := 0
	emitted := 0
	for scanner.Scan() {
		currentLine++
		if currentLine < offset {
			continue
		}
		if emitted >= limit {
			break
		}
		b.WriteString(scanner.Text())
		b.WriteByte('\n')
		emitted++
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read %s: %w", cleaned, err)
	}

	result := map[string]any{
		"file_path":   cleaned,
		"offset":      offset,
		"limit":       limit,
		"lines_read":  emitted,
		"total_lines": currentLine,
		"content":     b.String(),
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("encode Read result: %w", err)
	}
	return string(payload), nil
}

// validatePath enforces allowed_roots if configured. Empty allowed_roots
// means project-wide access.
func (t *ReadTool) validatePath(p string) error {
	if len(t.cfg.AllowedRoots) == 0 {
		return nil
	}
	cleaned := filepath.Clean(p)
	for _, root := range t.cfg.AllowedRoots {
		rootClean := filepath.Clean(root)
		if cleaned == rootClean || strings.HasPrefix(cleaned, rootClean+string(filepath.Separator)) {
			return nil
		}
	}
	return fmt.Errorf("path %q is outside the configured allowed_roots", p)
}

func isImagePath(p string) bool {
	switch strings.ToLower(filepath.Ext(p)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".tiff":
		return true
	}
	return false
}

func toInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float32:
		return int(x), true
	case float64:
		return int(x), true
	}
	return 0, false
}
