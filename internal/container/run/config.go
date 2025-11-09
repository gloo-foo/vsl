package run

import (
	"github.com/gloo-foo/vsl/internal/app"
	"github.com/gloo-foo/vsl/internal/app/log"
	"github.com/gloo-foo/vsl/internal/container"
)

// Config holds configuration for running a container.
type Config struct {
	// Container configuration
	Image       container.Image         `up:"image"`        // Docker image to run
	Command     []container.Command     `up:"command"`      // Command to execute
	Entrypoint  []container.Entrypoint  `up:"entrypoint"`   // Container entrypoint
	WorkingDir  container.WorkingDir    `up:"workdir"`      // Working directory
	Environment []container.Environment `up:"env"`          // Environment variables
	Volumes     []container.Volume      `up:"volume"`       // Volume mounts
	User        container.User          `up:"user"`         // User to run as
	NetworkMode container.NetworkMode   `up:"network_mode"` // Network mode

	// Behavior flags
	Interactive bool `up:"interactive"` // Run interactively with TTY
	NoGit       bool `up:"-"`           // Disable git repository discovery
	Privileged  bool `up:"privileged"`  // Run in privileged mode

	// Script handling
	ScriptPath container.ScriptPath `up:"-"` // Path to UP script file (if running as interpreter)
	ScriptArgs []string             `up:"-"` // Arguments passed to the script

	// Output and logging
	Output  app.FilePath `up:"-"`
	Logging log.Config   `up:"-"`
}

func (c Config) OutputFilePath() app.FilePath { return c.Output }
func (c Config) LoggerConfig() log.Config     { return c.Logging }
