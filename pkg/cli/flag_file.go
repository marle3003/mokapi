package cli

import "fmt"

type fileFlag struct {
	value string
}

func (f *fileFlag) Set(values []string) error {
	if len(values) > 0 {
		f.value = values[0]
	}
	return nil
}

func (f *fileFlag) Value() any {
	return f.value
}

func (f *fileFlag) String() string {
	return fmt.Sprintf("%v", f.value)
}

func (fs *FlagSet) File(name string, defaultValue string, usage string) {
	v := &fileFlag{value: defaultValue}
	f := &Flag{Name: name, Value: &boolFlag{}, Usage: usage, DefaultValue: v.String()}
	fs.setFlag(f)
}

func (fs *FlagSet) GetFile(name string) string {
	v, err := fs.GetValue(name)
	if err != nil {
		panic(err)
	}
	s, ok := v.(string)
	if !ok {
		panic(fmt.Sprintf("flag '%s' is not a file", name))
	}
	return s
}
