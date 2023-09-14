package parameter

import (
	"fmt"
	"mokapi/config/dynamic/common"
)

func (r *Ref) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return fmt.Errorf("unable to resolve parameter reference %v: %v", r.Ref, err)
		}
	}

	if r.Value == nil {
		return nil
	}

	if err := r.Value.Schema.Parse(config, reader); err != nil {
		return err
	}

	return nil
}

func (r *NamedParameters) Parse(config *common.Config, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref, &r.Value, config, reader); err != nil {
			return err
		}
	}

	return nil
}
