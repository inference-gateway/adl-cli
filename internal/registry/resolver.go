package registry

import (
	"context"
	"fmt"
	"os"

	"github.com/inference-gateway/adl-cli/internal/schema"
)

// SkillFile is the canonical filename of a skill's playbook within the
// skill directory.
const SkillFile = "SKILL.md"

// ResolvedSkill is the result of resolving a spec.skills[] entry. For
// non-bare skills, Files contains every blob fetched from the source
// (the registry default, or a GitHub tree URL via Installer), keyed by
// path relative to the skill's directory. SKILL.md is always present
// for non-bare skills. For bare skills, Files is empty - the template
// engine scaffolds SKILL.md from the ADL metadata.
type ResolvedSkill struct {
	ID          string
	Name        string
	Description string
	Tags        []string
	Version     string
	License     string
	Bare        bool
	Files       map[string][]byte
}

// Resolver coordinates the GitHub Installer, the default registry
// Client, the on-disk Cache, and bare-skill scaffolding to produce
// ResolvedSkill entries the generator can consume.
type Resolver struct {
	Client    *Client
	Installer *Installer
	Cache     *Cache
	Offline   bool
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
		Client:    NewClient(base),
		Installer: NewInstaller(),
		Cache:     cache,
	}, nil
}

// Resolve fetches or scaffolds metadata + files for a single skill entry.
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
			License:     string(skill.License),
			Bare:        true,
		}, nil
	}

	cacheRef, files, err := r.loadFiles(ctx, skill)
	if err != nil {
		return nil, err
	}

	skillMD, ok := files[SkillFile]
	if !ok {
		return nil, fmt.Errorf("skill %q: no %s found in fetched files", skill.ID, SkillFile)
	}

	doc, err := ParseSkillDocument(skillMD)
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
	if version == "" {
		version = cacheRef
	}
	license := doc.Frontmatter.License
	if skill.License != "" {
		license = string(skill.License)
	}

	return &ResolvedSkill{
		ID:          skill.ID,
		Name:        name,
		Description: description,
		Tags:        tags,
		Version:     version,
		License:     license,
		Files:       files,
	}, nil
}

// loadFiles returns the cached or freshly fetched files for a non-bare
// skill, along with the cache ref used (either skill.Version or, for
// source-pulled skills, the GitHub ref from the URL).
func (r *Resolver) loadFiles(ctx context.Context, skill schema.Skill) (string, map[string][]byte, error) {
	if skill.Source != "" {
		expanded := ExpandShorthand(skill.Source)
		loc, err := ParseGitHubTreeURL(expanded)
		if err != nil {
			return "", nil, fmt.Errorf("skill %q: %w", skill.ID, err)
		}
		ref := loc.Ref
		if cached, ok, err := r.Cache.Get(skill.ID, ref); err != nil {
			return "", nil, err
		} else if ok {
			return ref, cached, nil
		}
		if r.Offline {
			return "", nil, fmt.Errorf("skill %q is not cached and --offline is set", skill.ID)
		}
		files, err := r.Installer.Fetch(ctx, loc)
		if err != nil {
			return "", nil, fmt.Errorf("skill %q: %w", skill.ID, err)
		}
		if err := r.Cache.Put(skill.ID, ref, files); err != nil {
			return "", nil, err
		}
		return ref, files, nil
	}

	ref := skill.Version
	if cached, ok, err := r.Cache.Get(skill.ID, ref); err != nil {
		return "", nil, err
	} else if ok {
		return ref, cached, nil
	}
	if r.Offline {
		return "", nil, fmt.Errorf("skill %q is not cached and --offline is set", skill.ID)
	}
	body, err := r.Client.FetchByID(ctx, skill.ID, skill.Version)
	if err != nil {
		return "", nil, err
	}
	files := map[string][]byte{SkillFile: body}
	if err := r.Cache.Put(skill.ID, ref, files); err != nil {
		return "", nil, err
	}
	return ref, files, nil
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
