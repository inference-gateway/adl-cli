package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/inference-gateway/adl-cli/internal/publisher"
	"github.com/inference-gateway/adl-cli/internal/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	publishURL    string
	publishRef    string
	publishDryRun bool
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish [adl-file]",
	Short: "Open a PR to add your agent to the public catalog",
	Long: `Publish your agent to the public catalog at registry.inference-gateway.com.

The catalog is sourced from the agents.yaml list in inference-gateway/agents,
where each entry is just a pointer to a repository ({ url, ref }). This command
validates your ADL manifest, resolves your agent's GitHub repository, and opens
a pull request against inference-gateway/agents that appends your repository as
a new catalog entry.

The agent's name, description, and version shown in the live catalog come from
your repository's agent.yaml (metadata.*) when the catalog is built - the PR
only contributes the { url, ref } pointer. The PR title and body are seeded from
your local agent.yaml so you can refine the summary before submitting.

The agent repository URL defaults to the 'origin' git remote in the current
directory and can be overridden with --url; the ref defaults to 'main' and can
be overridden with --ref.

Publishing requires the GitHub CLI ('gh') to be installed and authenticated
('gh auth login'). It forks inference-gateway/agents to your account, commits
the new entry on a branch, and opens the pull request from your fork. Use
--dry-run to preview the catalog entry, PR title, and PR body without making
any changes.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPublish,
}

func init() {
	rootCmd.AddCommand(publishCmd)

	publishCmd.Flags().StringVar(&publishURL, "url", "", "Agent repository URL (defaults to the 'origin' git remote)")
	publishCmd.Flags().StringVar(&publishRef, "ref", publisher.DefaultRef, "Git ref (branch, tag, or SHA) to record in the catalog entry")
	publishCmd.Flags().BoolVar(&publishDryRun, "dry-run", false, "Print the catalog entry, PR title, and PR body without opening a PR")
}

func runPublish(cmd *cobra.Command, args []string) error {
	adlFile := "agent.yaml"
	if len(args) > 0 {
		adlFile = args[0]
	}

	if _, err := os.Stat(adlFile); os.IsNotExist(err) {
		return fmt.Errorf("ADL file '%s' does not exist", adlFile)
	}

	validator := schema.NewValidator()
	warnings, err := validator.ValidateFile(adlFile)
	if err != nil {
		return fmt.Errorf("ADL validation failed: %w", err)
	}
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "⚠️  %s\n", w)
	}

	data, err := os.ReadFile(adlFile)
	if err != nil {
		return fmt.Errorf("failed to read ADL file: %w", err)
	}
	var adl schema.ADL
	if err := yaml.Unmarshal(data, &adl); err != nil {
		return fmt.Errorf("failed to parse ADL file: %w", err)
	}

	ctx := context.Background()
	pub := publisher.New(os.Stdout)

	repoURL, err := resolveRepoURL(ctx, pub.Runner)
	if err != nil {
		return err
	}

	meta := publisher.Metadata{
		Name:        adl.Metadata.Name,
		Description: adl.Metadata.Description,
		Version:     adl.Metadata.Version,
	}

	fmt.Printf("Publishing '%s' (%s) to the agents catalog...\n", meta.Name, repoURL)

	prURL, err := pub.Publish(ctx, meta, publisher.Options{
		RepoURL: repoURL,
		Ref:     publishRef,
		DryRun:  publishDryRun,
	})
	if err != nil {
		return fmt.Errorf("publish failed: %w", err)
	}

	if publishDryRun {
		return nil
	}

	fmt.Println("✅ Pull request opened!")
	fmt.Printf("🔗 %s\n", prURL)
	return nil
}

func resolveRepoURL(ctx context.Context, runner publisher.Runner) (string, error) {
	if publishURL != "" {
		normalized, err := publisher.NormalizeRepoURL(publishURL)
		if err != nil {
			return "", fmt.Errorf("invalid --url: %w", err)
		}
		return normalized, nil
	}

	resolved, err := publisher.ResolveRepoURLFromGit(ctx, runner, ".")
	if err != nil {
		return "", fmt.Errorf("could not determine the agent repository URL: %w\nPass --url https://github.com/<owner>/<repo> to set it explicitly", err)
	}
	return resolved, nil
}
