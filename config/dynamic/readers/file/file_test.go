package file

import (
	"fmt"
	"mokapi/config/dynamic/common"
	"mokapi/test"
	"strings"
	"testing"
)

type foo struct {
	Bar string `json:"bar"`
}

func TestYaml(t *testing.T) {
	t.Run("empty unknown", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte(""), nil
		}
		f, err := r.Read(MustParseUrl("foo.yml"))
		test.Equals(t, common.UnknownFile, err)
		test.Equals(t, nil, f)
	})
	t.Run("empty", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte(""), nil
		}
		f, err := r.Read(MustParseUrl("foo.yml"), common.WithData(&foo{}))
		test.Ok(t, err)
		test.Equals(t, &foo{}, f.Data)
	})
	t.Run("simple", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte("bar: foobar"), nil
		}
		f, err := r.Read(MustParseUrl("foo.yml"), common.WithData(&foo{}))
		test.Ok(t, err)
		test.Equals(t, "foobar", f.Data.(*foo).Bar)
	})
	t.Run("syntax error", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte("bar- foobar"), nil
		}
		_, err := r.Read(MustParseUrl("foo.yml"), common.WithData(&foo{}))
		test.Error(t, err)
		test.Assert(t, strings.HasPrefix(err.Error(), "parsing yaml file"), "yaml parser error")
		test.Assert(t, len(r.files) == 0, "files empty")
	})
}

func TestTxt(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte(""), nil
		}
		f, err := r.Read(MustParseUrl("foo.txt"))
		test.Ok(t, err)
		test.Equals(t, "", f.Data)
	})
	t.Run("text", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte("foobar"), nil
		}
		f, err := r.Read(MustParseUrl("foo.txt"))
		test.Ok(t, err)
		test.Equals(t, "foobar", f.Data)
	})
	t.Run("read second time", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return nil, fmt.Errorf("read file request: %v", s)
		}
		u := MustParseUrl("foo.txt")
		r.files[r.name(u)] = &common.File{Data: "foobar"}
		f, err := r.Read(MustParseUrl("foo.txt"))
		test.Ok(t, err)
		test.Equals(t, "foobar", f.Data)
	})
	t.Run("read second time with options", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return nil, fmt.Errorf("read file request: %v", s)
		}
		u := MustParseUrl("foo.txt")
		r.files[r.name(u)] = &common.File{Data: "foobar"}
		ch := make(chan *common.File, 1)
		f1, err := r.Read(u, common.WithListener(ch))
		test.Ok(t, err)
		f1.Changed()
		f2 := <-ch
		test.Equals(t, f1, f2)
	})
}

func TestJson(t *testing.T) {
	t.Run("empty unknown", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte("{}"), nil
		}
		f, err := r.Read(MustParseUrl("foo.json"))
		test.Equals(t, common.UnknownFile, err)
		test.Equals(t, nil, f)
	})
	t.Run("empty", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte("{}"), nil
		}
		f, err := r.Read(MustParseUrl("foo.json"), common.WithData(&foo{}))
		test.Ok(t, err)
		test.Equals(t, &foo{}, f.Data)
	})
	t.Run("simple", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte(`{"bar": "foobar"}`), nil
		}
		f, err := r.Read(MustParseUrl("foo.json"), common.WithData(&foo{}))
		test.Ok(t, err)
		test.Equals(t, "foobar", f.Data.(*foo).Bar)
	})
	t.Run("syntax error", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte("bar }"), nil
		}
		_, err := r.Read(MustParseUrl("foo.json"), common.WithData(&foo{}))
		test.Error(t, err)
		test.Assert(t, strings.HasPrefix(err.Error(), "parsing json file"), "json parser error")
		test.Assert(t, len(r.files) == 0, "files empty")
	})
}

func TestTemplate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte(""), nil
		}
		_, err := r.Read(MustParseUrl("foo.tmpl"))
		test.Ok(t, err)
	})
	t.Run("simple", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte(`{{trim "   foobar    "}}`), nil
		}
		f, err := r.Read(MustParseUrl("foo.tmpl"), common.WithData(&foo{}))
		test.Ok(t, err)
		test.Equals(t, "foobar", f.Data)
	})
	t.Run("with yaml", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte(`bar: {{trim "    foobar     "}}`), nil
		}
		f, err := r.Read(MustParseUrl("foo.yml.tmpl"), common.WithData(&foo{}))
		test.Ok(t, err)
		test.Equals(t, "foobar", f.Data.(*foo).Bar)
	})
	t.Run("syntax error", func(t *testing.T) {
		r := New(nil)
		r.readFileFunc = func(s string) ([]byte, error) {
			return []byte("{{bar"), nil
		}
		_, err := r.Read(MustParseUrl("foo.tmpl"), common.WithData(&foo{}))
		test.Error(t, err)
		test.Assert(t, strings.HasPrefix(err.Error(), "template"), "template parser error")
		test.Assert(t, len(r.files) == 0, "files empty")
	})
}
