package schema

type IndexData struct {
	Title       string                `json:"name"`
	Description string                `json:"description"`
	Children    map[string]*IndexData `json:"children"`
}

func NewIndexData(s *Schema) *IndexData {
	if s == nil {
		return nil
	}

	d := &IndexData{
		Title:       s.Title,
		Description: s.Description,
		Children:    make(map[string]*IndexData),
	}

	if s.Properties != nil {
		for it := s.Properties.Iter(); it.Next(); {
			d.Children[it.Key()] = NewIndexData(it.Value())
		}
	}

	return d
}
