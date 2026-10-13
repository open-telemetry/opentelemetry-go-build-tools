// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package environment

import (
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

// RunTestsConfig holds the configuration for running tests.
type RunTestsConfig struct {
	mainModule  module.Module
	baseRef  string
	headRef  string
	dependents []module.Module
	injections []module.Module
}

// MainModule returns the main module for the run tests configuration.
func (cfg *RunTestsConfig) MainModule() module.Module {
	return cfg.mainModule
}

// BaseRef returns the base reference for the run tests configuration.
func (cfg *RunTestsConfig) BaseRef() string {
	return cfg.baseRef
}

// HeadRef returns the head reference for the run tests configuration.
func (cfg *RunTestsConfig) HeadRef() string {
	return cfg.headRef
}

// Dependents returns the dependent modules for the run tests configuration.
func (cfg *RunTestsConfig) Dependents() []module.Module {
	return cfg.dependents
}

// Injections returns the injected modules for the run tests configuration.
func (cfg *RunTestsConfig) Injections() []module.Module {
	return cfg.injections
}

// NewRunTestsConfig applies all the options to a returned RunTestsConfig.
func NewRunTestsConfig(options ...RunTestsOption) RunTestsConfig {
	var config RunTestsConfig
	for _, option := range options {
		config = option.apply(config)
	}
	return config
}

// RunTestsOption applies an option to the RunTestsConfig.
type RunTestsOption interface {
	apply(cfg RunTestsConfig) RunTestsConfig
}

type RunTestsOptionFunc func(RunTestsConfig) RunTestsConfig

func (fn RunTestsOptionFunc) apply(cfg RunTestsConfig) RunTestsConfig {
	return fn(cfg)
}

// WithMainModule sets the main module.
func WithMainModule(mainModule module.Module) RunTestsOption {
	return RunTestsOptionFunc(func(cfg RunTestsConfig) RunTestsConfig {
		cfg.mainModule = mainModule
		return cfg
	})
}

// WithBaseRef sets the base reference.
func WithBaseRef(baseRef string) RunTestsOption {
	return RunTestsOptionFunc(func(cfg RunTestsConfig) RunTestsConfig {
		cfg.baseRef = baseRef
		return cfg
	})
}

// WithHeadRef sets the head reference.
func WithHeadRef(headRef string) RunTestsOption {
	return RunTestsOptionFunc(func(cfg RunTestsConfig) RunTestsConfig {
		cfg.headRef = headRef
		return cfg
	})
}

// WithDependents sets the dependent modules.
func WithDependents(dependents []module.Module) RunTestsOption {
	return RunTestsOptionFunc(func(cfg RunTestsConfig) RunTestsConfig {
		cfg.dependents = dependents
		return cfg
	})
}

// WithInjections sets the injected modules.
func WithInjections(injections []module.Module) RunTestsOption {
	return RunTestsOptionFunc(func(cfg RunTestsConfig) RunTestsConfig {
		cfg.injections = injections
		return cfg
	})
}
