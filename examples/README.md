# Example ADL files for the ADL CLI

This directory contains example Agent Definition Language (ADL) files that demonstrate agent configurations.

## Files

- `go-agent.yaml` - Complete Go agent example with AI capabilities, multiple tools, and enterprise features
- `go-agent-builtin-tools.yaml` - Workspace assistant that exercises every reserved built-in tool (`read`, `write`, `edit`, `bash`, `fetch`) with sensible whitelists for each
- `rust-agent.yaml` - Complete Rust agent example with web scraping tools and browser capabilities
- `typescript-agent.yaml` - Minimal TypeScript agent built with the TypeScript ADK
- `typescript-agent-tools.yaml` - TypeScript agent with tools, services, and dependency injection
- `typescript-agent-ai.yaml` - AI-powered TypeScript agent with LLM-driven tools
- `vercel-agent.yaml` - TypeScript agent configured for Vercel deployment from source (`spec.deployment.type: vercel`), exercising the edge runtime, regions, function limits, and env vars

## Usage

Generate an agent from the examples:

```bash
# Generate Go agent from example
adl generate --file examples/go-agent.yaml --output ./test-go-agent

# Generate Rust agent from example
adl generate --file examples/rust-agent.yaml --output ./test-rust-agent

# Generate TypeScript agent from example
adl generate --file examples/typescript-agent.yaml --output ./test-typescript-agent

# Validate the examples
adl validate examples/go-agent.yaml
adl validate examples/rust-agent.yaml
adl validate examples/typescript-agent.yaml
```
