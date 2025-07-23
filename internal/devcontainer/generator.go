package devcontainer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/inference-gateway/a2a-cli/internal/schema"
	"gopkg.in/yaml.v3"
)

// Generator generates devcontainer configurations for A2A agents
type Generator struct{}

// New creates a new devcontainer generator
func New() *Generator {
	return &Generator{}
}

// Generate generates devcontainer configuration files from an ADL file
func (g *Generator) Generate(adlFile, outputDir string) error {
	adl, err := g.parseADL(adlFile)
	if err != nil {
		return fmt.Errorf("failed to parse ADL file: %w", err)
	}

	if err := g.validateADL(adl); err != nil {
		return fmt.Errorf("ADL validation failed: %w", err)
	}

	devcontainerDir := filepath.Join(outputDir, ".devcontainer")
	if err := os.MkdirAll(devcontainerDir, 0755); err != nil {
		return fmt.Errorf("failed to create .devcontainer directory: %w", err)
	}

	if err := g.generateDevcontainerJSON(adl, devcontainerDir); err != nil {
		return fmt.Errorf("failed to generate devcontainer.json: %w", err)
	}

	if err := g.generateDockerfile(adl, devcontainerDir); err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	return nil
}

// parseADL parses an ADL file
func (g *Generator) parseADL(adlFile string) (*schema.ADL, error) {
	data, err := os.ReadFile(adlFile)
	if err != nil {
		return nil, err
	}

	var adl schema.ADL
	if err := yaml.Unmarshal(data, &adl); err != nil {
		return nil, err
	}

	return &adl, nil
}

// validateADL validates the ADL structure for devcontainer generation
func (g *Generator) validateADL(adl *schema.ADL) error {
	if adl.Spec.Language == nil {
		return fmt.Errorf("spec.language is required for devcontainer generation")
	}

	languageCount := 0
	if adl.Spec.Language.Go != nil {
		languageCount++
	}
	if adl.Spec.Language.TypeScript != nil {
		languageCount++
	}

	if languageCount == 0 {
		return fmt.Errorf("at least one programming language must be defined in spec.language")
	}
	if languageCount > 1 {
		return fmt.Errorf("exactly one programming language must be defined for devcontainer generation, found %d", languageCount)
	}

	return nil
}

