// ABOUTME: Skill registry for storing and retrieving loaded skills
// ABOUTME: Thread-safe storage with lookup by name, tags, and patterns

package skills

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Registry stores and manages loaded skills
type Registry struct {
	skills map[string]*Skill // Skills indexed by name
	mu     sync.RWMutex
}

// NewRegistry creates a new skill registry
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]*Skill),
	}
}

// Register adds a skill to the registry
func (r *Registry) Register(skill *Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if skill.Name == "" {
		return fmt.Errorf("skill has no name")
	}

	// Allow overriding (project skills override user skills, etc.)
	r.skills[skill.Name] = skill
	return nil
}

// Get retrieves a skill by name
func (r *Registry) Get(name string) (*Skill, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, exists := r.skills[name]
	if !exists {
		return nil, fmt.Errorf("skill not found: %s", name)
	}

	return skill, nil
}

// List returns all registered skill names sorted alphabetically
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.skills))
	for name := range r.skills {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// All returns all registered skills sorted by priority (high to low)
func (r *Registry) All() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}

	// Sort by priority (higher first), then name
	sort.Slice(skills, func(i, j int) bool {
		if skills[i].Priority != skills[j].Priority {
			return skills[i].Priority > skills[j].Priority
		}
		return skills[i].Name < skills[j].Name
	})

	return skills
}

// FindByTags returns skills that have any of the specified tags
func (r *Registry) FindByTags(tags ...string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*Skill
	for _, skill := range r.skills {
		if hasAnyTag(skill.Tags, tags) {
			matches = append(matches, skill)
		}
	}

	// Sort by priority
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority > matches[j].Priority
		}
		return matches[i].Name < matches[j].Name
	})

	return matches
}

// FindByPattern returns skills that match the given message pattern
func (r *Registry) FindByPattern(message string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*Skill
	for _, skill := range r.skills {
		if skill.MatchesPattern(message) {
			matches = append(matches, skill)
		}
	}

	// Sort by priority
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority > matches[j].Priority
		}
		return matches[i].Name < matches[j].Name
	})

	return matches
}

// Search finds skills by name, description, or tags (case-insensitive)
func (r *Registry) Search(query string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lowerQuery := strings.ToLower(query)
	var matches []*Skill

	for _, skill := range r.skills {
		matched := false

		// Check name
		if strings.Contains(strings.ToLower(skill.Name), lowerQuery) {
			matched = true
		}

		// Check description
		if !matched && strings.Contains(strings.ToLower(skill.Description), lowerQuery) {
			matched = true
		}

		// Check tags
		if !matched {
			for _, tag := range skill.Tags {
				if strings.Contains(strings.ToLower(tag), lowerQuery) {
					matched = true
					break
				}
			}
		}

		if matched {
			matches = append(matches, skill)
		}
	}

	// Sort by priority
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Priority != matches[j].Priority {
			return matches[i].Priority > matches[j].Priority
		}
		return matches[i].Name < matches[j].Name
	})

	return matches
}

// Count returns the number of registered skills
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}

// Clear removes all skills from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills = make(map[string]*Skill)
}

// hasAnyTag checks if slice a contains any element from slice b
func hasAnyTag(skillTags, searchTags []string) bool {
	for _, st := range skillTags {
		for _, search := range searchTags {
			if strings.EqualFold(st, search) {
				return true
			}
		}
	}
	return false
}
