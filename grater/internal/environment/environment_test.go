// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build integration
// +build integration

package environment

import (
	"context"
	"fmt"
	"testing"
	"os"
	"path"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/build-tools/grater/internal/container"
	"go.opentelemetry.io/build-tools/grater/internal/dockercontainer"
	"go.opentelemetry.io/build-tools/grater/internal/module"
)

func TestSetUpContainerForTestRemoteDependent(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	dependent := *module.NewModule("github.com/open-telemetry/opentelemetry-go-build-tools", "main")

	useContainerResp, err := env.setUpContainerForTest(ctx, dependent, map[string]string{})
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	lsResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"ls", "/dependent"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, lsResp.ExitCode)
}

func TestSetUpContainerForTestLocalDependent(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	// Create a local dependent directory on the host.
	localModPath := t.TempDir()
	err = os.WriteFile(path.Join(localModPath, "hello.txt"), []byte("hello from dependent"), 0644)
	require.NoError(t, err)

	dependent := *module.NewModule(localModPath, "")

	useContainerResp, err := env.setUpContainerForTest(ctx, dependent, map[string]string{})
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	// Assert the local dependent is accessible inside the container.
	catResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"cat", "/dependent/" + dependent.ModuleName + "/hello.txt"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, catResp.ExitCode, "local dependent should be accessible at /dependent/%s", dependent.ModuleName)
	assert.Equal(t, "hello from dependent", catResp.Output)
}

func TestSetUpContainerForTestBindsAreAccessible(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	// Set up main module first to get a real volume bind.
	mainMod := *module.NewModule("github.com/open-telemetry/opentelemetry-go-build-tools", "")
	mainModResp, err := env.setUpMainModule(ctx, mainMod, "main", "main")
	require.NoError(t, err)
	defer mainModResp.Cleanup()

	localModPath := t.TempDir()
	err = os.WriteFile(path.Join(localModPath, "hello.txt"), []byte("hello from dependent"), 0644)
	require.NoError(t, err)

	dependent := *module.NewModule(localModPath, "")

	useContainerResp, err := env.setUpContainerForTest(ctx, dependent, mainModResp.Binds)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	// Assert the main module volume is accessible via the binds.
	lsResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"ls", "/mainModule/" + mainMod.ModuleName + "/main"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, lsResp.ExitCode, "main module should be accessible via binds at /mainModule/%s/main", mainMod.ModuleName)
}

func TestSetUpMainModuleRemote(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	mod := *module.NewModule("github.com/open-telemetry/opentelemetry-go-build-tools", "")

	resp, err := env.setUpMainModule(ctx, mod, "main", "main")
	require.NoError(t, err)
	defer resp.Cleanup()

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
			container.WithBindMounts(resp.Binds),
		),
	)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	baseRefPath := "/mainModule/" + mod.ModuleName + "/main"
	baseResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"ls", baseRefPath}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, baseResp.ExitCode, "baseRef clone should be accessible at %s", baseRefPath)

	headRefPath := "/mainModule/" + mod.ModuleName + "/main"
	headResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"ls", headRefPath}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, headResp.ExitCode, "headRef clone should be accessible at %s", headRefPath)
}

func TestSetUpMainModuleLocal(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	localModPath := t.TempDir()
	err = os.WriteFile(path.Join(localModPath, "hello.txt"), []byte("hello from local"), 0644)
	require.NoError(t, err)

	mod := *module.NewModule(localModPath, "")

	resp, err := env.setUpMainModule(ctx, mod, "", "")
	require.NoError(t, err)
	defer resp.Cleanup()

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
			container.WithBindMounts(resp.Binds),
		),
	)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	containerModPath := "/mainModule/" + mod.ModuleName
	catResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"cat", containerModPath + "/hello.txt"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, catResp.ExitCode, "local module should be accessible at %s", containerModPath)
	assert.Equal(t, "hello from local", catResp.Output)
}

func TestInjectReplacesModule(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
		),
	)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	injectModulePath := "/modules/injectModule"
	dependentModulePath := "/modules/dependentModulePath"

	// Initialise a module to inject with a method to verify.
	_, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`
mkdir -p %s && cd %s && go mod init example.com/local
cat > hello.go << 'EOF'
package local

func Hello() string {
	return "hello from local"
}
EOF`, injectModulePath, injectModulePath),
			}),
		),
	)
	require.NoError(t, err)

	// Intialise the dependent that depends on a remote module.
	_, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`
mkdir -p %s && cd %s && go mod init example.com/consumer
cat > main.go << 'EOF'
package main

import (
	"fmt"
	"example.com/remote"
)

func main() {
	fmt.Println(local.Hello())
}
EOF`, dependentModulePath, dependentModulePath),
			}),
		),
	)
	require.NoError(t, err)

	// Add the remote module as required in dependent's go.mod.
	_, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`cd %s && go mod edit -require example.com/remote@v0.0.0`, dependentModulePath),
			}),
		),
	)
	require.NoError(t, err)

	// Inject the inject module in place of dummy remote module.
	err = env.inject(ctx, useContainerResp, "example.com/remote@v0.0.0", injectModulePath, dependentModulePath)
	require.NoError(t, err)

	// Assert the presence of replace directive.
	catResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`grep -c "replace example.com/remote" %s/go.mod`, dependentModulePath),
			}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, catResp.ExitCode, "replace directive should be present in go.mod")

	// Assert that the method of inject module is called.
	buildResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`cd %s && go run main.go`, dependentModulePath),
			}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, 0, buildResp.ExitCode, "consumer module should build and run successfully")
	assert.Equal(t, "hello from local", buildResp.Output)
}

