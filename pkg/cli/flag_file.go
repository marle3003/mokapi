package cli

import "fmt"

type fileFlag struct {
	value         string
	isSet         bool
	setConfigFile func(string)
	source        Source
}

func (f *fileFlag) Set(values []string, source Source) error {
	if f.source > source {
		return nil
	}

	if len(values) > 0 {
		f.value = values[0]
		f.setConfigFile(f.value)
		f.source = source
	}
	return nil
}

func (f *fileFlag) Value() any {
	return f.value
}

func (f *fileFlag) IsSet() bool {
	return f.isSet
}

func (f *fileFlag) String() string {
	return fmt.Sprintf("%v", f.value)
}

func (fs *FlagSet) File(name string, doc FlagDoc) *FlagBuilder {
	v := &fileFlag{setConfigFile: fs.setConfigFile}
	f := &Flag{Name: name, Value: v, FlagDoc: doc, Type: "file"}
	fs.setFlag(f)
	return &FlagBuilder{flag: f}
}
