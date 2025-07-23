# Example ADL files for the A2A CLI

This directory contains example Agent Definition Language (ADL) files that demonstrate agent configurations.

## Files

- `minimal-agent.yaml` - Complete agent example with AI capabilities, multiple tools, and enterprise features

## Usage

Generate an agent from the example:

```bash
# Generate agent from example
a2a generate --file examples/minimal-agent.yaml --output ./my-agent

# Validate the example
a2a validate examples/minimal-agent.yaml
```
