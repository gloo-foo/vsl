package app

import (
	"github.com/urfave/cli/v2"
)

// AppEnvPrefix is a type for application environment variable prefixes
type AppEnvPrefix string

// WithOutputFlags appends output flags to the provided flag list
func WithOutputFlags(prefix AppEnvPrefix, output *FilePath, flags []cli.Flag) []cli.Flag {
	return append(flags, OutputFlags(prefix, output)...)
}

// OutputFlags returns standard output flags
func OutputFlags(prefix AppEnvPrefix, output *FilePath) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "output",
			Aliases:     []string{"o"},
			Usage:       "Output file path (default: stdout)",
			EnvVars:     []string{string(prefix) + "OUTPUT"},
			Destination: (*string)(output),
		},
	}
}
