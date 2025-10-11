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
