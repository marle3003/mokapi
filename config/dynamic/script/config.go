package script

type Script struct {
	Code     string
	Filename string
}

func New(filename string, code []byte) *Script {
	return &Script{Filename: filename, Code: string(code)}
}
