package flags

import "mokapi/pkg/cli"

func RegisterEventStoreFlags(cmd *cli.Command) {
	cmd.Flags().Int("event-store-default-size", 100, eventStoreDefaultSize)
	cmd.Flags().String("event-store", "", eventStore)
	cmd.Flags().DynamicInt("event-store-<name>-size", eventStoreName)
}

var eventStoreDefaultSize = cli.FlagDoc{
	Short: "Default maximum number of stored events per API",
	Long: `Defines the default maximum number of events stored per API.
When the limit is reached, older events are discarded. This helps control memory usage while still allowing recent events to be inspected.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--event-store-default-size 500"},
				{Title: "Env", Source: "MOKAPI_EVENT_STORE_DEFAULT_SIZE=500"},
				{Title: "File", Source: "event:\n  store:\n    default: 500", Language: "yaml"},
			},
		},
	},
}

var eventStore = cli.FlagDoc{
	Short: "Configure event store using shorthand syntax",
	Long: `Configures the event store using a shorthand syntax.
The event store keeps track of events produced by mocked APIs, allowing them to be inspected via the API server or web dashboard.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--event-store foo={\"size\":250}"},
				{Title: "Env", Source: "MOKAPI_EVENT_STORE=foo={\"size\":250}"},
				{Title: "File", Source: "event:\n  store:\n    foo: 250", Language: "yaml"},
			},
		},
	},
}

var eventStoreName = cli.FlagDoc{
	Short: "Sets event store size for a specific API",
	Long: `Configures the event store.
The event store keeps track of events produced by mocked APIs, allowing them to be inspected via the API server or web dashboard.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--event-store-foo-size 500"},
				{Title: "Env", Source: "MOKAPI_EVENT_STORE_FOO=500"},
				{Title: "File", Source: "event:\n  store:\n    foo: 250", Language: "yaml"},
			},
		},
	},
}
