// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package container

// ExecuteCommandConfig is a group of options for executing a command in a container.
type ExecuteCommandConfig struct {
	containerID string
	cmd         string
}

// NewExecuteCommandConfig applies all the options to a returned ExecuteCommandConfig.
func NewExecuteCommandConfig(options ...ExecuteCommandOption) ExecuteCommandConfig {
	config := ExecuteCommandConfig{containerID: "", cmd: ""}
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
	imageName              string
	binds                  []string
	hostPaths              []string
}

// NewUseContainerConfig applies all the options to a returned UseContainerConfig.
func NewUseContainerConfig(options ...UseContainerOption) UseContainerConfig {
	config := UseContainerConfig{
		imageName:  "",
		binds:      []string{},
		hostPaths:  []string{},
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
func WithCommand(cmd string) ExecuteCommandOption {
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

// WithBinds sets the binds.
func WithBinds(binds []string) UseContainerOption {
	return useContainerOptionFunc(func(cfg UseContainerConfig) UseContainerConfig {
		cfg.binds = binds
		return cfg
	})
}

// WithHostPaths sets the host paths.
func WithHostPaths(hostPaths []string) UseContainerOption {
	return useContainerOptionFunc(func(cfg UseContainerConfig) UseContainerConfig {
		cfg.hostPaths = hostPaths
		return cfg
	})
}
