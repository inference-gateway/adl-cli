package generator

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreChecker handles .adl-ignore file parsing and matching
type IgnoreChecker struct {
	patterns []string
}

// NewIgnoreChecker creates a new ignore checker
func NewIgnoreChecker(outputDir string) (*IgnoreChecker, error) {
	ignoreFile := filepath.Join(outputDir, ".adl-ignore")

	patterns := []string{}

	if file, err := os.Open(ignoreFile); err == nil {
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				log.Printf("Warning: failed to close .adl-ignore file: %v", closeErr)
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

// ShouldIgnore checks if a file should be ignored based on .adl-ignore patterns
func (ic *IgnoreChecker) ShouldIgnore(filePath string) bool {
	normalizedPath := filepath.ToSlash(filePath)

	for _, pattern := range ic.patterns {
		if strings.HasSuffix(pattern, "/") {
			if strings.HasPrefix(normalizedPath, pattern) {
				return true
			}
			continue
		}

		if strings.Contains(pattern, "*") {
			if matched, _ := filepath.Match(pattern, normalizedPath); matched {
				return true
			}

			if strings.HasSuffix(pattern, "/*") {
				dirPattern := strings.TrimSuffix(pattern, "/*")
				if strings.HasPrefix(normalizedPath, dirPattern+"/") {
					return true
				}
			}
			continue
		}

		if pattern == normalizedPath {
			return true
		}

		if strings.Contains(normalizedPath, pattern) {
			return true
		}
	}

	return false
}
