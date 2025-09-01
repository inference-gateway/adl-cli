# Example ADL files for the ADL CLI

This directory contains example Agent Definition Language (ADL) files that demonstrate agent configurations.

## Files

- `go-agent.yaml` - Complete Go agent example with AI capabilities, multiple tools, and enterprise features
- `rust-agent.yaml` - Complete Rust agent example with web scraping tools and browser capabilities

## Usage

Generate an agent from the examples:

```bash
# Generate Go agent from example
adl generate --file examples/go-agent.yaml --output ./test-go-agent

# Generate Rust agent from example
adl generate --file examples/rust-agent.yaml --output ./test-rust-agent

# Validate the examples
adl validate examples/go-agent.yaml
adl validate examples/rust-agent.yaml
```
