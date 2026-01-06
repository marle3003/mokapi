package cli

import "fmt"

type fileFlag struct {
	value         string
	isSet         bool
	setConfigFile func(string)
}

func (f *fileFlag) Set(values []string) error {
	if len(values) > 0 {
		f.value = values[0]
		f.setConfigFile(f.value)
	}
	return nil
}

func (f *fileFlag) Value() any {
	return f.value
}

func (f *fileFlag) IsSet() bool {
	return f.isSet
}

func (f *fileFlag) Type() string {
	return "file"
}

func (f *fileFlag) String() string {
	return fmt.Sprintf("%v", f.value)
}

func (fs *FlagSet) File(name string, usage string) {
	v := &fileFlag{setConfigFile: fs.setConfigFile}
	f := &Flag{Name: name, Value: v, Usage: usage}
	fs.setFlag(f)
}
