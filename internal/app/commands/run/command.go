// Package run implements the "run" command.
package run

import (
	"os"

	"github.com/gloo-foo/vsl/internal/app"
	"github.com/gloo-foo/vsl/internal/container"
	"github.com/gloo-foo/vsl/internal/container/run"
	"github.com/gloo-foo/vsl/internal/script"
	"github.com/urfave/cli/v2"
)

// Command metadata
const (
	Name        = "run"
	usage       = "Run a Docker container with smart mounting"
	argsUsage   = "[command args...]"
	description = `Run a Docker container with automatic directory mounting and git repository awareness.

This command provides intelligent mounting of the current directory and git repositories,
making it easy to run containerized tools and services with access to your local files.

Features:
  - Automatic mounting of current working directory
  - Git repository discovery and mounting
  - Support for UP script files (shebang-style execution)
  - Interactive mode with TTY support
  - Custom volume mounts
  - Network mode configuration

Examples:
  # Run a command in an Ubuntu container
  vsl run --image ubuntu:latest -- echo "Hello World"

  # Run interactively
  vsl run --image alpine:latest --interactive

  # Disable git repository discovery
  vsl run --image node:latest --no-git -- npm test

  # Run with custom volumes
  vsl run --image postgres:latest --volume /data:/var/lib/postgresql/data

  # Execute an UP script file (shebang mode)
  vsl my-script.up arg1 arg2
`
)

// Flag names
const (
	flagImage       = "image"
	flagNoGit       = "no-git"
	flagInteractive = "interactive"
	flagWorkingDir  = "working-dir"
	flagUser        = "user"
	flagEnv         = "env"
	flagVolume      = "volume"
	flagEntrypoint  = "entrypoint"
	flagNetworkMode = "network-mode"
	flagPrivileged  = "privileged"
)

// Package-level config populated by urfave/cli via Destination
var cfg run.Config

var runAction = run.Run

// Command returns the CLI command for running containers
func Command(prefix app.AppEnvPrefix) *cli.Command {
	return &cli.Command{
		Name:        Name,
		Usage:       usage,
		ArgsUsage:   argsUsage,
		Description: description,
		Flags:       flags(prefix),
		Action:      action,
	}
}

// action handles the run command, including script file detection
func action(c *cli.Context) error {
	// Check if we're being used as a shebang interpreter
	// If first arg is a file, try to parse it as an UP script
	if c.NArg() > 0 {
		firstArg := c.Args().Get(0)
		if info, err := os.Stat(firstArg); err == nil && !info.IsDir() {
			// First argument is a file - try to parse as UP script
			scriptCfg, err := script.ParseFile(firstArg)
			if err == nil && scriptCfg != nil {
				scriptCfg.ScriptPath = container.ScriptPath(firstArg)
				scriptCfg.ScriptArgs = c.Args().Slice()[1:]
				return app.Action(c, *scriptCfg, runAction)
			}
			// If parsing failed, fall through to normal CLI mode
		}
	}

	// Normal CLI mode - collect command arguments
	args := c.Args().Slice()
	for _, arg := range args {
		cfg.Command = append(cfg.Command, container.Command(arg))
	}

	// If no image specified and no script, error
	if cfg.Image == "" {
		return cli.Exit("--image flag is required when not running as script interpreter", 1)
	}

	return app.Action(c, cfg, runAction)
}

// flags defines all command flags
func flags(prefix app.AppEnvPrefix) []cli.Flag {
	envPrefix := string(prefix) + "RUN_"

	baseFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        flagImage,
			Aliases:     []string{"i"},
			Usage:       "Docker image to run",
			EnvVars:     []string{envPrefix + "IMAGE"},
			Destination: (*string)(&cfg.Image),
		},
		&cli.BoolFlag{
			Name:        flagNoGit,
			Aliases:     []string{"ng"},
			Usage:       "Disable git repository discovery and mounting",
			EnvVars:     []string{envPrefix + "NO_GIT"},
			Value:       false,
			Destination: &cfg.NoGit,
		},
		&cli.BoolFlag{
			Name:        flagInteractive,
			Aliases:     []string{"it"},
			Usage:       "Run container in interactive mode with TTY",
			EnvVars:     []string{envPrefix + "INTERACTIVE"},
			Value:       true,
			Destination: &cfg.Interactive,
		},
		&cli.StringFlag{
			Name:        flagWorkingDir,
			Aliases:     []string{"w"},
			Usage:       "Working directory inside the container",
			EnvVars:     []string{envPrefix + "WORKING_DIR"},
			Destination: (*string)(&cfg.WorkingDir),
		},
		&cli.StringFlag{
			Name:        flagUser,
			Aliases:     []string{"u"},
			Usage:       "User to run as (uid:gid or username)",
			EnvVars:     []string{envPrefix + "USER"},
			Destination: (*string)(&cfg.User),
		},
		&cli.StringSliceFlag{
			Name:    flagEnv,
			Aliases: []string{"e"},
			Usage:   "Set environment variables (KEY=value)",
			EnvVars: []string{envPrefix + "ENV"},
		},
		&cli.StringSliceFlag{
			Name:    flagVolume,
			Aliases: []string{"v"},
			Usage:   "Bind mount a volume (source:target[:ro])",
			EnvVars: []string{envPrefix + "VOLUME"},
		},
		&cli.StringSliceFlag{
			Name:    flagEntrypoint,
			Usage:   "Override the default entrypoint",
			EnvVars: []string{envPrefix + "ENTRYPOINT"},
		},
		&cli.StringFlag{
			Name:        flagNetworkMode,
			Usage:       "Network mode (bridge, host, none, container:name)",
			EnvVars:     []string{envPrefix + "NETWORK_MODE"},
			Destination: (*string)(&cfg.NetworkMode),
		},
		&cli.BoolFlag{
			Name:        flagPrivileged,
			Usage:       "Give extended privileges to this container",
			EnvVars:     []string{envPrefix + "PRIVILEGED"},
			Value:       false,
			Destination: &cfg.Privileged,
		},
	}

	return app.WithOutputFlags(prefix, &cfg.Output, baseFlags)
}
