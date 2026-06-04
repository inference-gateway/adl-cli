{
  description = "ADL CLI - Generate enterprise-ready A2A agents from YAML";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        inherit (pkgs) lib;

        version = "0.43.2";

        adl = pkgs.buildGoModule (finalAttrs: {
          __structuredAttrs = true;

          pname = "adl";
          inherit version;

          src = lib.cleanSourceWith {
            src = ./.;
            filter =
              path: type:
              let
                baseName = baseNameOf (toString path);
                relPath = lib.removePrefix (toString ./. + "/") (toString path);
              in
              !(
                baseName == ".git"
                || baseName == "dist"
                || baseName == "result"
                || baseName == "bin"
                || baseName == ".flox"
                || baseName == ".infer"
                || baseName == ".task"
                || baseName == "node_modules"
                || baseName == "test-output"
                || (type == "regular" && relPath == "adl")
              );
          };

          vendorHash = "sha256-zDITDCNONsbqc04CoU8uqUo/11/hxN94gW8Bc6I0gcQ=";

          goSum = ./go.sum;

          proxyVendor = true;

          env.CGO_ENABLED = "0";

          ldflags = [
            "-s"
            "-w"
            "-X=main.Version=${version}"
          ];

          preCheck = ''
            export HOME=$TMPDIR
          '';

          nativeBuildInputs = [
            pkgs.installShellFiles
          ];

          postInstall = ''
            mv $out/bin/adl-cli $out/bin/adl
            installShellCompletion --cmd adl \
              --bash <($out/bin/adl completion bash) \
              --fish <($out/bin/adl completion fish) \
              --zsh <($out/bin/adl completion zsh)
          '';

          meta = {
            description = "Generate enterprise-ready A2A agents from YAML Agent Definition Language files";
            longDescription = ''
              The ADL CLI is a command-line tool that generates enterprise-ready A2A
              (Agent-to-Agent) agent projects from YAML-based Agent Definition Language
              (ADL) files. It produces complete project scaffolding with business logic
              placeholders, CI/CD pipelines, and sandbox environments (Flox,
              DevContainer) for Go and Rust agents.
            '';
            homepage = "https://github.com/inference-gateway/adl-cli";
            changelog = "https://github.com/inference-gateway/adl-cli/blob/v${version}/CHANGELOG.md";
            license = lib.licenses.mit;
            maintainers = [
              {
                name = "Eden Reich";
                email = "eden.reich@gmail.com";
                github = "edenreich";
                githubId = 26537388;
              }
            ];
            mainProgram = "adl";
            platforms = lib.platforms.unix;
          };
        });
      in
      {
        packages = {
          default = adl;
          inherit adl;
        };

        apps.default = {
          type = "app";
          program = "${adl}/bin/adl";
          meta = {
            description = "Run the adl CLI";
            mainProgram = "adl";
          };
        };

        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.go
            pkgs.go-task
            pkgs.golangci-lint
            pkgs.gopls
            pkgs.goreleaser
          ];
        };
      }
    );
}
