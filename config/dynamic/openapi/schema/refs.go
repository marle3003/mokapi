package schema

import (
	"fmt"
	"mokapi/config/dynamic/common"
)

func (s *Schemas) Parse(file *common.File, reader common.Reader) error {
	if s == nil {
		return nil
	}

	for it := s.Iter(); it.Next(); {
		if err := it.Value().(*Ref).Parse(file, reader); err != nil {
			return err
		}
	}

	return nil
}

func (s *SchemasRef) Parse(file *common.File, reader common.Reader) error {
	if s == nil {
		return nil
	}
	if len(s.Ref()) > 0 && s.Value == nil {
		if err := common.Resolve(s.Ref(), &s.Value, file, reader); err != nil {
			return fmt.Errorf("error on parsing file %v: %v", file.Url, err)
		}
	}

	if s.Value == nil {
		return nil
	}

	return s.Value.Parse(file, reader)
}

func (r *Ref) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}
	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	if r.Value == nil {
		return nil
	}

	return r.Value.Parse(file, reader)
}

func (s *Schema) Parse(file *common.File, reader common.Reader) error {
	if s == nil {
		return nil
	}

	if err := s.Items.Parse(file, reader); err != nil {
		return err
	}

	if err := s.Properties.Parse(file, reader); err != nil {
		return err
	}

	if err := s.AdditionalProperties.Parse(file, reader); err != nil {
		return err
	}

	return nil
}

func (ap *AdditionalProperties) Parse(file *common.File, reader common.Reader) error {
	if ap == nil {
		return nil
	}

	return ap.Ref.Parse(file, reader)
}
