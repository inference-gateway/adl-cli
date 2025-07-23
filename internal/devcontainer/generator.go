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
	var remoteUser string
	var customizations map[string]interface{}

	if adl.Spec.Language.Go != nil {
		remoteUser = "vscode"
		customizations = map[string]interface{}{
			"vscode": map[string]interface{}{
				"extensions": []string{
					"golang.go",
					"ms-vscode.vscode-json",
					"redhat.vscode-yaml",
				},
				"settings": map[string]interface{}{
					"terminal.integrated.defaultProfile.linux": "zsh",
					"go.toolsManagement.checkForUpdates":       "local",
					"go.useLanguageServer":                     true,
					"go.gopath":                                "/home/vscode/go",
					"go.goroot":                                "/usr/local/go",
				},
			},
		}
	} else if adl.Spec.Language.TypeScript != nil {
		remoteUser = "node"
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
		"dockerFile":     "Dockerfile",
		"customizations": customizations,
		"mounts": []string{
			"source=/var/run/docker.sock,target=/var/run/docker.sock,type=bind",
		},
		"remoteUser":        remoteUser,
		"postCreateCommand": "go mod tidy",
		"features": map[string]interface{}{
			"ghcr.io/devcontainers/features/docker-in-docker:latest": map[string]interface{}{
				"version": "latest",
			},
		},
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

# Use Powerlevel10k theme
RUN git clone --depth=1 https://github.com/romkatv/powerlevel10k.git /home/vscode/.powerlevel10k

# Configure Powerlevel10k
RUN echo 'source /home/vscode/.powerlevel10k/powerlevel10k.zsh-theme' >> /home/vscode/.zshrc && \
    echo 'POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD=true' >> /home/vscode/.zshrc && \
    echo 'POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(dir vcs)' >> /home/vscode/.zshrc && \
    echo 'POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(command_execution_time status)' >> /home/vscode/.zshrc && \
    echo 'POWERLEVEL9K_COMMAND_EXECUTION_TIME_THRESHOLD=0' >> /home/vscode/.zshrc && \
    echo 'POWERLEVEL9K_COMMAND_EXECUTION_TIME_PRECISION=2' >> /home/vscode/.zshrc && \
    echo 'POWERLEVEL9K_COMMAND_EXECUTION_TIME_FORMAT="duration"' >> /home/vscode/.zshrc

# Set working directory
WORKDIR /workspace

# Create Go directories and set proper ownership
RUN mkdir -p /home/vscode/.cache/go-mod /home/vscode/.cache/go-build /home/vscode/go && \
    chown -R vscode:vscode /home/vscode/.cache /home/vscode/go

# Switch to vscode user
USER vscode

ENV GOPATH=/home/vscode/go
ENV GOMODCACHE=/home/vscode/.cache/go-mod
ENV GOCACHE=/home/vscode/.cache/go-build
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
`, adl.Metadata.Name, goVersion)

	} else if adl.Spec.Language.TypeScript != nil {
		nodeVersion := "20"
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

# Use Powerlevel10k theme
RUN git clone --depth=1 https://github.com/romkatv/powerlevel10k.git /home/node/.powerlevel10k

# Configure Powerlevel10k
RUN echo 'source /home/node/.powerlevel10k/powerlevel10k.zsh-theme' >> /home/node/.zshrc && \
    echo 'POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD=true' >> /home/node/.zshrc && \
    echo 'POWERLEVEL9K_LEFT_PROMPT_ELEMENTS=(dir vcs)' >> /home/node/.zshrc && \
    echo 'POWERLEVEL9K_RIGHT_PROMPT_ELEMENTS=(command_execution_time status)' >> /home/node/.zshrc && \
    echo 'POWERLEVEL9K_COMMAND_EXECUTION_TIME_THRESHOLD=0' >> /home/node/.zshrc && \
    echo 'POWERLEVEL9K_COMMAND_EXECUTION_TIME_PRECISION=2' >> /home/node/.zshrc && \
    echo 'POWERLEVEL9K_COMMAND_EXECUTION_TIME_FORMAT="duration"' >> /home/node/.zshrc

# Set working directory
WORKDIR /workspace

# Switch to node user
USER node
`, adl.Metadata.Name, nodeVersion)
	}

	dockerfilePath := filepath.Join(devcontainerDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		return err
	}

	fmt.Printf("✅ Generated: %s\n", dockerfilePath)
	return nil
}
