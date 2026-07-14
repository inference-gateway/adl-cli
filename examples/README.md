# Example ADL files for the ADL CLI

This directory contains example Agent Definition Language (ADL) files that
demonstrate agent configurations across Go, Rust, and TypeScript, plus a range
of deployment targets and optional features.

Every example here is part of the regression suite: `task examples:test`
validates each manifest and `task examples:generate` scaffolds each one into
`test-output/`.

## Files

### Go

- `go-agent.yaml` - Go agent with services (database, notifications, reports),
  multiple tools, and an incident-response orchestration skill
- `go-agent-builtin-tools.yaml` - Workspace assistant that exercises every
  reserved built-in tool (`read`, `write`, `edit`, `bash`, `fetch`) with
  sensible allow-lists for each
- `go-agent-ai.yaml` - AI-powered Go agent with LLM-driven tools and the `infer`
  orchestrator
- `go-agent-artifacts-filesystem.yaml` - Go agent with the filesystem-backed
  artifacts server enabled
- `go-agent-artifacts-minio.yaml` - Go agent with the MinIO-backed artifacts
  server enabled
- `go-agent-telemetry.yaml` - Go agent wired for OpenTelemetry - metrics server,
  OTLP traces, and per-tool-call spans

### Rust

- `rust-agent.yaml` - Minimal Rust agent built with the Rust ADK, including an
  echo tool and a formatting skill
- `rust-agent-ai.yaml` - AI-powered Rust agent with LLM-driven tool skills
- `rust-agent-redis.yaml` - Rust agent with the Redis-backed task queue feature
  enabled

### TypeScript

- `typescript-agent.yaml` - Minimal TypeScript agent built with the TypeScript
  ADK
- `typescript-agent-tools.yaml` - TypeScript agent with tools, services, and
  dependency injection
- `typescript-agent-ai.yaml` - AI-powered TypeScript agent with LLM-driven tools
- `typescript-agent-telemetry.yaml` - TypeScript agent wired for OpenTelemetry
  via the ADK's Node SDK provider

### Deployment targets

- `cloudrun-agent.yaml` - Go agent configured for Google Cloud Run using Google
  Container Registry (`spec.deployment.type: cloudrun`)
- `cloudrun-ghcr-agent.yaml` - Go agent configured for Google Cloud Run using
  GitHub Container Registry
- `cloudflare-agent.yaml` - TypeScript agent configured for Cloudflare Workers
  (`spec.deployment.type: cloudflare`)
- `kubernetes-agent.yaml` - Go agent configured for Kubernetes deployment via the
  inference-gateway operator (`spec.deployment.type: kubernetes`)
- `vercel-agent.yaml` - TypeScript agent configured for Vercel deployment from
  source (`spec.deployment.type: vercel`), exercising the edge runtime, regions,
  function limits, and env vars

## Usage

Validate an example manifest:

```bash
adl validate examples/go-agent.yaml
```

Generate a project from an example:

```bash
# Go agent
adl generate --file examples/go-agent.yaml --output ./test-go-agent

# Rust agent
adl generate --file examples/rust-agent.yaml --output ./test-rust-agent

# TypeScript agent
adl generate --file examples/typescript-agent.yaml --output ./test-typescript-agent
```

Deployment targets are selected by the manifest's `spec.deployment.type`; the
matching `--deployment` flag is optional and simply asserts the expected target:

```bash
# Cloud Run
adl generate --file examples/cloudrun-agent.yaml --output ./test-cloudrun-agent --deployment cloudrun

# Kubernetes
adl generate --file examples/kubernetes-agent.yaml --output ./test-kubernetes-agent --deployment kubernetes
```

Layer on CI and/or CD pipelines with `--ci` and `--cd`:

```bash
adl generate --file examples/go-agent.yaml --output ./enterprise-agent --ci --cd
```

Validate or generate every example at once via the Taskfile:

```bash
task examples:test      # validate all manifests
task examples:generate  # scaffold every example into test-output/
```
