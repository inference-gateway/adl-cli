# Changelog

All notable changes to this project will be documented in this file.

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
* Rename .adl-ignore to .a2a-ignore and update related functionality ([86ea2b5](https://github.com/inference-gateway/a2a-cli/commit/86ea2b58c58f7f3cc3173cc5aa8d4a47bc73efcb))

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
