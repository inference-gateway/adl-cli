package registry

import (
	"context"
	"fmt"
	"os"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// ResolvedSkill is the result of resolving a spec.skills[] entry: either
// fetched from a registry, read from cache, or scaffolded from a bare
// declaration. Body holds the markdown that should be written to
// skills/<id>.md in the generated project (empty for bare skills, which
// are scaffolded by the template engine instead).
type ResolvedSkill struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Version     string
	Bare        bool
	Body        []byte
}

// Resolver coordinates registry fetches, cache lookups, and bare-skill
// scaffolding to produce ResolvedSkill entries the generator can consume.
type Resolver struct {
	Client  *Client
	Cache   *Cache
	Offline bool
}

// NewDefaultResolver builds a Resolver using ADL_SKILLS_REGISTRY (or the
// hardcoded default) and the user's home cache directory.
func NewDefaultResolver() (*Resolver, error) {
	base := os.Getenv(EnvBaseURL)
	cache, err := NewCache("")
	if err != nil {
		return nil, err
	}
	return &Resolver{
		Client: NewClient(base),
		Cache:  cache,
	}, nil
}

// Resolve fetches or scaffolds metadata for a single skill entry.
func (r *Resolver) Resolve(ctx context.Context, skill schema.Skill) (*ResolvedSkill, error) {
	if skill.ID == "" {
		return nil, fmt.Errorf("skill id is required")
	}

	if skill.Bare {
		if skill.Name == "" || skill.Description == "" {
			return nil, fmt.Errorf("bare skill %q requires both name and description", skill.ID)
		}
		return &ResolvedSkill{
			ID:          skill.ID,
			Name:        skill.Name,
			Description: skill.Description,
			Tags:        skill.Tags,
			Version:     skill.Version,
			Bare:        true,
		}, nil
	}

	cached, ok, err := r.Cache.Get(skill.ID, skill.Version)
	if err != nil {
		return nil, err
	}
	if !ok {
		if r.Offline {
			return nil, fmt.Errorf("skill %q is not cached and --offline is set", skill.ID)
		}
		var body []byte
		var fetchErr error
		if skill.Source != "" {
			body, fetchErr = r.Client.FetchURL(ctx, skill.Source)
		} else {
			body, fetchErr = r.Client.FetchByID(ctx, skill.ID, skill.Version)
		}
		if fetchErr != nil {
			return nil, fetchErr
		}
		if err := r.Cache.Put(skill.ID, skill.Version, body); err != nil {
			return nil, err
		}
		cached = body
	}

	doc, err := ParseSkillDocument(cached)
	if err != nil {
		return nil, fmt.Errorf("skill %q: %w", skill.ID, err)
	}

	name := doc.Frontmatter.Name
	if skill.Name != "" {
		name = skill.Name
	}
	description := doc.Frontmatter.Description
	if skill.Description != "" {
		description = skill.Description
	}
	tags := doc.Frontmatter.Tags
	if len(skill.Tags) > 0 {
		tags = skill.Tags
	}
	version := doc.Frontmatter.Version
	if skill.Version != "" {
		version = skill.Version
	}

	return &ResolvedSkill{
		ID:          skill.ID,
		Name:        name,
		Description: description,
		Tags:        tags,
		Version:     version,
		Body:        cached,
	}, nil
}

// ResolveAll resolves every skill in skills, returning the resolved
// entries in the same order. Resolution fails fast on the first error.
func (r *Resolver) ResolveAll(ctx context.Context, skills []schema.Skill) ([]*ResolvedSkill, error) {
	resolved := make([]*ResolvedSkill, 0, len(skills))
	for _, skill := range skills {
		rs, err := r.Resolve(ctx, skill)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, rs)
	}
	return resolved, nil
}
