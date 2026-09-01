package openapi

type Tag struct {
	Name         string        `yaml:"name" json:"name"`
	Summary      string        `yaml:"summary" json:"summary"`
	Description  string        `yaml:"description" json:"description"`
	ExternalDocs *ExternalDocs `yaml:"externalDocs" json:"externalDocs"`
	Parent       string        `yaml:"parent" json:"parent"`
	Kind         string        `yaml:"kind" json:"kind"`
}

func (t *Tag) patch(patch *Tag) {
	if patch == nil {
		return
	}
	if len(patch.Summary) > 0 {
		t.Summary = patch.Summary
	}
	if len(patch.Description) > 0 {
		t.Description = patch.Description
	}
	if patch.ExternalDocs != nil {
		if t.ExternalDocs == nil {
			t.ExternalDocs = patch.ExternalDocs
		}
	}
	if patch.Parent != "" {
		t.Parent = patch.Parent
	}
	if len(patch.Kind) > 0 {
		t.Kind = patch.Kind
	}
}

// tagPath returns the chain from root ancestor down to the tag itself,
// e.g. ["billing", "billing.invoices"] for a tag with Parent: "billing".
func tagPath(tagsByName map[string]*Tag, name string) []*Tag {
	var path []*Tag
	seen := map[string]bool{}

	for name != "" {
		if seen[name] {
			// cycle in parent chain — bail out rather than loop forever
			break
		}
		seen[name] = true
		tag, ok := tagsByName[name]
		if !ok {
			break // parent references an undefined tag
		}
		path = append([]*Tag{tag}, path...) // prepend
		name = tag.Parent
	}
	return path
}

func tagsByName(tags []*Tag) map[string]*Tag {
	result := make(map[string]*Tag)
	for _, tag := range tags {
		result[tag.Name] = tag
	}
	return result
}
