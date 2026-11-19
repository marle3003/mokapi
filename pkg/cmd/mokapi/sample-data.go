package mokapi

import (
	"encoding/json"
	"fmt"
	"io"
	"mokapi/config/dynamic"
	"mokapi/config/static"
	"mokapi/media"
	"mokapi/pkg/cli"
	"mokapi/providers/openapi"
	"mokapi/providers/openapi/schema"
	"mokapi/runtime"
	avro "mokapi/schema/avro/schema"
	"mokapi/schema/encoding"
	"mokapi/schema/json/generator"
	jsonSchema "mokapi/schema/json/schema"
	"mokapi/server"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func NewCmdSampleData() *cli.Command {
	var sampleDataCommand = &cli.Command{
		Name:  "sample-data",
		Use:   "mokapi sample-data [flags] [FILE|URL|SCHEMA]",
		Short: "Generate sample data from a schema",
		Long: `Generates random example data from a schema definition.

The input can be:
  • A local file path (e.g., ./schema.json)
  • A URL (e.g., https://example.com/schema.yaml)
  • A JSON Schema string passed via stdin or directly as an argument.

Supports JSON Schema, OpenAPI Schema, and Avro Schema formats.`,
		Example: `
  # From local file
  mokapi sample-data ./schema.json

  # From URL
  mokapi sample-data https://example.com/schema.yaml

  # From stdin (pipe)
  cat schema.json | mokapi sample-data

  # Direct JSON string
  mokapi sample-data '{"type":"object","properties":{"id":{"type":"integer"}}}'

`,
		Run: func(cmd *cli.Command, args []string) error {
			var input string
			if len(args) > 0 {
				input = args[0]
			}
			return sampleData(cmd.Flags(), input)
		},
	}

	sampleDataCommand.Flags().IntShort("count", "n", 1, cli.FlagDoc{Short: "Number of samples to generate (default 1)"})
	sampleDataCommand.Flags().String("input-type", "jsonschema", cli.FlagDoc{Short: "Type of the input schema: jsonschema|openapi|avro (default: jsonschema)"})
	sampleDataCommand.Flags().StringShort("output", "o", "application/json",
		cli.FlagDoc{Short: "Output format: json, xml, yaml, binary (or full content type, e.g. application/json)"})

	return sampleDataCommand
}

func sampleData(flags *cli.FlagSet, input string) error {
	inputType := flags.GetString("input-type")
	output := flags.GetString("output")
	n := flags.GetInt("count")

	ct := media.ParseContentType(output)
	if ct.Subtype != "" {
		output = ct.Subtype
	} else {
		switch ct.Type {
		case "json":
			ct = media.ParseContentType("application/json")
		case "xml":
			ct = media.ParseContentType("application/xml")
		case "yaml":
			ct = media.ParseContentType("application/yaml")
		case "binary":
			ct = media.ParseContentType("application/octet-stream")
		}
	}

	cw := server.NewConfigWatcher(&static.Config{})

	s, paths, err := readInput(input, &inputType, cw)
	if err != nil {
		return fmt.Errorf("failed to read %s: %v", input, err)
	}

	r := &generator.Request{
		Path: paths,
	}

	var marshal func(v any, ct media.ContentType) ([]byte, error)
	switch inputType {
	case "jsonschema":
		if js, ok := s.(*jsonSchema.Schema); ok {
			r.Schema = js
			e := encoding.NewEncoder(r.Schema)
			marshal = e.Write
		} else {
			return errors.New("expected to get JSON schema from input")
		}
	case "openapi", "swagger":
		if openSchema, ok := s.(*schema.Schema); ok {
			r.Schema = schema.ConvertToJsonSchema(openSchema)
			marshal = openSchema.Marshal
		} else {
			return errors.New("expected to get OpenAPI schema from input")
		}
	case "avro":
		if avroSchema, ok := s.(*avro.Schema); ok {
			js := avro.ConvertToJsonSchema(avroSchema)
			r.Schema = js
			if ct.Subtype == "octet-stream" {
				marshal = func(v any, _ media.ContentType) ([]byte, error) {
					return avroSchema.Marshal(v)
				}
			} else {
				e := encoding.NewEncoder(js)
				marshal = e.Write
			}
		} else {
			return errors.New("expected to get Avro schema from input")
		}
	}

	var items []string
	for i := 0; i < n; i++ {
		v, err := generator.New(r)
		if err != nil {
			return fmt.Errorf("failed to generate data: %v", err)
		}

		b, err := marshal(v, ct)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %v", err)
		}
		items = append(items, string(b))
		r.Context = nil
	}

	switch ct.Subtype {
	case "yaml":
		fmt.Printf("%s", strings.Join(items, "\n---\n"))
	case "xml":
		if n > 1 {
			fmt.Printf("<samples>\n%s\n</samples>", strings.Join(items, "\n"))
		} else if n == 1 {
			fmt.Print(items[0])
		}
	default:
		fmt.Printf("%s", strings.Join(items, "\n"))
	}

	return nil
}

