# Changelog

All notable changes to this project will be documented in this file.

## [0.20.6](https://github.com/inference-gateway/adl-cli/compare/v0.20.5...v0.20.6) (2025-09-17)

### ♻️ Improvements

* Improve dynamic type handling for configuration fields in Go template ([19c5731](https://github.com/inference-gateway/adl-cli/commit/19c57317b4a31ee94f77b5a90ee2b289c9799102))

### 📚 Documentation

* Document configuration and dependency management in go-agent example ([2122111](https://github.com/inference-gateway/adl-cli/commit/2122111121831cb7591127c32332fd6be649fd56))

## [0.20.5](https://github.com/inference-gateway/adl-cli/compare/v0.20.4...v0.20.5) (2025-09-16)

### ♻️ Improvements

* Improve dependency handling and configuration structure in ADL spec ([339b1ba](https://github.com/inference-gateway/adl-cli/commit/339b1ba05e39357351b0ca8944fb20eb20310add))

## [0.20.4](https://github.com/inference-gateway/adl-cli/compare/v0.20.3...v0.20.4) (2025-09-16)

### 🐛 Bug Fixes

* **gitignore:** Update environment file patterns to include local files and exclude example files ([4a0300c](https://github.com/inference-gateway/adl-cli/commit/4a0300ce90dcf2ecbdd8529539dd0eb5a4bb17c7))

## [0.20.3](https://github.com/inference-gateway/adl-cli/compare/v0.20.2...v0.20.3) (2025-09-16)

### 🐛 Bug Fixes

* **acronyms:** Update ADL acronyms from language - add it to spec ([5f1d6f1](https://github.com/inference-gateway/adl-cli/commit/5f1d6f1bd83437a90e87417e4f0a9473a1524553))

## [0.20.2](https://github.com/inference-gateway/adl-cli/compare/v0.20.1...v0.20.2) (2025-09-16)

### ♻️ Improvements

* **dependencies:** Improve dependency injection support with built-in logger and configuration management ([006b2b7](https://github.com/inference-gateway/adl-cli/commit/006b2b745df4060a938f8af01f8dc1a00f042d7f))

## [0.20.1](https://github.com/inference-gateway/adl-cli/compare/v0.20.0...v0.20.1) (2025-09-16)

### 📚 Documentation

* Improve documentation for dependency injection support in ADL CLI ([c6543cf](https://github.com/inference-gateway/adl-cli/commit/c6543cf1175e8a59cf462479e01249092a41f6d4))

## [0.20.0](https://github.com/inference-gateway/adl-cli/compare/v0.19.14...v0.20.0) (2025-09-16)

### ✨ Features

* **dependencies:** Add dependency injection support for skills ([#75](https://github.com/inference-gateway/adl-cli/issues/75)) ([a16faec](https://github.com/inference-gateway/adl-cli/commit/a16faec9db0115d7ecbb99a7fe65734373334e78))

## [0.19.14](https://github.com/inference-gateway/adl-cli/compare/v0.19.13...v0.19.14) (2025-09-15)

### ♻️ Improvements

* Update adk dependency version to v0.10.1 - remove default configuration values ([#73](https://github.com/inference-gateway/adl-cli/issues/73)) ([e142508](https://github.com/inference-gateway/adl-cli/commit/e14250807e5e4f830aeb73c9196cdf7fa8b30a68))

## [0.19.13](https://github.com/inference-gateway/adl-cli/compare/v0.19.12...v0.19.13) (2025-09-15)

### 🐛 Bug Fixes

* **agent:** Add MaxChatCompletionIterations to agent configuration ([ee8f54a](https://github.com/inference-gateway/adl-cli/commit/ee8f54a1c3c42cca42244e178bb2029c1c838e3e))

## [0.19.12](https://github.com/inference-gateway/adl-cli/compare/v0.19.11...v0.19.12) (2025-09-12)

### 📚 Documentation

* Update README and gitattributes to reflect changes in CloudRun deployment task ([d070236](https://github.com/inference-gateway/adl-cli/commit/d0702360a93c5f334f8ad9eac6003c9d2421eb54))

### 🔧 Miscellaneous

* Delete .github/workflows/claude-code-review.yml ([e1d1a15](https://github.com/inference-gateway/adl-cli/commit/e1d1a1583653641391174bc57546939c585e4a5b))

## [0.19.11](https://github.com/inference-gateway/adl-cli/compare/v0.19.10...v0.19.11) (2025-09-12)

### ♻️ Improvements

* **headers:** Add DO NOT EDIT headers to generated CI/CD workflow files ([#72](https://github.com/inference-gateway/adl-cli/issues/72)) ([5db9169](https://github.com/inference-gateway/adl-cli/commit/5db91691963ec8213b4e30e8676c22994ca95b7d)), closes [#71](https://github.com/inference-gateway/adl-cli/issues/71)

## [0.19.10](https://github.com/inference-gateway/adl-cli/compare/v0.19.9...v0.19.10) (2025-09-12)

### 🐛 Bug Fixes

* **cd.yaml.tmpl:** Add missing spacing when CloudRun or Kubernetes deployment is chosen with cd flag enabled ([5de56a6](https://github.com/inference-gateway/adl-cli/commit/5de56a648b11714a8f3091fcf9f50e8b33e05d86))
* **cd.yaml.tmpl:** Adjust spacing for better readability in the checkout step ([61f007c](https://github.com/inference-gateway/adl-cli/commit/61f007cdd6bca6a7e2828789dc0718d9be1ee55b))

## [0.19.9](https://github.com/inference-gateway/adl-cli/compare/v0.19.8...v0.19.9) (2025-09-12)

### ♻️ Improvements

* **templates:** Add spacing when deploy job is needed ([e9e62a0](https://github.com/inference-gateway/adl-cli/commit/e9e62a0a4d1fecc8d7cce19a2a8529d596bfcdcd))

## [0.19.8](https://github.com/inference-gateway/adl-cli/compare/v0.19.7...v0.19.8) (2025-09-12)

### 🐛 Bug Fixes

* **go.mod.tmpl:** Update adk dependency version to v0.10.0 ([5827a35](https://github.com/inference-gateway/adl-cli/commit/5827a35e081bcfbf19f24ec25efea40272e6b816))

## [0.19.7](https://github.com/inference-gateway/adl-cli/compare/v0.19.6...v0.19.7) (2025-09-08)

### 🐛 Bug Fixes

* **go.mod.tmpl:** Update adk dependency version to v0.9.6 ([f40acfb](https://github.com/inference-gateway/adl-cli/commit/f40acfb5ee095169766fae58ea5dec1b5dfa6ef8))

## [0.19.6](https://github.com/inference-gateway/adl-cli/compare/v0.19.5...v0.19.6) (2025-09-08)

### 🐛 Bug Fixes

* **go.mod.tmpl:** Update adk dependency version to v0.9.4 ([9f077a6](https://github.com/inference-gateway/adl-cli/commit/9f077a60fe50ef917355ac340b1fc503071b3f0f))

## [0.19.5](https://github.com/inference-gateway/adl-cli/compare/v0.19.4...v0.19.5) (2025-09-07)

### 🐛 Bug Fixes

* **go.mod.tmpl:** Update adk dependency version to v0.9.3 ([64f31bb](https://github.com/inference-gateway/adl-cli/commit/64f31bb97111de903ec258506f3fa65a3c978923))

## [0.19.4](https://github.com/inference-gateway/adl-cli/compare/v0.19.3...v0.19.4) (2025-09-06)

### 🐛 Bug Fixes

* **Skills:** Skip "DO NOT EDIT" headers for skill files ([#70](https://github.com/inference-gateway/adl-cli/issues/70)) ([2a9d2a5](https://github.com/inference-gateway/adl-cli/commit/2a9d2a55aa580fd7174e351ed740969fb7c91571)), closes [#69](https://github.com/inference-gateway/adl-cli/issues/69)

## [0.19.3](https://github.com/inference-gateway/adl-cli/compare/v0.19.2...v0.19.3) (2025-09-06)

### ♻️ Improvements

* **templates:** Fix logger injection and syntax error in Go skill template ([#68](https://github.com/inference-gateway/adl-cli/issues/68)) ([96e4695](https://github.com/inference-gateway/adl-cli/commit/96e46951cb0a462fabc35dbf922a51e312b04a38)), closes [#67](https://github.com/inference-gateway/adl-cli/issues/67)

## [0.19.2](https://github.com/inference-gateway/adl-cli/compare/v0.19.1...v0.19.2) (2025-09-06)

### ♻️ Improvements

* **generator:** Convert skill filenames to snake_case and use explicit .adl-ignore paths ([#66](https://github.com/inference-gateway/adl-cli/issues/66)) ([52a2e9d](https://github.com/inference-gateway/adl-cli/commit/52a2e9dd052ca57efe5ea2e22224e49525019dee))

## [0.19.1](https://github.com/inference-gateway/adl-cli/compare/v0.19.0...v0.19.1) (2025-09-06)

### 🐛 Bug Fixes

* Replace non-existent .Type fields in agents.md template ([#64](https://github.com/inference-gateway/adl-cli/issues/64)) ([6f73bdc](https://github.com/inference-gateway/adl-cli/commit/6f73bdc3faf85182d7f04672760012ff1b0c6173))

## [0.19.0](https://github.com/inference-gateway/adl-cli/compare/v0.18.2...v0.19.0) (2025-09-06)

### ✨ Features

* **acronyms:** Add configurable acronyms for improved code generation ([#62](https://github.com/inference-gateway/adl-cli/issues/62)) ([2964c5f](https://github.com/inference-gateway/adl-cli/commit/2964c5f9cd73dc93187605145d55c766f5a6895b))

### 🐛 Bug Fixes

* **templates:** Handle dashes in skill names for Go identifier generation ([#60](https://github.com/inference-gateway/adl-cli/issues/60)) ([44cc232](https://github.com/inference-gateway/adl-cli/commit/44cc23280e960333f63d678b7d38ebd6d2d43715)), closes [#59](https://github.com/inference-gateway/adl-cli/issues/59)

## [0.18.2](https://github.com/inference-gateway/adl-cli/compare/v0.18.1...v0.18.2) (2025-09-05)

### ♻️ Improvements

* Revise project overview and development guidelines in CLAUDE.md ([#58](https://github.com/inference-gateway/adl-cli/issues/58)) ([55ef0db](https://github.com/inference-gateway/adl-cli/commit/55ef0db945efc9f81339d0779fd6065a37dd085f))

### 📚 Documentation

* Update Go version requirement to 1.24 in contributing guide ([fe0c38a](https://github.com/inference-gateway/adl-cli/commit/fe0c38a66af3a071d94a737a906966202a4a8e32))

## [0.18.1](https://github.com/inference-gateway/adl-cli/compare/v0.18.0...v0.18.1) (2025-09-05)

### ♻️ Improvements

* **docs:** Update deployment section ([7056f94](https://github.com/inference-gateway/adl-cli/commit/7056f942ff95adbab6d3d153b30c0e6a28fb1707))
* **templates:** Clean up feature request template ([bfb1c31](https://github.com/inference-gateway/adl-cli/commit/bfb1c3112aacb1b0055eb77eecf2918bddf3f2c1))
* **templates:** Refactor request template cleanup ([791e432](https://github.com/inference-gateway/adl-cli/commit/791e43289dc59d3770da6e6486ffa1a0516daa8c))

### 📚 Documentation

* Update bug_report template with placeholders ([a766f3e](https://github.com/inference-gateway/adl-cli/commit/a766f3e47e4df9af896bb77acb595cee307d20a9))

## [0.18.0](https://github.com/inference-gateway/adl-cli/compare/v0.17.0...v0.18.0) (2025-09-05)

### ✨ Features

* **deployment:** Add CloudRun deployment support to ADL CLI ([#55](https://github.com/inference-gateway/adl-cli/issues/55)) ([5d3f549](https://github.com/inference-gateway/adl-cli/commit/5d3f5498849e4705104698bb6ec67187a5de8a47)), closes [#52](https://github.com/inference-gateway/adl-cli/issues/52)

### ♻️ Improvements

* **templates:** Simplify GitHub issue templates for better flexibility ([#57](https://github.com/inference-gateway/adl-cli/issues/57)) ([025b6bd](https://github.com/inference-gateway/adl-cli/commit/025b6bd798da00a81b3dcd4f62e3183424098a21)), closes [#56](https://github.com/inference-gateway/adl-cli/issues/56)
* **templates:** Use GitHub repository name variable in CD workflow ([#54](https://github.com/inference-gateway/adl-cli/issues/54)) ([f493c63](https://github.com/inference-gateway/adl-cli/commit/f493c63ffafa504ad0afff7310f5600f532fb84c)), closes [#53](https://github.com/inference-gateway/adl-cli/issues/53)

## [0.17.0](https://github.com/inference-gateway/adl-cli/compare/v0.16.0...v0.17.0) (2025-09-05)

### ✨ Features

* **issue-templates:** Add GitHub issue templates generation support ([#51](https://github.com/inference-gateway/adl-cli/issues/51)) ([ff79602](https://github.com/inference-gateway/adl-cli/commit/ff79602573ff7bf1ee173c6f09fd42021ae5d4a7)), closes [#47](https://github.com/inference-gateway/adl-cli/issues/47)

### 🐛 Bug Fixes

* **templates:** Only include skills import when skills are defined ([#48](https://github.com/inference-gateway/adl-cli/issues/48)) ([1974df0](https://github.com/inference-gateway/adl-cli/commit/1974df006616de43a77a4e1efa9528ff909520d6)), closes [#46](https://github.com/inference-gateway/adl-cli/issues/46)
* **templates:** Preserve original generate command flags in Taskfile.yml ([#50](https://github.com/inference-gateway/adl-cli/issues/50)) ([1da3d8b](https://github.com/inference-gateway/adl-cli/commit/1da3d8b66a43214fafff3ef6ef4b2ba6735e278a)), closes [#49](https://github.com/inference-gateway/adl-cli/issues/49)

## [0.16.0](https://github.com/inference-gateway/adl-cli/compare/v0.15.0...v0.16.0) (2025-09-05)

### ✨ Features

* **cd:** Add GitHub App support for CD execution ([#41](https://github.com/inference-gateway/adl-cli/issues/41)) ([855d8ea](https://github.com/inference-gateway/adl-cli/commit/855d8ea25f32f8466a03996403cf6725e25f7e1c)), closes [#37](https://github.com/inference-gateway/adl-cli/issues/37)
* **generate:** Add --ai flag to generate AI assistant instructions ([#38](https://github.com/inference-gateway/adl-cli/issues/38)) ([ee5fa01](https://github.com/inference-gateway/adl-cli/commit/ee5fa01ffa99e2d226f75c4cee32e000a18176df)), closes [#11](https://github.com/inference-gateway/adl-cli/issues/11)

### ♻️ Improvements

* Delete .env.example template ([#42](https://github.com/inference-gateway/adl-cli/issues/42)) ([436efdd](https://github.com/inference-gateway/adl-cli/commit/436efdd2b622f91aff66464c8245a3be24b1c01d))
* **init:** Make default AI provider and model vendor-neutral ([#45](https://github.com/inference-gateway/adl-cli/issues/45)) ([149e057](https://github.com/inference-gateway/adl-cli/commit/149e0575164054aaa2f4ded17ba663ece97a4b5a))

### 🐛 Bug Fixes

* **init:** Add default SCM configuration when using --defaults flag ([#44](https://github.com/inference-gateway/adl-cli/issues/44)) ([01ce420](https://github.com/inference-gateway/adl-cli/commit/01ce420c9ef7db7de352290aceaf3827142c1e09)), closes [#43](https://github.com/inference-gateway/adl-cli/issues/43)
* **templates:** Remove double escaping of GITHUB_OUTPUT in CD template ([#40](https://github.com/inference-gateway/adl-cli/issues/40)) ([4997dd3](https://github.com/inference-gateway/adl-cli/commit/4997dd3164fb2682fb3885b780c0421c46f6a274)), closes [#36](https://github.com/inference-gateway/adl-cli/issues/36)

## [0.15.0](https://github.com/inference-gateway/adl-cli/compare/v0.14.6...v0.15.0) (2025-09-03)

### ✨ Features

* Implement CD scaffolding with semantic-release ([#35](https://github.com/inference-gateway/adl-cli/issues/35)) ([8c5df8d](https://github.com/inference-gateway/adl-cli/commit/8c5df8d09008e29f529aca307dcdc0c5cf7545b3)), closes [#34](https://github.com/inference-gateway/adl-cli/issues/34)

### 📚 Documentation

* Update AI provider list and installation instructions in README ([2b88445](https://github.com/inference-gateway/adl-cli/commit/2b884454536aaaf5f6e2f9136ae36cb761274271))
* Update Go version requirement to 1.24 in README ([a1d8451](https://github.com/inference-gateway/adl-cli/commit/a1d8451980a8af9f5cd02f36f4543bd3a614beec))

## [0.14.6](https://github.com/inference-gateway/adl-cli/compare/v0.14.5...v0.14.6) (2025-09-03)

### ♻️ Improvements

* **docs:** Add server timeout and health check logging configuration options to README ([ed104ce](https://github.com/inference-gateway/adl-cli/commit/ed104ceadac4fd92e0b36190d408c44bac1ff3e8))

## [0.14.5](https://github.com/inference-gateway/adl-cli/compare/v0.14.4...v0.14.5) (2025-09-03)

### 🐛 Bug Fixes

* **template:** Add golangci-lint configuration for Go language support in manifest.toml.tmpl ([4e8ee24](https://github.com/inference-gateway/adl-cli/commit/4e8ee24542ed76a98810ba89c3257f644ebaa140))

## [0.14.4](https://github.com/inference-gateway/adl-cli/compare/v0.14.3...v0.14.4) (2025-09-03)

### ♻️ Improvements

* **template:** Pass logger to skill initialization in main.go.tmpl ([fbcb8ef](https://github.com/inference-gateway/adl-cli/commit/fbcb8ef4bd0707e471eee42def4d2736605c86bd))

## [0.14.3](https://github.com/inference-gateway/adl-cli/compare/v0.14.2...v0.14.3) (2025-09-03)

### 🐛 Bug Fixes

* **tmpl:** Add LLM client creation and error handling in main.go.tmpl ([a002e08](https://github.com/inference-gateway/adl-cli/commit/a002e0800ab509fd9b3dd1baf876bcb3322d2888))

## [0.14.2](https://github.com/inference-gateway/adl-cli/compare/v0.14.1...v0.14.2) (2025-09-03)

### ♻️ Improvements

* **tmpl:** Improve logging of loaded configuration in main.go.tmpl ([b89d715](https://github.com/inference-gateway/adl-cli/commit/b89d715282cd82904957839e7762623f3473c196))

## [0.14.1](https://github.com/inference-gateway/adl-cli/compare/v0.14.0...v0.14.1) (2025-09-03)

### 🐛 Bug Fixes

* **tmpl:** Set default value for A2A_AGENT_CLIENT_SYSTEM_PROMPT in env.example.tmpl ([0991744](https://github.com/inference-gateway/adl-cli/commit/0991744fef1864e52ff37d6d14076e68dc37a295))

## [0.14.0](https://github.com/inference-gateway/adl-cli/compare/v0.13.3...v0.14.0) (2025-09-03)

### ✨ Features

* **docs:** Add Docker build instructions and available build arguments to README ([1b5c0d2](https://github.com/inference-gateway/adl-cli/commit/1b5c0d2e023c835401dd8509ce7802f8ff6feada))

### 🐛 Bug Fixes

* **docs:** Update A2A Debugger commands in README for correct usage and network settings ([07ea90b](https://github.com/inference-gateway/adl-cli/commit/07ea90b4221a4bfccfaad773efbb0dce77f2e4a1))

## [0.13.3](https://github.com/inference-gateway/adl-cli/compare/v0.13.2...v0.13.3) (2025-09-03)

### 📚 Documentation

* Update terminology from "A2A Framework" to "A2A" in README and configuration templates ([4c058c8](https://github.com/inference-gateway/adl-cli/commit/4c058c8de8f9dd329da62004e6ff54070e350d15))

## [0.13.2](https://github.com/inference-gateway/adl-cli/compare/v0.13.1...v0.13.2) (2025-09-03)

### ♻️ Improvements

* Remove Docker deployment instructions from README - bad practice, you don't deploy with docker ([bc86a36](https://github.com/inference-gateway/adl-cli/commit/bc86a360ef7c63deeebaddc347ed095da4f18935))

### 📚 Documentation

* Add debugging section with A2A Debugger usage instructions to README ([86f4f71](https://github.com/inference-gateway/adl-cli/commit/86f4f71b5312123df01c59e174ca7efa1c9a0d4d))

## [0.13.1](https://github.com/inference-gateway/adl-cli/compare/v0.13.0...v0.13.1) (2025-09-03)

### 🐛 Bug Fixes

* Remove optional inference gateway URL from env.example template ([83fb697](https://github.com/inference-gateway/adl-cli/commit/83fb697d7d26956f4ab56a5931c1506a827b6f10))

## [0.13.0](https://github.com/inference-gateway/adl-cli/compare/v0.12.3...v0.13.0) (2025-09-03)

### ✨ Features

* Add env.example template for configuration and update registry mappings ([46bc43a](https://github.com/inference-gateway/adl-cli/commit/46bc43a12d0a90331c820241e2c3ddc17c71878a))

### ♻️ Improvements

* Update agent configuration and improve README with environment variable details ([f0cdf36](https://github.com/inference-gateway/adl-cli/commit/f0cdf367457a74b024a0be7c9deb03d2d197744f))

## [0.12.3](https://github.com/inference-gateway/adl-cli/compare/v0.12.2...v0.12.3) (2025-09-02)

### 🐛 Bug Fixes

* Correct command in taskfile to use 'adl' instead of 'a2a' for code generation ([e1cdd40](https://github.com/inference-gateway/adl-cli/commit/e1cdd40fab59113229940cc6416574c49a5e94f4))
* Improve agent card generation with default values and protocol handling ([61b4e32](https://github.com/inference-gateway/adl-cli/commit/61b4e32e1f9e18912724353b8a0fd2571d39f8b0))
* Reorder Go post-generation commands for improved execution flow ([56a6a7d](https://github.com/inference-gateway/adl-cli/commit/56a6a7d9a2159daca55959726ebc9c5e96ef0050))

## [0.12.2](https://github.com/inference-gateway/adl-cli/compare/v0.12.1...v0.12.2) (2025-09-02)

### 🐛 Bug Fixes

* Update ignore patterns from 'tools' to 'skills' in generator and tests ([cbb30b7](https://github.com/inference-gateway/adl-cli/commit/cbb30b74a893a9b706fba396c69c1d405e17e9a8))

### 📚 Documentation

* Add card configuration to go-agent and rust-agent YAML files ([aa171ba](https://github.com/inference-gateway/adl-cli/commit/aa171ba138123fb202d7e7364b396bd4ba8104b1))

## [0.12.1](https://github.com/inference-gateway/adl-cli/compare/v0.12.0...v0.12.1) (2025-09-02)

### ♻️ Improvements

* Clean up whitespace in skill description prompt and card struct fields ([ef55a43](https://github.com/inference-gateway/adl-cli/commit/ef55a431d5f416f103e1401dd842e2b4cd89e592))
* Rename tools to skills in documentation and code templates ([9b21eeb](https://github.com/inference-gateway/adl-cli/commit/9b21eeb9c7042b3f1d3781af98becc7a8a3511d2))

### 🐛 Bug Fixes

* Update AI provider options in init command for accuracy ([b4d8481](https://github.com/inference-gateway/adl-cli/commit/b4d8481bc8eabaef6316b474b1f244d7c6e33512))

### ✅ Miscellaneous

* Update test cases to replace 'tools' with 'skills' and adjust README for agent configuration ([30c978e](https://github.com/inference-gateway/adl-cli/commit/30c978e39f28bf9f0ab038a0c5b7bfc0b477e018))

## [0.12.0](https://github.com/inference-gateway/adl-cli/compare/v0.11.0...v0.12.0) (2025-09-02)

### ✨ Features

* **docs:** Improve README with structured skills and configuration tables ([6f502a0](https://github.com/inference-gateway/adl-cli/commit/6f502a0405aa561d29b75542e182f237a1699eb2))

### ♻️ Improvements

* Refactor tools to skills and enhance schema with tags and IDs ([31939b8](https://github.com/inference-gateway/adl-cli/commit/31939b8d1aba3717d525845d554c27f515685115))

## [0.11.0](https://github.com/inference-gateway/adl-cli/compare/v0.10.3...v0.11.0) (2025-09-02)

### ✨ Features

* **generator:** Add GenerateCI field to context and update README template for CI badge ([b17a64e](https://github.com/inference-gateway/adl-cli/commit/b17a64e02b5cfdc68edc1afe2c66b0fa6877f35b))

## [0.10.3](https://github.com/inference-gateway/adl-cli/compare/v0.10.2...v0.10.3) (2025-09-02)

### 🐛 Bug Fixes

* **dockerfile:** Update Go base image to 1.24-alpine and remove A2A_SERVER_HOST environment variable ([316963c](https://github.com/inference-gateway/adl-cli/commit/316963ce3e6ffd05e0545fb9ad36f204cf7026f7))

## [0.10.2](https://github.com/inference-gateway/adl-cli/compare/v0.10.1...v0.10.2) (2025-09-02)

### 🐛 Bug Fixes

* **go.mod:** Update adk dependency version to v0.9.2 ([5babaf3](https://github.com/inference-gateway/adl-cli/commit/5babaf320a3908cf84b40d6fc162c216bbc595db))

## [0.10.1](https://github.com/inference-gateway/adl-cli/compare/v0.10.0...v0.10.1) (2025-09-02)

### 🐛 Bug Fixes

* **main:** Correct path for agent card file in A2A server initialization ([d639a10](https://github.com/inference-gateway/adl-cli/commit/d639a10718d73f06a8be336a2a5ec9a36b75f31a))

## [0.10.0](https://github.com/inference-gateway/adl-cli/compare/v0.9.1...v0.10.0) (2025-09-02)

### ✨ Features

* **server:** Add default task handlers for A2A server ([532ea20](https://github.com/inference-gateway/adl-cli/commit/532ea20a6532c63f63826fc3ce64ba11095498d9))

### ♻️ Improvements

* **schema:** Replace map[string]interface{} with map[string]any for improved type handling ([131071e](https://github.com/inference-gateway/adl-cli/commit/131071e8f28f13cec8be58e7f827f4ee39df14f8))

## [0.9.1](https://github.com/inference-gateway/adl-cli/compare/v0.9.0...v0.9.1) (2025-09-02)

### ♻️ Improvements

* **generator:** Add toPascalCase function for improved tool name formatting ([a819a86](https://github.com/inference-gateway/adl-cli/commit/a819a86398ec51df1f03aa381c15938bb4157ef0))

### 📚 Documentation

* **hooks:** Improve formatting in post-generation hook comments ([a96deae](https://github.com/inference-gateway/adl-cli/commit/a96deae4234e1fc7e98fff8333ebbb44903a9183))

## [0.9.0](https://github.com/inference-gateway/adl-cli/compare/v0.8.1...v0.9.0) (2025-09-02)

### ✨ Features

* **generator:** Implement post-generation hooks for custom commands ([0148ccf](https://github.com/inference-gateway/adl-cli/commit/0148ccf74cccb8260ff331e364760dd93bd26957))

### 📚 Documentation

* **readme:** Update Table of Contents with additional sections ([b2ee05a](https://github.com/inference-gateway/adl-cli/commit/b2ee05ae92ce8a179a191cb5a31f9e77bd9d9d2c))

## [0.8.1](https://github.com/inference-gateway/adl-cli/compare/v0.8.0...v0.8.1) (2025-09-02)

### 🐛 Bug Fixes

* **template:** Correctly format required fields in tool schema ([7374038](https://github.com/inference-gateway/adl-cli/commit/7374038b85636262cb5d64e871740d7c85356401))

## [0.8.0](https://github.com/inference-gateway/adl-cli/compare/v0.7.9...v0.8.0) (2025-09-02)

### ✨ Features

* **generator:** Add Go file formatting after project generation ([6fd5e7d](https://github.com/inference-gateway/adl-cli/commit/6fd5e7d4c94ce38838885e6f0ae589a0d6bf0bfc))

### 🐛 Bug Fixes

* **template:** Correctly format required fields in tool schema ([80351a6](https://github.com/inference-gateway/adl-cli/commit/80351a6e3797f5c2aff84644c17cf2c04a2d83ce))

## [0.7.9](https://github.com/inference-gateway/adl-cli/compare/v0.7.8...v0.7.9) (2025-09-02)

### 🐛 Bug Fixes

* **template:** Update type declarations to use map[string]any for consistency ([225b918](https://github.com/inference-gateway/adl-cli/commit/225b918e6f6383f701ff34f957787f2194a06601))

## [0.7.8](https://github.com/inference-gateway/adl-cli/compare/v0.7.7...v0.7.8) (2025-09-02)

### 🐛 Bug Fixes

* **main:** Remove host logging from A2A server startup ([c295a14](https://github.com/inference-gateway/adl-cli/commit/c295a145a9bd75e14b12b941e3db2ee8a84592bb))
* **template:** Add blank line before tool registration for improved readability ([916ae05](https://github.com/inference-gateway/adl-cli/commit/916ae05fbe3383aebf8c2e327e1fccaa08c733d7))
* **template:** Update property type declaration to use map[string]any for consistency ([b8f2b17](https://github.com/inference-gateway/adl-cli/commit/b8f2b17ee58d6f4bbcbe5f23bfb7d2d46f609ecc))

## [0.7.7](https://github.com/inference-gateway/adl-cli/compare/v0.7.6...v0.7.7) (2025-09-02)

### ♻️ Improvements

* **editorconfig:** Rearrange language-specific sections for better clarity and organization ([9b9a780](https://github.com/inference-gateway/adl-cli/commit/9b9a7800c978b31954bc033b62692c3f41d4ef17))

### 🐛 Bug Fixes

* **template:** Update Go installation configuration in manifest template ([f80751c](https://github.com/inference-gateway/adl-cli/commit/f80751c279327777fe2e516d1f96f86ca6562495))
* **validator:** Improve ADL schema validation with additional properties and requirements ([6316b39](https://github.com/inference-gateway/adl-cli/commit/6316b3985f556e07d4bb6d2997a7046b347a22a9))

## [0.7.6](https://github.com/inference-gateway/adl-cli/compare/v0.7.5...v0.7.6) (2025-09-02)

### 🐛 Bug Fixes

* **validator:** Update provider enum to include additional options ([e22ec05](https://github.com/inference-gateway/adl-cli/commit/e22ec057ff9896854dbd1c3edff3d55e810ea943))

## [0.7.5](https://github.com/inference-gateway/adl-cli/compare/v0.7.4...v0.7.5) (2025-09-02)

### 🐛 Bug Fixes

* **generator:** Remove unnecessary agent provider validation in ADL ([b405ef4](https://github.com/inference-gateway/adl-cli/commit/b405ef49a0dab776068d189ad3b6bd900624c28a))

## [0.7.4](https://github.com/inference-gateway/adl-cli/compare/v0.7.3...v0.7.4) (2025-09-02)

### 🐛 Bug Fixes

* **release:** Set GORELEASER_CURRENT_TAG for versioning in release command ([addc2fe](https://github.com/inference-gateway/adl-cli/commit/addc2fe837e1fa8702e5938f4cc5d10674c44aeb))

### 🔧 Miscellaneous

* **release:** Update git plugin configuration for changelog and Taskfile assets ([61bdc4a](https://github.com/inference-gateway/adl-cli/commit/61bdc4a4202818f49c3045f991d41c01f14caed1))
* Update version to 0.7.3 and remove snapshot flag from release command ([d7a004f](https://github.com/inference-gateway/adl-cli/commit/d7a004f19d55c10ba0e42afdd6170668797614a3))

## [0.7.3](https://github.com/inference-gateway/adl-cli/compare/v0.7.2...v0.7.3) (2025-09-02)

### ♻️ Improvements

* **commands:** Reconcile init and generate command options ([#33](https://github.com/inference-gateway/adl-cli/issues/33)) ([7937bf7](https://github.com/inference-gateway/adl-cli/commit/7937bf7fd09d5c204f9a9c37fe5a076f4fdd08fb)), closes [#32](https://github.com/inference-gateway/adl-cli/issues/32)

### 🔧 Miscellaneous

* Update manifest and Taskfile for goreleaser and go-task integration ([6d22b64](https://github.com/inference-gateway/adl-cli/commit/6d22b64dd084d5cfb3e2d8058719104235687370))

## [0.7.2](https://github.com/inference-gateway/adl-cli/compare/v0.7.1...v0.7.2) (2025-09-02)

### ♻️ Improvements

* Improve release process ([#31](https://github.com/inference-gateway/adl-cli/issues/31)) ([ea2bf45](https://github.com/inference-gateway/adl-cli/commit/ea2bf45183174db9106ea4af9a97125aea9ae08a)), closes [#30](https://github.com/inference-gateway/adl-cli/issues/30)

## [0.7.1](https://github.com/inference-gateway/adl-cli/compare/v0.7.0...v0.7.1) (2025-09-02)

### ♻️ Improvements

* Remove redundant comments in main.go.tmpl for cleaner code ([44385dc](https://github.com/inference-gateway/adl-cli/commit/44385dc1ddf7304119015dc9d624cc4148408748))

### 🐛 Bug Fixes

* **schema:** Make agent.spec.provider and agent.spec.model optional fields ([#29](https://github.com/inference-gateway/adl-cli/issues/29)) ([b52d515](https://github.com/inference-gateway/adl-cli/commit/b52d51573697143bc2f0e76bc0d7ee1c8481a1fd)), closes [#28](https://github.com/inference-gateway/adl-cli/issues/28)

### 📚 Documentation

* Update A2A protocol link in README.md.tmpl to point to the correct repository ([a55b5f9](https://github.com/inference-gateway/adl-cli/commit/a55b5f9a0a33b20bc0be7eca36a46db569c2536f))

## [0.7.0](https://github.com/inference-gateway/adl-cli/compare/v0.6.0...v0.7.0) (2025-09-02)

### ✨ Features

* Add template registry and language-specific templates for Go, Rust, and TypeScript ([e4bb4f5](https://github.com/inference-gateway/adl-cli/commit/e4bb4f50f45f2c7c563b51b1cb2ec9e183ca4e5a))
* **deployment:** Add optional deployment configuration ([#21](https://github.com/inference-gateway/adl-cli/issues/21)) ([c76f9aa](https://github.com/inference-gateway/adl-cli/commit/c76f9aac34b209a764a94f8a35ad06c65ae205b3)), closes [#19](https://github.com/inference-gateway/adl-cli/issues/19)
* **rust:** Implement missing Rust ADK templates and fix API patterns ([#18](https://github.com/inference-gateway/adl-cli/issues/18)) ([32a19eb](https://github.com/inference-gateway/adl-cli/commit/32a19eb577aa8eb93cdf7143a8e5d229276bac7b))

### ♻️ Improvements

* **cli:** Remove --devcontainer flag ([#22](https://github.com/inference-gateway/adl-cli/issues/22)) ([4262e96](https://github.com/inference-gateway/adl-cli/commit/4262e96efbb35e84c22c39450e1209ce1caa85ca))
* Refactor sandbox configurations for extensible multi-environment support ([#27](https://github.com/inference-gateway/adl-cli/issues/27)) ([60cc096](https://github.com/inference-gateway/adl-cli/commit/60cc09696e521e8d6e971e729ab7fc3347d16d54))

### 🐛 Bug Fixes

* Remove unnecessary blank lines in multiple files for cleaner code ([3b2a475](https://github.com/inference-gateway/adl-cli/commit/3b2a4757e3a8c5c3672e28ccfe4ba979c4461364))

### 📚 Documentation

* Clarify YAML ADL usage in README ([#26](https://github.com/inference-gateway/adl-cli/issues/26)) ([b6513d2](https://github.com/inference-gateway/adl-cli/commit/b6513d2afb5346ac4279314f9d626783e84b2d00))
* Reconcile README with improved feature documentation ([#24](https://github.com/inference-gateway/adl-cli/issues/24)) ([16a8c0c](https://github.com/inference-gateway/adl-cli/commit/16a8c0c31735e96d3fdc7bf6efb3937cbc9720da)), closes [#23](https://github.com/inference-gateway/adl-cli/issues/23)
* Update CLAUDE.md and README.md for improved template system and usage examples ([8eb2328](https://github.com/inference-gateway/adl-cli/commit/8eb2328a0a43578e46fd6b45b698fbdcac4d2d13))

### 🔧 Miscellaneous

* Add golangci-lint package support and improve flag binding error handling ([6a07aab](https://github.com/inference-gateway/adl-cli/commit/6a07aaba25290bff063509d1ae9c2f5e91d94358))
* Add issue templates for bug reports, feature requests, and refactor requests ([a473ca5](https://github.com/inference-gateway/adl-cli/commit/a473ca5c16027b2ea0a99b8665c15962d5497117))
* Improve Rust templates and add support for additional files in flox environment ([a7c7869](https://github.com/inference-gateway/adl-cli/commit/a7c7869e545473a14da90fe574529aa2f10d0880))

## [0.6.0](https://github.com/inference-gateway/adl-cli/compare/v0.5.0...v0.6.0) (2025-09-01)

### ✨ Features

* Add support for using an existing ADL schema file during project initialization ([c4b6638](https://github.com/inference-gateway/adl-cli/commit/c4b6638a2d5a1c63565acd22726ef877bc258fc5))
* Improve ADL templates with TypeScript and Rust support in gitattributes ([56c728b](https://github.com/inference-gateway/adl-cli/commit/56c728bddcb5aeed747dcd8b714aade7a189af38))
* Improve project initialization with directory prompts and overwrite option ([273ded1](https://github.com/inference-gateway/adl-cli/commit/273ded18c4ce19834ec4e478c5b23b45b191fddd))
* **rust:** Add Rust language support using Rust-ADK ([#16](https://github.com/inference-gateway/adl-cli/issues/16)) ([59d2674](https://github.com/inference-gateway/adl-cli/commit/59d2674105720d858346fa6a8804c5762b6d6b7d))

### ♻️ Improvements

* Improve Go module prompt with default value from Git remote URL ([ec208c6](https://github.com/inference-gateway/adl-cli/commit/ec208c61349321cb0728cbec5d5cf57a00e9ecf2))
* Refactor input handling by removing scanner dependency and adding readline support ([6d47f0d](https://github.com/inference-gateway/adl-cli/commit/6d47f0d64abdf4a49615ee3d4b555d054144daa8))

### 🐛 Bug Fixes

* Improve error handling on readline closure in prompt functions ([4bacd91](https://github.com/inference-gateway/adl-cli/commit/4bacd9187d80c3a2fd043bdc8769bd754670d09f))

### 🔧 Miscellaneous

* Add initial configuration files for flox environment ([0e826e2](https://github.com/inference-gateway/adl-cli/commit/0e826e2dbec692f08183770ea535d84b8b72d816))
* Update Go module dependency from a2a to adk v0.9.0 ([eae1973](https://github.com/inference-gateway/adl-cli/commit/eae19736089db717704e2abac8800aa4d00ab7ac))

## [0.5.0](https://github.com/inference-gateway/adl-cli/compare/v0.4.3...v0.5.0) (2025-07-30)

### ✨ Features

* Add generate task to create code from ADL ([a3d37bb](https://github.com/inference-gateway/adl-cli/commit/a3d37bbc316dcb5a547284c26b44150621f1ae5b))

### ♻️ Improvements

* Cleanup - update generated file patterns in gitattributes and generator ([456818d](https://github.com/inference-gateway/adl-cli/commit/456818d681d9d1f1ac943c571c0838101457e733))
* General Improvements ([#12](https://github.com/inference-gateway/adl-cli/issues/12)) ([1e892f2](https://github.com/inference-gateway/adl-cli/commit/1e892f2fc02b5f86b5810902f9f870eb791ba66e)), closes [#14](https://github.com/inference-gateway/adl-cli/issues/14)

## [0.5.0-rc.2](https://github.com/inference-gateway/adl-cli/compare/v0.5.0-rc.1...v0.5.0-rc.2) (2025-07-30)

### 🐛 Bug Fixes

* Remove default value from ENVIRONMENT variable documentation ([8f55f13](https://github.com/inference-gateway/adl-cli/commit/8f55f1358c2b564725065b81b676029c952523f3))
* Update TASK_VERSION to v3.44.1 in Dockerfile ([2eb9aa3](https://github.com/inference-gateway/adl-cli/commit/2eb9aa3f2142d43c23def0aeea8ef5e19cb8c79b))

### 👷 CI

* Update repository name from a2a-cli to adl-cli in release workflow ([d37fa54](https://github.com/inference-gateway/adl-cli/commit/d37fa54be184f054790496e50414cd1633c9f826))

### 🔧 Miscellaneous

* Rename project from A2A CLI to ADL CLI, updating all references, file names, and documentation accordingly. Adjusted build scripts, configuration files, and generated files to reflect the new project name and API version. Updated ignore file handling and validation logic to ensure compatibility with the new naming conventions. ([03f4cf7](https://github.com/inference-gateway/adl-cli/commit/03f4cf7da2c5539e3568463faf531a971c00e658))

## [0.5.0-rc.1](https://github.com/inference-gateway/a2a-cli/compare/v0.4.3...v0.5.0-rc.1) (2025-07-23)

### ✨ Features

* Add generate task to create code from ADL ([a3d37bb](https://github.com/inference-gateway/a2a-cli/commit/a3d37bbc316dcb5a547284c26b44150621f1ae5b))

### ♻️ Improvements

* Cleanup - update generated file patterns in gitattributes and generator ([456818d](https://github.com/inference-gateway/a2a-cli/commit/456818d681d9d1f1ac943c571c0838101457e733))
* Enhance configuration management and logging in main application ([8b65915](https://github.com/inference-gateway/a2a-cli/commit/8b65915150fc5773e221b479d57ad817729d1294))

## [0.4.3](https://github.com/inference-gateway/a2a-cli/compare/v0.4.2...v0.4.3) (2025-07-23)

### 🐛 Bug Fixes

* Update Go paths and environment variables in devcontainer configuration ([73e7626](https://github.com/inference-gateway/a2a-cli/commit/73e76268a387e70cc496f8502bd4d82258228c9d))

## [0.4.2](https://github.com/inference-gateway/a2a-cli/compare/v0.4.1...v0.4.2) (2025-07-23)

### 🐛 Bug Fixes

* **ci:** Update installation step from GoReleaser to golangci-lint in workflow ([654f9e0](https://github.com/inference-gateway/a2a-cli/commit/654f9e07a0d112863afb06b133c59c293a768991))

## [0.4.1](https://github.com/inference-gateway/a2a-cli/compare/v0.4.0...v0.4.1) (2025-07-23)

### ♻️ Improvements

* Cleanups ([93a26fe](https://github.com/inference-gateway/a2a-cli/commit/93a26fefb13219142d40d49545426c6ee323c774))

## [0.4.0](https://github.com/inference-gateway/a2a-cli/compare/v0.3.0...v0.4.0) (2025-07-23)

### ✨ Features

* **ci:** Add --ci flag to generate GitHub Actions workflows ([#10](https://github.com/inference-gateway/a2a-cli/issues/10)) ([b6bc96b](https://github.com/inference-gateway/a2a-cli/commit/b6bc96b0a88b7a432f8ba00d8c9e4d1416283296)), closes [#8](https://github.com/inference-gateway/a2a-cli/issues/8)

### 🐛 Bug Fixes

* Add support for 'deepseek' AI provider in project initialization and validation ([a5204c6](https://github.com/inference-gateway/a2a-cli/commit/a5204c60e49d9d21d99406de886d5b619c7b34d8))

### 🔧 Miscellaneous

* Update model in minimal-agent.yaml to 'deepseek-chat' ([cf2d781](https://github.com/inference-gateway/a2a-cli/commit/cf2d781bb448c845e7df2c0bc4b9c2967b08bce7))

## [0.3.0](https://github.com/inference-gateway/a2a-cli/compare/v0.2.5...v0.3.0) (2025-07-23)

### ✨ Features

* Implement `a2a generate --devcontainer` command ([#9](https://github.com/inference-gateway/a2a-cli/issues/9)) ([6b4556a](https://github.com/inference-gateway/a2a-cli/commit/6b4556ade1af97a5f68a7de60d43eef70cbf4838)), closes [#7](https://github.com/inference-gateway/a2a-cli/issues/7)

### 📚 Documentation

* Remove unrelated sections and update links ([8d7bb16](https://github.com/inference-gateway/a2a-cli/commit/8d7bb16066f5fd47b84f5667112515bc9358c054))

### 🔧 Miscellaneous

* Remove old examples ([83d871c](https://github.com/inference-gateway/a2a-cli/commit/83d871ce167c33987805f66bb25ac038722e6f91))

## [0.2.5](https://github.com/inference-gateway/a2a-cli/compare/v0.2.4...v0.2.5) (2025-07-23)

### 🐛 Bug Fixes

* **templates:** Remove language field from agent.json card ([#6](https://github.com/inference-gateway/a2a-cli/issues/6)) ([8dfb519](https://github.com/inference-gateway/a2a-cli/commit/8dfb51908674b0f23acaa54789a2fa3d1dc8c358)), closes [#5](https://github.com/inference-gateway/a2a-cli/issues/5)

## [0.2.4](https://github.com/inference-gateway/a2a-cli/compare/v0.2.3...v0.2.4) (2025-07-23)

### ♻️ Improvements

* Refactor templates to match inference-gateway structure ([#4](https://github.com/inference-gateway/a2a-cli/issues/4)) ([a836935](https://github.com/inference-gateway/a2a-cli/commit/a836935e6ca9529a4388f3c90e410c46e056d98f))

### 🐛 Bug Fixes

* Update dependencies in go.mod and go.sum ([a44c568](https://github.com/inference-gateway/a2a-cli/commit/a44c568efdc2fcfb90a2dcb5117f8e31ef0f2f19))

### 👷 CI

* Update Claude PR Assistant workflow ([#3](https://github.com/inference-gateway/a2a-cli/issues/3)) ([3df9330](https://github.com/inference-gateway/a2a-cli/commit/3df93302f46e382fe3115e0287965e5aaa14f9bf))

### 📚 Documentation

* Add CLAUDE.md for project guidance and development instructions ([0aacc20](https://github.com/inference-gateway/a2a-cli/commit/0aacc20657fa22ea08d0301231716e78667135fa))

## [0.2.3](https://github.com/inference-gateway/a2a-cli/compare/v0.2.2...v0.2.3) (2025-07-21)

### 🐛 Bug Fixes

* Update project name in goreleaser configuration ([15873c1](https://github.com/inference-gateway/a2a-cli/commit/15873c1ec8d0f5a5e80dec779707f1695056fc3e))

## [0.2.2](https://github.com/inference-gateway/a2a-cli/compare/v0.2.1...v0.2.2) (2025-07-21)

### 📚 Documentation

* Add early development warning to README ([26f9bcd](https://github.com/inference-gateway/a2a-cli/commit/26f9bcdc0d3c13828f70ec867b1ed53158dd13f2))
* Add installation instructions and script to README ([7cc1125](https://github.com/inference-gateway/a2a-cli/commit/7cc1125b7297a969e223892ccaa33dfe77cdd3b1))

### 🔨 Miscellaneous

* Update artifact upload script and add installation script for A2A CLI ([a0e810b](https://github.com/inference-gateway/a2a-cli/commit/a0e810bcb76212a03a7ba7568026452b4c0f55e3))

## [0.2.1](https://github.com/inference-gateway/a2a-cli/compare/v0.2.0...v0.2.1) (2025-07-21)

### 🐛 Bug Fixes

* Correct artifact filename pattern in upload step ([0941b93](https://github.com/inference-gateway/a2a-cli/commit/0941b93c0df767b6cd96e3b72e752202b0e14343))

### 🔧 Miscellaneous

* Remove Docker Buildx setup and GitHub Container Registry login from artifacts workflow ([0ddea6e](https://github.com/inference-gateway/a2a-cli/commit/0ddea6e59fa607f03229a2a3017f38fb9240797e))

## [0.2.0](https://github.com/inference-gateway/a2a-cli/compare/v0.1.0...v0.2.0) (2025-07-21)

### ✨ Features

* Add support for generated file headers and .gitattributes template ([90bbc46](https://github.com/inference-gateway/a2a-cli/commit/90bbc46890ebe22b27c5a294111c3b2cf48a219f))

### ♻️ Improvements

* Introduce .adl-ignore file support to protect implementations during generation ([377bfb0](https://github.com/inference-gateway/a2a-cli/commit/377bfb085b487e73c75b887188ba0205dfba1122))
* Remove commented-out code and clean up validation functions ([9ba910f](https://github.com/inference-gateway/a2a-cli/commit/9ba910fc1e71c4b2dd67fbb511f75c064749d210))

### 👷 CI

* Add GoReleaser configuration and GitHub Actions workflow for artifact uploads ([cd7ae6d](https://github.com/inference-gateway/a2a-cli/commit/cd7ae6d1bb781d4a308d1b79caa12417ec108e97))

### 📚 Documentation

* Add contributing guide and roadmap for language support ([c24afd5](https://github.com/inference-gateway/a2a-cli/commit/c24afd5ec8cadce2ab43bedd2778bfee9c74053c))
* Update README.md to enhance formatting and visibility of project badges ([9b7df5c](https://github.com/inference-gateway/a2a-cli/commit/9b7df5cf6d77664b681b5ea80ba64fb6cdff0bed))

### 🔧 Miscellaneous

* Add .releaserc.yaml for semantic release configuration ([36e2341](https://github.com/inference-gateway/a2a-cli/commit/36e2341abfb3e010b1d085b1ef1f2e8ca50d66b5))
* Add initial .editorconfig file for consistent coding styles ([4d35977](https://github.com/inference-gateway/a2a-cli/commit/4d359772a6029d89dbc65fe36c8d26f169778c41))
* Create LICENSE ([2176bf0](https://github.com/inference-gateway/a2a-cli/commit/2176bf07575e198aeb332f0ff1317df1c2de105e))
* Simplify ldflags in GoReleaser configuration ([39cd5bc](https://github.com/inference-gateway/a2a-cli/commit/39cd5bcb566c040dc11bfb7f119db290d538d702))
