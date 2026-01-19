package flags

import "mokapi/pkg/cli"

func RegisterDataGeneratorFlags(cmd *cli.Command) {
	cmd.Flags().String("data-gen-optional-properties", "0.85", generatorOptionalProperties)
}

var generatorOptionalProperties = cli.FlagDoc{
	Short: "Probability for generating optional properties",
	Long: `Controls how often optional properties are included when generating example or mock data.
The value can be specified either as a number between 0 and 1, or as a predefined string:

- always     → 1.0
- often      → 0.85
- sometimes  → 0.5
- rarely     → 0.15
- never      → 0.0

Higher values result in more optional properties being present in generated data.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--data-gen-optional-properties 0.9\n--data-gen-optional-properties sometimes"},
				{Title: "Env", Source: "MOKAPI_DATA_GEN_OPTIONAL_PROPERTIES=0.9"},
				{Title: "File", Source: "data-gen: optionalProperties: sometimes", Language: "yaml"},
			},
		},
	},
}