// generateDevcontainerJSON generates the devcontainer.json configuration
func (g *Generator) generateDevcontainerJSON(adl *schema.ADL, devcontainerDir string) error {
	var image string
	var features map[string]interface{}
	var customizations map[string]interface{}

	// Determine language-specific configuration
	if adl.Spec.Language.Go != nil {
		image = "mcr.microsoft.com/devcontainers/go:1-1.23-bookworm"
		features = map[string]interface{}{
			"ghcr.io/devcontainers/features/docker-in-docker:2": map[string]interface{}{},
			"ghcr.io/devcontainers/features/git:1":              map[string]interface{}{},
		}
		customizations = map[string]interface{}{
			"vscode": map[string]interface{}{
				"extensions": []string{
					"golang.go",
					"ms-vscode.vscode-json",
					"redhat.vscode-yaml",
					"ms-vscode.makefile-tools",
				},
				"settings": map[string]interface{}{
					"go.toolsManagement.checkForUpdates": "local",
					"go.useLanguageServer":                true,
					"go.gopath":                          "/go",
					"go.goroot":                          "/usr/local/go",
				},
			},
		}
	} else if adl.Spec.Language.TypeScript != nil {
		nodeVersion := "18"
		if adl.Spec.Language.TypeScript.NodeVersion != "" {
			nodeVersion = adl.Spec.Language.TypeScript.NodeVersion
		}
		image = fmt.Sprintf("mcr.microsoft.com/devcontainers/typescript-node:%s", nodeVersion)
		features = map[string]interface{}{
			"ghcr.io/devcontainers/features/docker-in-docker:2": map[string]interface{}{},
			"ghcr.io/devcontainers/features/git:1":              map[string]interface{}{},
		}
		customizations = map[string]interface{}{
			"vscode": map[string]interface{}{
				"extensions": []string{
					"ms-vscode.vscode-typescript-next",
					"ms-vscode.vscode-json",
					"redhat.vscode-yaml",
					"bradlc.vscode-tailwindcss",
				},
			},
		}
	}

	devcontainerConfig := map[string]interface{}{
		"name":           fmt.Sprintf("%s Development", adl.Metadata.Name),
		"image":          image,
		"features":       features,
		"customizations": customizations,
		"postCreateCommand": []string{
			"/bin/bash", ".devcontainer/setup.sh",
		},
		"mounts": []string{
			"source=/var/run/docker.sock,target=/var/run/docker.sock,type=bind",
		},
		"remoteUser": "vscode",
	}

	jsonData, err := json.MarshalIndent(devcontainerConfig, "", "  ")
	if err != nil {
		return err
	}

	devcontainerPath := filepath.Join(devcontainerDir, "devcontainer.json")
	if err := os.WriteFile(devcontainerPath, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Generated: %s\n", devcontainerPath)
	return nil
}

// generateDockerfile generates the language-specific Dockerfile
func (g *Generator) generateDockerfile(adl *schema.ADL, devcontainerDir string) error {
	var dockerfileContent string

	if adl.Spec.Language.Go != nil {
		goVersion := "1.23"
		if adl.Spec.Language.Go.Version != "" {
			goVersion = adl.Spec.Language.Go.Version
		}

		dockerfileContent = fmt.Sprintf(`# Development environment for %s A2A Agent
FROM mcr.microsoft.com/devcontainers/go:1-%s-bookworm

# Install additional tools
RUN apt-get update && apt-get install -y \
    curl \
    git \
    make \
    wget \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Install Task (Taskfile)
RUN curl -sL https://taskfile.dev/install.sh | sh

# Install A2A CLI
RUN curl -fsSL https://raw.githubusercontent.com/inference-gateway/a2a-cli/main/install.sh | bash

# Install Go tools for development
RUN go install -v golang.org/x/tools/gopls@latest \
    && go install -v github.com/go-delve/delve/cmd/dlv@latest \
    && go install -v honnef.co/go/tools/cmd/staticcheck@latest

# Set working directory
WORKDIR /workspace

# Copy setup script
COPY setup.sh /tmp/setup.sh
RUN chmod +x /tmp/setup.sh

# Switch to vscode user
USER vscode
`, adl.Metadata.Name, goVersion)

	} else if adl.Spec.Language.TypeScript != nil {
		nodeVersion := "18"
		if adl.Spec.Language.TypeScript.NodeVersion != "" {
			nodeVersion = adl.Spec.Language.TypeScript.NodeVersion
		}

		dockerfileContent = fmt.Sprintf(`# Development environment for %s A2A Agent
FROM mcr.microsoft.com/devcontainers/typescript-node:%s

# Install additional tools
RUN apt-get update && apt-get install -y \
    curl \
    git \
    make \
    wget \
    unzip \
    && rm -rf /var/lib/apt/lists/*

# Install Task (Taskfile)
RUN curl -sL https://taskfile.dev/install.sh | sh

# Install A2A CLI
RUN curl -fsSL https://raw.githubusercontent.com/inference-gateway/a2a-cli/main/install.sh | bash

# Install global npm packages for development
RUN npm install -g \
    typescript \
    ts-node \
    @types/node \
    prettier \
    eslint

# Set working directory
WORKDIR /workspace

# Copy setup script
COPY setup.sh /tmp/setup.sh
RUN chmod +x /tmp/setup.sh

# Switch to node user
USER node
`, adl.Metadata.Name, nodeVersion)
	}

	dockerfilePath := filepath.Join(devcontainerDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return err
	}

	// Also generate setup script
	setupScriptContent := `#!/bin/bash
# Post-create setup script for A2A development environment

echo "🚀 Setting up A2A development environment..."

# Verify A2A CLI installation
if command -v a2a &> /dev/null; then
    echo "✅ A2A CLI is installed: $(a2a --version)"
else
    echo "❌ A2A CLI not found"
fi

# Verify Task installation  
if command -v task &> /dev/null; then
    echo "✅ Task is installed: $(task --version)"
else
    echo "❌ Task not found"
fi

echo "🎉 Development environment setup complete!"
echo "💡 Try running 'a2a --help' to get started"
`

	setupScriptPath := filepath.Join(devcontainerDir, "setup.sh")
	if err := os.WriteFile(setupScriptPath, []byte(setupScriptContent), 0755); err != nil {
		return err
	}

	fmt.Printf("✅ Generated: %s\n", dockerfilePath)
	fmt.Printf("✅ Generated: %s\n", setupScriptPath)
	return nil
}
