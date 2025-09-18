package mail

import (
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/dynamic/dynamictest"
	"mokapi/providers/mail"
	"testing"
)

func TestConfig_Parse(t *testing.T) {
	hook := test.NewGlobal()
	logrus.SetOutput(io.Discard)

	var cfg *Config
	err := cfg.Parse(&dynamic.Config{Info: dynamictest.NewConfigInfo()}, &dynamictest.Reader{})
	require.NoError(t, err)
	require.Len(t, hook.Entries, 0)

	src := `smtp: 1.0`
	err = yaml.Unmarshal([]byte(src), &cfg)
	require.NoError(t, err)

	c := &dynamic.Config{Info: dynamictest.NewConfigInfo()}
	err = cfg.Parse(c, &dynamictest.Reader{})
	require.NoError(t, err)

	require.IsType(t, &mail.Config{}, c.Data)

	require.Len(t, hook.Entries, 1)
	require.Equal(t, logrus.WarnLevel, hook.Entries[0].Level)
	require.Equal(t, "Deprecated mail configuration in file://foo.yml. This format is deprecated and will be removed in future versions. Please migrate to the new format. More info: https://mokapi.io/docs/guides/email", hook.Entries[0].Message)
}
