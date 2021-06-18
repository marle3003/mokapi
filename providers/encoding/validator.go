package encoding

import (
	"fmt"
	"mokapi/config/dynamic/openapi"
	"regexp"
	"time"
)

func validateString(s string, schema *openapi.SchemaRef) error {
	if len(schema.Value.Pattern) > 0 {
		r, err := regexp.Compile(`{{\s*([\w\.]+)\s*}}`)
		if err != nil {
			return err
		}
		if !r.MatchString(s) {
			return fmt.Errorf("%q does not match pattern %q", s, schema.Value.Pattern)
		}
	}

	switch schema.Value.Format {
	case "date":
		_, err := time.Parse("2006-01-02", s)
		return err
	case "date-time":
		_, err := time.Parse(time.RFC3339, s)
		return err
	}
	return nil
}

func validateFloat64(n float64, schema *openapi.SchemaRef) error {
	if schema.Value.Minimum != nil {
		min := *schema.Value.Minimum
		if schema.Value.ExclusiveMinimum != nil && (*schema.Value.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower as the expected mininmum %v", n, min)
		} else if n < min {
			return fmt.Errorf("%v is lower as the expected mininmum %v", n, min)
		}
	}
	if schema.Value.Maximum != nil {
		max := *schema.Value.Maximum
		if schema.Value.ExclusiveMaximum != nil && (*schema.Value.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is higher as the expected maximum %v", n, max)
		} else if n > max {
			return fmt.Errorf("%v is higher as the expected maximum %v", n, max)
		}
	}
	return nil
}

func validateInt64(n int64, schema *openapi.SchemaRef) error {
	if schema.Value.Minimum != nil {
		min := int64(*schema.Value.Minimum)
		if schema.Value.ExclusiveMinimum != nil && (*schema.Value.ExclusiveMinimum) && n <= min {
			return fmt.Errorf("%v is lower as the expected mininmum %v", n, min)
		} else if n < min {
			return fmt.Errorf("%v is lower as the expected mininmum %v", n, min)
		}
	}
	if schema.Value.Maximum != nil {
		max := int64(*schema.Value.Maximum)
		if schema.Value.ExclusiveMaximum != nil && (*schema.Value.ExclusiveMaximum) && n >= max {
			return fmt.Errorf("%v is higher as the expected maximum %v", n, max)
		} else if n > max {
			return fmt.Errorf("%v is higher as the expected maximum %v", n, max)
		}
	}
	return nil
}
