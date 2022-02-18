# Crosslink

Crosslink is a tool to assist in managing go repositories that contain multiple
intra-reposistory go.mod files. Crosslink automatically scans and inserts
replace statements for direct and transitive intra-repository dependencies.
Crosslink also contains functionality to remove any extra replace statements
that are no longer required within reason (see below).

## Rules

Crosslink makes certain assumptions about your repository. Due to these
assumptions the tool maintains constraints to avoid any undesirable changes.

1. By default all crosslink actions are non destructive unless specific command
    flags are provided.
2. Crosslink will only work with modules that fall under the root module
    namespace.
   - For example `example.com/crosslink/foo` falls under the
    `example.com/crosslink` namespace but `example.com/bar` does not.
   - Root module namespace is defined either by the module that exists in the
    provided `--root` flag directory or the `go.mod` file located at highest
    level of the repository.
3. Crosslink does not maintain or include version numbers in replace
   statements. Replace statements are always inserted or overwritten with no
   version numbers.
4. Crosslink allows users to `exclude` modules. Exclude means that crosslink
   will not perform any replace or pruning operations where the
   `old path == exclude path`. Operations will still be performed inside
   modules where `module path == exclude path`.

`Note:` Crosslink was developed for use in the OpenTelemetry organization.
See [opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go)
and
[opentelemetry-collector-contrib](https://github.com/open-telemetry/opentelemetry-collector-contrib)
for working examples. If you experience a use case that crosslink fails
too handle properly please open an issue (or even a PR!) highlighting
the discrepancy.

## Usage

### Latest Release

To utilize the latest release of crosslink clone and build the repository and
use the commands listed below.

Crosslink supports the following commands and flags.

### –-root

Used to provide the path to a directory where a go.mod file must exist. If a
root flag is not provided, crosslink will attempt to find the root directory of
a git repository using a tool available in the go-build-tools repo. The root
flag is available to all crosslink subcommands.

**Note: If no --root flag is provided than crosslink attempts to identify a git
repository in the current or a parent directory. If no git repository exists
crosslink will panic.**

    crosslink --root=/users/foo/multimodule-go-repo

### prune / –-prune

`CAUTION: DESTRUCTIVE`

The prune command or flag will run the prune action on the current dependency.
Pruning will remove any dependencies that are not in the current
intra-repository dependency graph. Pruning will only remove go modules that
fall under the same module path as the root module. For example,
If the root module is named `example.com/foo` and there exists a replace
statement of `example.com/foo/bar => ./bar` that is not a direct or transitive
dependency of the current go.mod file, it will be pruned.

**Crosslink will not remove replace statements for modules that do not
fall under the root module path even if they are not in the current
dependency graph.**

Pruning can be executed independently with no replace statements being inserted.

    crosslink prune

Pruning can also be executed in addition to a standard crosslink replace operation.

    crosslink --root=/users/foo/multimodule-go-repo --prune

### –-overwrite

`CAUTION: DESTRUCTIVE`

Overwrite gives crosslink permission to update existing replacement paths
in addition to adding new ones. If crosslink identifies a direct or
transitive dependency in the intra-repository graph then it will insert
or update the corresponding replace statement for that requirement.

    crosslink --root=/users/foo/multimodule-go-repo --overwrite

### –-exclude

Exclude is a set of go modules that crosslink will ignore when replacing or pruning.
It is expected that a list of comma separated values will be provided in one or
multiple calls to exclude. Excluded module names should be the old value in the
replacement path. For example, passing `example.com/test` would exclude this replace
statement from any operation `replace example.com/test => ./test`.

A single call to exclude
    crosslink --overwrite --exclude=github.com/foo/bar/modA,github.com/foo/bar/modB

Multiple calls to exclude can also be made

    crosslink –prune --exclude=example.com/foo/bar/modA,example.com/foo/bar/modB\
    --exclude=example.com/foo/bar/modC \
    --exclude=example.com/foo/bar/modJ,example.com/modZ

### –-verbose

Verbose enables crosslink to log all replace (destructive and non-destructive) and
pruning operations to the terminal. By default this is disabled but enabled
automatically when the –overwrite flag is provided. Verbosity can be enabled or
disabled at any time.
For non destructive operations

    crosslink --root=/users/foo/multimodule-go-repo -v

Can be disabled when overwriting.

    crosslink --root=/users/foot/multimodule-go-repo --overwrite -v=false

`Quick Tip: Make sure your go.mod files are tracked and committed in a VCS
before running crosslink.`
