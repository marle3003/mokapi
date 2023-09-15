package schema

import (
	"fmt"
	"mokapi/config/dynamic/common"
)

func (s *Schemas) Parse(config *common.Config, reader common.Reader) error {
	if s == nil {
		return nil
	}

	for it := s.Iter(); it.Next(); {
		if err := it.Value().Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (r *Ref) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}
	if len(r.Ref) > 0 {
		if err := common.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return fmt.Errorf("parse schema failed: %w", err)
		}
	}

	if r.Value == nil {
		return nil
	}

	return r.Value.Parse(config, reader)
}

func (s *Schema) Parse(config *common.Config, reader common.Reader) error {
	if s == nil {
		return nil
	}

	if err := s.Items.Parse(config, reader); err != nil {
		return err
	}

	if err := s.Properties.Parse(config, reader); err != nil {
		return err
	}

	if err := s.AdditionalProperties.Parse(config, reader); err != nil {
		return err
	}

	for _, r := range s.AnyOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, r := range s.AllOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	for _, r := range s.OneOf {
		if err := r.Parse(config, reader); err != nil {
			return err
		}
	}

	return nil
}

func (ap *AdditionalProperties) Parse(config *common.Config, reader common.Reader) error {
	if ap == nil {
		return nil
	}

	return ap.Ref.Parse(config, reader)
}

func (r *Ref) String() string {
	if r.Value == nil && len(r.Ref) == 0 {
		return fmt.Sprintf("no schema defined")
	}
	if r.Value == nil {
		return fmt.Sprintf("unresolved schema %v", r.Ref)
	}
	return r.Value.String()
}
