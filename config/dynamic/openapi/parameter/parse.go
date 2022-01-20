package parameter

import "mokapi/config/dynamic/common"

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

	if err := r.Value.Schema.Parse(file, reader); err != nil {
		return err
	}

	return nil
}

func (r *NamedParameters) Parse(file *common.File, reader common.Reader) error {
	if r == nil {
		return nil
	}

	if len(r.Ref()) > 0 && r.Value == nil {
		if err := common.Resolve(r.Ref(), &r.Value, file, reader); err != nil {
			return err
		}
	}

	return nil
}
