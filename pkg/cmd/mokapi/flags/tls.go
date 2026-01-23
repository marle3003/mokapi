package flags

import "mokapi/pkg/cli"

func RegisterTlsFlags(cmd *cli.Command) {
	cmd.Flags().String("root-ca-cert", "", caCert)
	cmd.Flags().String("root-ca-key", "", caKey)
}

var caCert = cli.FlagDoc{
	Short: "Private key of the root CA",
	Long: `Specifies the private key corresponding to the root CA certificate.
The private key is required to sign generated TLS certificates. It must match the certificate provided via root-ca-cert.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--root-ca-key /path/to/caKey.pem"},
				{Title: "Env", Source: "MOKAPI_ROOT_CA_KEY=/path/to/caKey.pem"},
				{Title: "File", Source: "rootCaKey: /path/to/caKey.pem", Language: "yaml"},
			},
		},
	},
}

var caKey = cli.FlagDoc{
	Short: "Root CA certificate used for signing generated certificates",
	Long: `Specifies the root certificate authority (CA) certificate used by Mokapi to sign generated TLS certificates.
This certificate is used when Mokapi dynamically generates certificates for mocked services that require TLS.`,
	Examples: []cli.Example{
		{
			Codes: []cli.Code{
				{Title: "CLI", Source: "--root-ca-cert /path/to/caCert.pem"},
				{Title: "Env", Source: "MOKAPI_ROOT_CA_CERT=/path/to/caCert.pem"},
				{Title: "File", Source: "rootCaCert: /path/to/caCert.pem", Language: "yaml"},
			},
		},
	},
}
