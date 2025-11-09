// Package run contains the container run logic.
package run

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	cont "github.com/gloo-foo/vsl/internal/container"
	"github.com/gloo-foo/vsl/internal/git"
)

// Result holds the result of a container run.
type Result struct {
	Success     bool             `json:"success"`
	ContainerID cont.ContainerID `json:"container_id"`
	Image       cont.Image       `json:"image"`
	WorkingDir  cont.WorkingDir  `json:"working_dir"`
	Mounts      []MountInfo      `json:"mounts"`
	GitRoot     cont.GitRoot     `json:"git_root,omitempty"`
	ScriptPath  cont.ScriptPath  `json:"script_path,omitempty"`
	Message     string           `json:"message"`
}

// MountInfo represents mount information for JSON output.
type MountInfo struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// MarshalJSON implements json.Marshaler
func (r Result) MarshalJSON() ([]byte, error) {
	type Alias Result
	return json.Marshal((Alias)(r))
}

// Run executes the container run logic.
func Run(ctx context.Context, logger *slog.Logger, cfg Config) (Result, error) {
	logger.Info("Starting container run",
		"image", cfg.Image,
		"interactive", cfg.Interactive,
		"no_git", cfg.NoGit,
	)

	// Initialize Docker client
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return Result{}, fmt.Errorf("failed to create docker client: %w", err)
	}
	defer func(dockerCli *client.Client) {
		err := dockerCli.Close()
		if err != nil {
			panic(err)
		}
	}(dockerCli)

	// Get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		return Result{}, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Build base mounts
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: pwd,
			Target: pwd,
		},
	}

	var gitRoot cont.GitRoot

	// Handle git repository discovery
	if !cfg.NoGit {
		logger.Debug("Discovering git repository")
		foundGitRoot, err := git.FindRoot(pwd)
		if err == nil && foundGitRoot != "" && string(foundGitRoot) != pwd {
			gitRoot = foundGitRoot
			logger.Info("Found git repository", "root", gitRoot)
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: string(foundGitRoot),
				Target: string(foundGitRoot),
			})

			realGitDir, err := git.FindRealGitDir(foundGitRoot)
			if err == nil && realGitDir != "" {
				gitDirPath := filepath.Join(string(foundGitRoot), ".git")
				if string(realGitDir) != gitDirPath {
					logger.Debug("Mounting real git directory", "path", realGitDir)
					mounts = append(mounts, mount.Mount{
						Type:   mount.TypeBind,
						Source: string(realGitDir),
						Target: gitDirPath,
					})
				}
			}
		}
	}

	// Configure from script or CLI
	image := cfg.Image
	cmd := make([]string, len(cfg.Command))
	for i, c := range cfg.Command {
		cmd[i] = string(c)
	}
	entrypoint := make([]string, len(cfg.Entrypoint))
	for i, e := range cfg.Entrypoint {
		entrypoint[i] = string(e)
	}
	env := make([]string, len(cfg.Environment))
	for i, e := range cfg.Environment {
		env[i] = string(e)
	}
	workingDir := string(cfg.WorkingDir)
	user := string(cfg.User)
	networkMode := string(cfg.NetworkMode)
	stdinOpen := cfg.Interactive
	tty := cfg.Interactive
	privileged := cfg.Privileged

	// If running from script, append script args to command
	if cfg.ScriptPath != "" {
		logger.Info("Running from script", "path", cfg.ScriptPath)
		cmd = append(cmd, cfg.ScriptArgs...)
	}

	// Default working dir to pwd if not specified
	if workingDir == "" {
		workingDir = pwd
	}

	logger.Debug("Container configuration",
		"image", image,
		"working_dir", workingDir,
		"user", user,
		"privileged", privileged,
		"network_mode", networkMode,
	)

	// Container configuration
	containerConfig := &container.Config{
		Image:        string(image),
		Cmd:          cmd,
		Entrypoint:   entrypoint,
		WorkingDir:   workingDir,
		Env:          env,
		User:         user,
		Tty:          tty,
		AttachStdin:  stdinOpen,
		AttachStdout: true,
		AttachStderr: true,
		OpenStdin:    stdinOpen,
	}

	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		AutoRemove:  true,
		Privileged:  privileged,
		NetworkMode: container.NetworkMode(networkMode),
	}

	// Create container
	logger.Info("Creating container")
	resp, err := dockerCli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return Result{}, fmt.Errorf("failed to create container: %w", err)
	}

	containerID := cont.ContainerID(resp.ID)
	logger.Info("Container created", "id", containerID)

	// Start container
	logger.Info("Starting container")
	if err := dockerCli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return Result{}, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to finish
	logger.Debug("Waiting for container to complete")
	statusCh, errCh := dockerCli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return Result{}, fmt.Errorf("error waiting for container: %w", err)
		}
	case <-statusCh:
	}

	logger.Info("Container completed successfully")

	// Build mount info for result
	mountInfo := make([]MountInfo, len(mounts))
	for i, m := range mounts {
		mountInfo[i] = MountInfo{
			Source: m.Source,
			Target: m.Target,
		}
	}

	return Result{
		Success:     true,
		ContainerID: containerID,
		Image:       image,
		WorkingDir:  cont.WorkingDir(workingDir),
		Mounts:      mountInfo,
		GitRoot:     gitRoot,
		ScriptPath:  cfg.ScriptPath,
		Message:     "Container executed successfully",
	}, nil
}
