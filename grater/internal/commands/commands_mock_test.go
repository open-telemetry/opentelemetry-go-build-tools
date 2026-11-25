// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/mockcontainer"
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

// goModSubcommand returns the go mod subcommand (e.g. "edit", "tidy") from a
// command invocation shaped like {"go", "mod", <subcommand>, ...}.
func goModSubcommand(args []string) string {
	if len(args) < 3 {
		return ""
	}
	return args[2]
}

func TestGetModuleFromProxy_Mock(t *testing.T) {
	tests := []struct {
		name    string
		execErr error
		wantErr string
	}{
		{
			name: "success",
		},
		{
			name:    "proxy download fails",
			execErr: errors.New("go: module lookup disabled by GOPROXY=off"),
			wantErr: "module lookup disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockcontainer.NewMockDockerContainer()
			mock.ExecuteCommandMock = func(_ context.Context, _ container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error) {
				if tt.execErr != nil {
					return container.ExecuteCommandResponse{}, tt.execErr
				}
				return container.ExecuteCommandResponse{ExitCode: 0}, nil
			}

			mod := module.NewModule("go.opentelemetry.io/otel", "v1.24.0")
			useContainerResp := container.UseContainerResponse{ContainerID: "fake-container-id"}

			err := GetModuleFromProxy(context.Background(), mock, useContainerResp, *mod, "/module/")

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSetReplaceDirective_Mock(t *testing.T) {
	tests := []struct {
		name      string
		editErr   error
		tidyErr   error
		wantErr   bool
		wantOrder []string
	}{
		{
			name:      "success",
			wantOrder: []string{"edit", "tidy"},
		},
		{
			name:      "edit fails, tidy must not run",
			editErr:   errors.New("malformed replace directive"),
			wantErr:   true,
			wantOrder: []string{"edit"},
		},
		{
			name:      "edit succeeds, tidy fails",
			tidyErr:   errors.New("missing go.sum entry"),
			wantErr:   true,
			wantOrder: []string{"edit", "tidy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mu    sync.Mutex
				order []string
			)

			mock := mockcontainer.NewMockDockerContainer()
			mock.ExecuteCommandMock = func(_ context.Context, cfg container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error) {
				sub := goModSubcommand(cfg.Cmd())

				mu.Lock()
				order = append(order, sub)
				mu.Unlock()

				switch sub {
				case "edit":
					if tt.editErr != nil {
						return container.ExecuteCommandResponse{}, tt.editErr
					}
					return container.ExecuteCommandResponse{ExitCode: 0}, nil
				case "tidy":
					if tt.tidyErr != nil {
						return container.ExecuteCommandResponse{}, tt.tidyErr
					}
					return container.ExecuteCommandResponse{ExitCode: 0}, nil
				default:
					return container.ExecuteCommandResponse{}, errors.New("unexpected go mod subcommand: " + sub)
				}
			}

			useContainerResp := container.UseContainerResponse{ContainerID: "fake-container-id"}

			err := SetReplaceDirective(context.Background(), mock, useContainerResp, "go.opentelemetry.io/build-tools/grater/internal/testdata/module", "../moduleFail", "/dependent/")

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mu.Lock()
			defer mu.Unlock()
			assert.Equal(t, tt.wantOrder, order, "unexpected go mod subcommand execution order")
		})
	}
}

func TestRunModuleTest_Mock(t *testing.T) {
	tests := []struct {
		name     string
		execErr  error
		execResp container.ExecuteCommandResponse
		wantErr  bool
		wantExit int
	}{
		{
			name:     "tests pass",
			execResp: container.ExecuteCommandResponse{ExitCode: 0, Output: "PASS"},
			wantExit: 0,
		},
		{
			name:    "container exec transport error",
			execErr: errors.New("container not running"),
			wantErr: true,
		},
		{
			name:     "go test reports a failure without a transport error",
			execResp: container.ExecuteCommandResponse{ExitCode: 1, Output: "--- FAIL: TestX"},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := mockcontainer.NewMockDockerContainer()
			mock.ExecuteCommandMock = func(_ context.Context, _ container.ExecuteCommandConfig) (container.ExecuteCommandResponse, error) {
				return tt.execResp, tt.execErr
			}

			useContainerResp := container.UseContainerResponse{ContainerID: "fake-container-id"}

			resp, err := RunModuleTest(context.Background(), mock, useContainerResp, "/dependent/")

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, container.ExecuteCommandResponse{}, resp)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantExit, resp.ExitCode)
			assert.Equal(t, tt.execResp.Output, resp.Output)
		})
	}
}
