package dynamic

type ChangeEventHandler func(path string, c Config, r ConfigReader) (bool, Config)

type Config interface {
}

type ConfigReader interface {
	Read(path string, c Config, h ChangeEventHandler) error
}
