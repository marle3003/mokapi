package runtime_test

import "mokapi/config/dynamic"

// filesOf returns the files of a snapshot holder (an Info that embeds a configSnapshot).
func filesOf[T any](s interface {
	Snapshot() (*T, []*dynamic.Config)
}) []*dynamic.Config {
	_, files := s.Snapshot()
	return files
}
