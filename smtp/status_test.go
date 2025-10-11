package smtp_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"mokapi/smtp"
	"testing"
)

func TestEnhancedStatusCode_UnmarshalJSON(t *testing.T) {
	var ehc smtp.EnhancedStatusCode
	err := json.Unmarshal([]byte(`"1.2.3"`), &ehc)
	require.NoError(t, err)
	require.Equal(t, smtp.EnhancedStatusCode{1, 2, 3}, ehc)

	err = json.Unmarshal([]byte(`"1.2"`), &ehc)
	require.EqualError(t, err, "unexpected value 1.2, expected format x.x.x")

	err = json.Unmarshal([]byte(`"1.2.b"`), &ehc)
	require.EqualError(t, err, "invalid status code component 'b' at position 3: invalid syntax")
}

func TestEnhancedStatusCode_UnmarshalYAML(t *testing.T) {
	var ehc smtp.EnhancedStatusCode
	err := yaml.Unmarshal([]byte(`"1.2.3"`), &ehc)
	require.NoError(t, err)
	require.Equal(t, smtp.EnhancedStatusCode{1, 2, 3}, ehc)

	err = yaml.Unmarshal([]byte(`"1.2"`), &ehc)
	require.EqualError(t, err, "unexpected value 1.2, expected format x.x.x")

	err = yaml.Unmarshal([]byte(`"1.2.b"`), &ehc)
	require.EqualError(t, err, "invalid status code component 'b' at position 3: invalid syntax")
}
