# opentelemetry-go-build-tools

[![Go Reference](https://pkg.go.dev/badge/go.opentelemetry.io/build-tools.svg)](https://pkg.go.dev/go.opentelemetry.io/build-tools)
[![Go Report Card](https://goreportcard.com/badge/github.com/open-telemetry/opentelemetry-go-build-tools)](https://goreportcard.com/report/github.com/open-telemetry/opentelemetry-go-build-tools)
[![ci](https://github.com/open-telemetry/opentelemetry-go-build-tools/actions/workflows/ci.yml/badge.svg)](https://github.com/open-telemetry/opentelemetry-go-build-tools/actions/workflows/ci.yml)

Build tools for use by the Go API/SDK, the collector, and their associated
contrib repositories

[Contributing](CONTRIBUTING.md)

## Tools

This repository provides tooling for OpenTelemetry Go projects. Below are
overviews and examples of each provided tools:

### [`gotmpl`](./gotmpl/README.md)

`gotmpl` generates files from Go templates and JSON data.

```sh
gotmpl --body=template.tmpl --data='{"key":"value"}' --out=output.go
```

`gotmpl` is designed to be used with `go generate` to generate code or
configuration files from templates.

```go
//go:generate gotmpl --body=../internal/shared/env.go.tmpl "--data={ \"pkg\": \"jaeger\" }" --out=env.go
```

### [`issuegenerator`](./issuegenerator/README.md)

Generates an issue if any test fails in CI.

### [`chloggen`](./chloggen/README.md)

Generates a `CHANGELOG` file from individual change YAML files.

```sh
# generates a new change YAML file from a template
chloggen new -filename <filename>
# validates all change YAML files
chloggen validate
# provide a preview of the generated changelog file
chloggen update -dry
# updates the changelog file
chloggen update -version <version>
```

### [`checkapi`](./checkapi/README.md)

Analyzes a Go module's API surface and enforces restrictions.

```sh
checkapi -folder . -config config.yaml
```

### [`crosslink`](./crosslink/README.md)

Manages multiple `go.mod` files and intra-repository dependencies.

```sh
# Insert/overwrite replace statements for intra-repo dependencies
crosslink --root=/path/to/repo

# Remove unnecessary replace statements
crosslink prune
crosslink --root=/path/to/repo --prune

# Overwrite existing replace statements
crosslink --root=/path/to/repo --overwrite

# Generate or update go.work file
crosslink work --root=/path/to/repo
```

### [`checkfile`](./checkfile/README.md)

Checks that components in OpenTelemetry core and contrib repos contain a
required file.

```sh
checkfile --project-path path/to/project \
          --component-rel-path service/defaultcomponents/defaults.go \
          --module-name go.opentelemetry.io/collector \
          --file-name README.md
```

### [`githubgen`](./githubgen/README.md)

Generates `.github/CODEOWNERS` and `.github/ALLOWLIST` files.

```sh
githubgen --skipgithub --folder . --github-org "open-telemetry" \
  --default-codeowner open-telemetry/opentelemetry-collector-approvers \
  --allowlist cmd/githubgen/allowlist.txt
```

To authenticate for GitHub API, set a `GITHUB_TOKEN` environment variable.

### [`dbotconf`](./dbotconf/README.md)

`dbotconf` is a Go-based tool for management of [dependabot] configuration. It
provides the `generate` and `verify` commands to create and validate
[dependabot configuration files].

```sh
# Generate configuration files
dbotconf generate

# Verify existing configuration files
dbotconf verify
```

[dependabot]: https://github.com/dependabot
[dependabot configuration files]: https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference#about-the-dependabotyml-file

### [`multimod`](./multimod/README.md)

Tooling to support versioning of multiple Go modules in a repository.

```sh
# Verify module versioning configuration.
./multimod verify
# Prepare a pre-release commit.
./multimod prerelease --module-set-name <name>
# Tag the new release commit.
./multimod tag --commit-hash <hash>
```
