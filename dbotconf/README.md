# dbotconf

## Overview

`dbotconf` is a Go-based tool designed to assist with configuration management
of the dependabot system. It is part of the OpenTelemetry build-tools suite.

## Installation

To install `dbotconf`, ensure you have Go installed and run:

```bash
go install go.opentelemetry.io/build-tools/dbotconf@latest
```

## Usage

```terminal
dbotconf manages Dependabot configuration for multi-module Go repositories.

Usage:
  dbotconf [command]

Examples:

  dbotconf generate > .github/dependabot.yml

  dbotconf verify .github/dependabot.yml

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  generate    Generate Dependabot configuration
  help        Help about any command
  verify      Verify Dependabot configuration is complete

Flags:
  -h, --help             help for dbotconf
      --ignore strings   glob patterns to ignore

Use "dbotconf [command] --help" for more information about a command.
```

The `generate` command creates configuration files based on predefined templates.

```bash
dbotconf generate
```

The `verify` command checks the validity of existing configuration files.

```bash
dbotconf verify
```

## License

This project is licensed under the Apache License 2.0.
