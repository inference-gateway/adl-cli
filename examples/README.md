# Example ADL files for the A2A CLI

This directory contains example Agent Definition Language (ADL) files that demonstrate different agent configurations.

## Files

- `weather-agent.yaml` - AI-powered agent with weather tools
- `minimal-agent.yaml` - Minimal agent without AI
- `enterprise-agent.yaml` - Enterprise agent with advanced features

## Usage

Generate an agent from any example:

```bash
# Generate AI-powered weather agent
a2a generate --file examples/weather-agent.yaml --output ./weather-agent

# Generate minimal agent
a2a generate --file examples/minimal-agent.yaml --output ./minimal-agent --template minimal

# Generate enterprise agent
a2a generate --file examples/enterprise-agent.yaml --output ./enterprise-agent --template enterprise
```
