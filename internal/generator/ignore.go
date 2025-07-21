package generator

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreChecker handles .a2a-ignore file parsing and matching
type IgnoreChecker struct {
	patterns []string
}

// NewIgnoreChecker creates a new ignore checker
func NewIgnoreChecker(outputDir string) (*IgnoreChecker, error) {
	ignoreFile := filepath.Join(outputDir, ".a2a-ignore")

	patterns := []string{}

	if file, err := os.Open(ignoreFile); err == nil {
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				log.Printf("Warning: failed to close .a2a-ignore file: %v", closeErr)
			}
		}()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			patterns = append(patterns, line)
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return &IgnoreChecker{patterns: patterns}, nil
}

// ShouldIgnore checks if a file should be ignored based on .a2a-ignore patterns
func (ic *IgnoreChecker) ShouldIgnore(filePath string) bool {
	normalizedPath := filepath.ToSlash(filePath)

	for _, pattern := range ic.patterns {
		if matched, _ := filepath.Match(pattern, normalizedPath); matched {
			return true
		}

		if strings.Contains(normalizedPath, pattern) {
			return true
		}

		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(normalizedPath, pattern) {
			return true
		}
	}

	return false
}
