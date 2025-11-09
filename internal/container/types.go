// Package container provides types for container configuration.
package container

// Image represents a Docker image name with optional tag.
type Image string

// Command represents a command to execute in the container.
type Command string

// Entrypoint represents a container entrypoint.
type Entrypoint string

// WorkingDir represents a container working directory path.
type WorkingDir string

// Environment represents an environment variable in KEY=VALUE format.
type Environment string

// Volume represents a volume mount specification (source:target[:ro]).
type Volume string

// User represents a container user (uid:gid or username).
type User string

// NetworkMode represents a container network mode (bridge, host, none, etc.).
type NetworkMode string

// ContainerID represents a Docker container identifier.
type ContainerID string

// GitRoot represents the root directory of a git repository.
type GitRoot string

// GitDir represents the path to a .git directory.
type GitDir string

// ScriptPath represents the path to a script file.
type ScriptPath string

// MountSource represents the source path for a bind mount.
type MountSource string

// MountTarget represents the target path for a bind mount.
type MountTarget string
