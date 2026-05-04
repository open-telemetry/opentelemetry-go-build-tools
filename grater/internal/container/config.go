// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package container

// ExecuteCommandConfig is a group of options for executing a command in a container.
type ExecuteCommandConfig struct {
	containerID string
	cmd         []string
}

// ContainerID returns the container ID.
func (cfg *ExecuteCommandConfig) ContainerID() string {
	return cfg.containerID
}

// Cmd returns the command to execute.
func (cfg *ExecuteCommandConfig) Cmd() []string {
	return cfg.cmd
}

// NewExecuteCommandConfig applies all the options to a returned ExecuteCommandConfig.
func NewExecuteCommandConfig(options ...ExecuteCommandOption) ExecuteCommandConfig {
	config := ExecuteCommandConfig{containerID: "", cmd: []string{}}
	for _, option := range options {
		config = option.apply(config)
	}
	return config
}

// ExecuteCommandOption applies an option to a ExecuteCommandConfig.
type ExecuteCommandOption interface {
	apply(ExecuteCommandConfig) ExecuteCommandConfig
}

type executeCommandOptionFunc func(ExecuteCommandConfig) ExecuteCommandConfig

func (fn executeCommandOptionFunc) apply(cfg ExecuteCommandConfig) ExecuteCommandConfig {
	return fn(cfg)
}

// CreateVolumeConfig is a group of options for creating a volume in a container.
type CreateVolumeConfig struct {
	volumeName string
}

// VolumeName returns the volume name.
func (cfg *CreateVolumeConfig) VolumeName() string {
	return cfg.volumeName
}

// NewCreateVolumeConfig applies all the options to a returned CreateVolumeConfig.
func NewCreateVolumeConfig(options ...CreateVolumeOption) CreateVolumeConfig {
	config := CreateVolumeConfig{volumeName: ""}
	for _, option := range options {
		config = option.apply(config)
	}
	return config
}

// CreateVolumeOption applies an option to a CreateVolumeConfig.
type CreateVolumeOption interface {
	apply(CreateVolumeConfig) CreateVolumeConfig
}

type createVolumeOptionFunc func(CreateVolumeConfig) CreateVolumeConfig

func (fn createVolumeOptionFunc) apply(cfg CreateVolumeConfig) CreateVolumeConfig {
	return fn(cfg)
}

// UseContainerConfig is a group of options for using a container.
type UseContainerConfig struct {
	imageName            string
	bindMounts           map[string]string
	hostToContainerPaths map[string]string
}

// ImageName returns the image name.
func (cfg *UseContainerConfig) ImageName() string {
	return cfg.imageName
}

// BindMounts returns the bind mounts.
func (cfg *UseContainerConfig) BindMounts() map[string]string {
	return cfg.bindMounts
}

// HostToContainerPaths returns the host to container paths.
func (cfg *UseContainerConfig) HostToContainerPaths() map[string]string {
	return cfg.hostToContainerPaths
}

// NewUseContainerConfig applies all the options to a returned UseContainerConfig.
func NewUseContainerConfig(options ...UseContainerOption) UseContainerConfig {
	config := UseContainerConfig{
		imageName:            "",
		bindMounts:           map[string]string{},
		hostToContainerPaths: map[string]string{},
	}
	for _, option := range options {
		config = option.apply(config)
	}
	return config
}

// UseContainerOption applies an option to a UseContainerConfig.
type UseContainerOption interface {
	apply(UseContainerConfig) UseContainerConfig
}

type useContainerOptionFunc func(UseContainerConfig) UseContainerConfig

func (fn useContainerOptionFunc) apply(cfg UseContainerConfig) UseContainerConfig {
	return fn(cfg)
}

// WithContainerID sets the container ID.
func WithContainerID(id string) ExecuteCommandOption {
	return executeCommandOptionFunc(func(cfg ExecuteCommandConfig) ExecuteCommandConfig {
		cfg.containerID = id
		return cfg
	})
}

// WithCommand sets the command to execute.
func WithCommand(cmd []string) ExecuteCommandOption {
	return executeCommandOptionFunc(func(cfg ExecuteCommandConfig) ExecuteCommandConfig {
		cfg.cmd = cmd
		return cfg
	})
}

// WithVolumeName sets the volume name.
func WithVolumeName(name string) CreateVolumeOption {
	return createVolumeOptionFunc(func(cfg CreateVolumeConfig) CreateVolumeConfig {
		cfg.volumeName = name
		return cfg
	})
}

// WithImageName sets the image name.
func WithImageName(name string) UseContainerOption {
	return useContainerOptionFunc(func(cfg UseContainerConfig) UseContainerConfig {
		cfg.imageName = name
		return cfg
	})
}

// WithBindMounts sets the source to container path mappings.
func WithBindMounts(bindMounts map[string]string) UseContainerOption {
	return useContainerOptionFunc(func(cfg UseContainerConfig) UseContainerConfig {
		cfg.bindMounts = bindMounts
		return cfg
	})
}

// WithHostToContainerPaths sets the host to container paths.
func WithHostToContainerPaths(hostToContainerPaths map[string]string) UseContainerOption {
	return useContainerOptionFunc(func(cfg UseContainerConfig) UseContainerConfig {
		cfg.hostToContainerPaths = hostToContainerPaths
		return cfg
	})
}
