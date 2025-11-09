// Package app provides the application framework including action handlers,
// flag helpers, and output formatting.
package app

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gloo-foo/vsl/internal/app/log"
	"github.com/urfave/cli/v2"
)

// Configurable interface for command configurations
type Configurable interface {
	HasLoggerConfig
	HasOutput
}

// HasLoggerConfig interface for configs that have logger configuration
type HasLoggerConfig interface {
	LoggerConfig() log.Config
}

// HasOutput interface for configs that have output configuration
type HasOutput interface {
	OutputFilePath() FilePath
}

// Runner is a generic function type for command runners
type Runner[CONFIG Configurable, RESULT json.Marshaler] func(context.Context, *slog.Logger, CONFIG) (RESULT, error)

var getLogger = log.GetLogger

// Action is a generic action handler that executes a runner and outputs the result
func Action[C Configurable, R json.Marshaler](c *cli.Context, cfg C, runner Runner[C, R]) error {
	logger := getLogger(c, cfg.LoggerConfig())

	result, err := runner(c.Context, logger, cfg)
	if err != nil {
		return err
	}

	return Output(logger, cfg.OutputFilePath(), result)
}

// Default creates a default action function that combines configuration and runner
func Default[C Configurable, R json.Marshaler](cfg C, runner Runner[C, R]) cli.ActionFunc {
	return func(c *cli.Context) error {
		return Action(c, cfg, runner)
	}
}
