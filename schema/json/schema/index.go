package schema

type IndexData struct {
	Title       string                `json:"name"`
	Description string                `json:"description"`
	Children    map[string]*IndexData `json:"children"`
}

func NewIndexData(s *Schema) *IndexData {
	return newIndexData(s, map[*Schema]bool{})
}

// newIndexData indexes s and its properties. path holds the schemas
// currently being indexed; a schema already on it is a cycle and becomes a leaf.
func newIndexData(s *Schema, path map[*Schema]bool) *IndexData {
	if s == nil {
		return nil
	}

	d := &IndexData{
		Title:       s.Title,
		Description: s.Description,
		Children:    make(map[string]*IndexData),
	}

	if path[s] || s.Properties == nil {
		return d
	}
	path[s] = true
	defer delete(path, s)

	for it := s.Properties.Iter(); it.Next(); {
		d.Children[it.Key()] = newIndexData(it.Value(), path)
	}

	return d
}