func readInput(input string, inputType *string, r dynamic.Reader) (schema any, path []string, err error) {
	c := &dynamic.Config{}

	// nothing provided, read from stdin (pipe)
	if input == "" {
		c.Raw, err = readPipedStdin()
		if err != nil {
			return nil, nil, err
		}
		if len(c.Raw) == 0 {
			return nil, nil, fmt.Errorf("no input provided (expected file, URL, or stdin)")
		}
	} else if isSchemaString(input) {
		c.Raw = []byte(input)
	} else {
		var u *url.URL
		u, err = url.Parse(input)
		if err != nil {
			return nil, nil, err
		}
		if u.Scheme == "" {
			u.Scheme = "file"
		}

		c, err = r.Read(u, nil)
		if err != nil {
			return nil, nil, err
		}
		return readFromConfig(c, inputType, r)
	}

	err = dynamic.Parse(c, r)
	if err != nil {
		return nil, nil, err
	}
	return readFromConfig(c, inputType, r)
}

// readPipedStdin returns the data piped to stdin, or nil if stdin is a terminal.
func readPipedStdin() ([]byte, error) {
	stat, err := os.Stdin.Stat()
	if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
		return nil, nil
	}
	return io.ReadAll(os.Stdin)
}

func readFromConfig(c *dynamic.Config, inputType *string, r dynamic.Reader) (schema any, path []string, err error) {
	if c.Info.Url == nil || c.Info.Url.Fragment == "" {
		schema, err = parseSchema(c, *inputType, r)
		return
	}

	ref := ""
	if c.Info.Url.Fragment != "" {
		ref = fmt.Sprintf("#%s", c.Info.Url.Fragment)
	}
	if h, ok := runtime.IsHttpConfig(c); ok {
		if ref != "" {
			path = getPathFromOpenAPI(h, ref)
		}
		*inputType = "openapi"
	}

	schema, err = (&dynamic.Reference[any]{Ref: fmt.Sprintf("#%s", c.Info.Url.Fragment)}).Resolve(c, r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve fragment %s: %v", c.Info.Url.Fragment, err)
	}

	if c.Info.Url.Fragment != "" {

	} else {
		if h, ok := runtime.IsHttpConfig(c); ok {
			if h.OpenApi.Major == 2 {
				return nil, nil, fmt.Errorf("failed to generate data for Swagger file %s: requiring fragment", c.Info.Url)
			}
			return nil, nil, fmt.Errorf("failed to generate data for OpenAPI file %s: requiring fragment", c.Info.Url)
		} else if _, ok := runtime.IsAsyncApiConfig(c); ok {
			return nil, nil, fmt.Errorf("failed to generate data for AsyncAPI file %s: requiring fragment", c.Info.Url)
		}
		return nil, nil, fmt.Errorf("unsupported file: %s", c.Info.Url)
	}
	return
}

// isSchemaString reports whether the argument is an inline JSON/YAML schema.
// A schema is always a mapping, whereas a file path or URL is a plain scalar.
func isSchemaString(input string) bool {
	var m map[string]any
	return decode([]byte(input), "", &m) == nil && m != nil
}

func parseSchema(c *dynamic.Config, inputType string, r dynamic.Reader) (any, error) {
	var s interface {
		Parse(*dynamic.Config, dynamic.Reader) error
	}
	switch inputType {
	case "jsonschema":
		s = &jsonSchema.Schema{}
	case "openapi":
		s = &schema.Schema{}
	case "avro":
		s = &avro.Schema{}
	default:
		return nil, fmt.Errorf("unsupported input type: %s", inputType)
	}

	ext := ""
	if c.Info.Url != nil {
		ext = filepath.Ext(c.Info.Url.Path)
	}
	if err := decode(c.Raw, ext, s); err != nil {
		return nil, err
	}

	c.Data = s
	return s, s.Parse(c, r)
}

// decode reads JSON or YAML into v. The file extension decides the format;
// without a known extension (raw input) JSON is tried first, then YAML.
func decode(raw []byte, ext string, v any) error {
	switch ext {
	case ".json":
		return json.Unmarshal(raw, v)
	case ".yaml", ".yml":
		return yaml.Unmarshal(raw, v)
	}
	if json.Unmarshal(raw, v) == nil {
		return nil
	}
	if yaml.Unmarshal(raw, v) == nil {
		return nil
	}
	return fmt.Errorf("input data format not supported")
}

func getPathFromOpenAPI(cfg *openapi.Config, ref string) []string {
	var result []string
	segments := strings.Split(ref, "/")
	switch {
	case strings.HasPrefix(ref, "#/paths") && len(segments) > 2:
		path := segments[2]
		result = append(result, strings.Split(strings.ReplaceAll(path, "~1", "/"), "/")...)
	}

	return result
}