func TestCheckoutBranch(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
		),
	)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	mod := *module.NewModule("github.com/open-telemetry/opentelemetry-go-build-tools", "")
	repoPath := "/mainModule/" + mod.ModuleName

	err = env.shallowClone(ctx, useContainerResp, mod, "main", repoPath)
	require.NoError(t, err)

	createResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"git", "-C", repoPath, "checkout", "-b", "test"}),
		),
	)
	require.NoError(t, err)
	require.Equal(t, 0, createResp.ExitCode)

	err = env.checkoutBranch(ctx, useContainerResp, "main", repoPath)
	require.NoError(t, err)

	branchResp, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"git", "-C", repoPath, "branch", "--show-current"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, "main", branchResp.Output)
}

func TestShallowClone(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:1.22"),
		),
	)
	module := *module.NewModule("github.com/open-telemetry/opentelemetry-go-build-tools", "")
	path := "/mainModule/"+module.ModuleName
	err = env.shallowClone(
		ctx,
		useContainerResp,
		module,
		"main",
		path,
	)
	require.NoError(t, err)

	respExecuteCommand, err := env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{"ls", "/mainModule/"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, "opentelemetry-go-build-tools", respExecuteCommand.Output)

	shallowResp, err := env.c.ExecuteCommand(ctx,
    container.NewExecuteCommandConfig(
        container.WithContainerID(useContainerResp.ContainerID),
        container.WithCommand([]string{"git", "-C", path, "rev-parse", "--is-shallow-repository"}),
		),
	)
	require.NoError(t, err)
	assert.Equal(t, "true", shallowResp.Output)
}

func TestGetTestReportFailingTest(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:alpine"),
		),
	)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	targetPath := "/testModule"
	_, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`
mkdir -p %s && cd %s && go mod init failing-test-module
cat > fail_test.go << 'EOF'
package main

import "testing"

func TestFail(t *testing.T) {
	t.Fatal("intentional failure")
}
EOF`, targetPath, targetPath),
			}),
		),
	)
	require.NoError(t, err)

	respReport, err := env.getTestReport(ctx, useContainerResp, targetPath)
	require.NoError(t, err)
	assert.Equal(t, 1, respReport.ExitCode)
}

func TestGetTestReportFailingBuild(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:alpine"),
		),
	)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	targetPath := "/testModule"
	_, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`
mkdir -p %s && cd %s && go mod init failing-build-module
cat > broken.go << 'EOF'
package main

func Broken() {
	this is not valid go syntax at all
}
EOF`, targetPath, targetPath),
			}),
		),
	)
	require.NoError(t, err)

	respReport, err := env.getTestReport(ctx, useContainerResp, targetPath)
	require.NoError(t, err)
	assert.Equal(t, 1, respReport.ExitCode)
}

func TestGetTestReportPass(t *testing.T) {
	ctx := context.Background()

	dc, err := dockercontainer.NewDockerContainer()
	require.NoError(t, err)

	env := NewEnvironment(dc)

	useContainerResp, err := env.c.UseContainer(ctx,
		container.NewUseContainerConfig(
			container.WithImageName("golang:alpine"),
		),
	)
	require.NoError(t, err)
	defer useContainerResp.Cleanup()

	targetPath := "/testModule"
	_, err = env.c.ExecuteCommand(ctx,
		container.NewExecuteCommandConfig(
			container.WithContainerID(useContainerResp.ContainerID),
			container.WithCommand([]string{
				"sh", "-c",
				fmt.Sprintf(`
mkdir -p %s && cd %s && go mod init passing-module
cat > pass_test.go << 'EOF'
package main

import "testing"

func TestPass(t *testing.T) {
}
EOF`, targetPath, targetPath),
			}),
		),
	)
	require.NoError(t, err)

	respReport, err := env.getTestReport(ctx, useContainerResp, targetPath)
	require.NoError(t, err)
	assert.Equal(t, 0, respReport.ExitCode)
}

func TestMergeMaps(t *testing.T) {
	a := map[string]string{"key1": "value1"}
	b := map[string]string{"key2": "value2"}

	comb := mergeMaps(a, b)

	assert.Equal(t, "value1", comb["key1"])
	assert.Equal(t, "value2", comb["key2"])
	assert.Len(t, comb, 2)
}

