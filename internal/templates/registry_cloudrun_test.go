package templates

import (
	"strings"
	"testing"

	schema "github.com/inference-gateway/adl-cli/internal/schema"
)

// cloudRunADL returns a minimal Go ADL wired for Cloud Run deployment.
// Individual tests tweak the returned cloudrun block to exercise edge cases.
func cloudRunADL() *schema.ADL {
	return &schema.ADL{
		APIVersion: "adl.inference-gateway.com/v1",
		Kind:       "Agent",
		Metadata: schema.Metadata{
			Name:        "cloudrun-agent",
			Description: "test",
			Version:     "1.0.0",
		},
		Spec: schema.Spec{
			Capabilities: schema.Capabilities{Streaming: true},
			Server:       schema.Server{Port: 8080},
			Language: schema.Language{
				Go: &schema.GoConfig{
					Module:  "github.com/example/cloudrun-agent",
					Version: "1.26.7",
				},
			},
			Deployment: &schema.DeploymentConfig{
				Type: schema.DeploymentConfigTypeCloudRun,
				CloudRun: &schema.CloudRunConfig{
					Image: &schema.ImageConfig{
						Registry:      "ghcr.io",
						Repository:    "example/cloudrun-agent",
						Tag:           "latest",
						UseCloudBuild: false,
					},
					Resources: &schema.ResourcesConfig{CPU: "2", Memory: "1Gi"},
					Scaling:   &schema.ScalingConfig{MinInstances: 1, MaxInstances: 50, Concurrency: 500},
					Service: &schema.ServiceConfig{
						Timeout:              1800,
						AllowUnauthenticated: false,
						ServiceAccount:       "sa@PROJECT_ID.iam.gserviceaccount.com",
						ExecutionEnvironment: "gen2",
					},
					Environment: schema.CloudRunConfigEnvironment{
						"LOG_LEVEL":     "debug",
						"CACHE_ENABLED": "true",
					},
				},
			},
		},
	}
}

func renderTaskfile(t *testing.T, adl *schema.ADL) string {
	t.Helper()
	r, err := NewRegistry("go")
	if err != nil {
		t.Fatalf("NewRegistry: %v", err)
	}
	out, err := NewWithRegistry("minimal", r).ExecuteTemplate("taskfile/taskfile.yml", Context{ADL: adl, Language: "go"})
	if err != nil {
		t.Fatalf("ExecuteTemplate(taskfile/taskfile.yml): %v", err)
	}
	return out
}

// TestTaskfile_CloudRun_HonorsManifest covers issue #434: image, auth and
// custom environment variables from the manifest must reach gcloud.
func TestTaskfile_CloudRun_HonorsManifest(t *testing.T) {
	out := renderTaskfile(t, cloudRunADL())

	for _, want := range []string{
		"--image ghcr.io/example/cloudrun-agent:latest",
		"--no-allow-unauthenticated",
		"LOG_LEVEL=debug",
		"CACHE_ENABLED=true",
		`--cpu "2"`,
		"--timeout 1800",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Taskfile missing %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, "gcr.io/${PROJECT_ID}") {
		t.Errorf("Taskfile still uses the hardcoded gcr.io image\n---\n%s", out)
	}
	if strings.Contains(out, " --allow-unauthenticated") {
		t.Errorf("allowUnauthenticated: false must not emit --allow-unauthenticated\n---\n%s", out)
	}
	if !strings.Contains(out, "docker push ghcr.io/example/cloudrun-agent:latest") {
		t.Errorf("useCloudBuild: false should build and push with docker\n---\n%s", out)
	}
}

// TestTaskfile_CloudRun_UseCloudBuild verifies that useCloudBuild swaps the
// local docker build for a Cloud Build submission in the deploy task.
func TestTaskfile_CloudRun_UseCloudBuild(t *testing.T) {
	adl := cloudRunADL()
	adl.Spec.Deployment.CloudRun.Image.UseCloudBuild = true
	adl.Spec.Deployment.CloudRun.Service.AllowUnauthenticated = true

	out := renderTaskfile(t, adl)

	if strings.Contains(out, "docker push") {
		t.Errorf("useCloudBuild: true should not push with docker\n---\n%s", out)
	}
	if !strings.Contains(out, "gcloud builds submit --tag ghcr.io/example/cloudrun-agent:latest") {
		t.Errorf("expected a Cloud Build submission\n---\n%s", out)
	}
	if !strings.Contains(out, "--allow-unauthenticated") {
		t.Errorf("allowUnauthenticated: true should emit --allow-unauthenticated\n---\n%s", out)
	}
}

// TestTaskfile_CloudRun_Defaults keeps the no-cloudrun-block behaviour: the
// gcr.io image, public access and the documented resource defaults.
func TestTaskfile_CloudRun_Defaults(t *testing.T) {
	adl := cloudRunADL()
	adl.Spec.Deployment.CloudRun = nil

	out := renderTaskfile(t, adl)

	for _, want := range []string{
		"--image gcr.io/${PROJECT_ID}/cloudrun-agent:1.0.0",
		"--allow-unauthenticated",
		`--cpu "1"`,
		`--memory "512Mi"`,
		"--max-instances 10",
		"--timeout 3600",
		"--service-account \"cloudrun-agent@${PROJECT_ID}.iam.gserviceaccount.com\"",
		"--execution-environment gen2",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Taskfile missing default %q\n---\n%s", want, out)
		}
	}
}
