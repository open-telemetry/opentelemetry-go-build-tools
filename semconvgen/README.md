# ⛔️ DEPRECATED - Go Semantic Convention Generator

> [!WARNING]
> This tool is deprecated and will be removed in a future release.
> Use [weaver] instead.

This tool is designed to generate constants in a semantic convention package
for the Go API and the collector.
It may be used by other systems,
but it's primary function beyond invoking the template processor is to ensure
that generated identifiers conform to Go's naming idiom,
particularly with respect to initialisms and acronyms.
Other users may be served just as well by using the template processor directly.

## Installation

```shell
go get go.opentelemetry.io/build-tools/semconvgen
```

## Usage

```shell
semconvgen -i <path to spec YAML> -t <path to template> -o <path to output>
```

A full list of available options:

````txt
  -z, --capitalizations-path string   Path to a file containing additional newline-separated capitalization strings.
  -c, --container string              Container image ID (default "otel/semconvgen")
  -f, --filename string               Filename for templated output. If not specified 'basename(inputPath).go' will be used.
  -i, --input string                  Path to semantic convention definition YAML. Should be a directory in the specification git repository.
      --only string                   Process only semantic conventions of the specified type. {span, resource, event, metric_group, metric, units, scope, attribute_group}
  -o, --output string                 Path to output target. Must be either an absolute path or relative to the repository root. If unspecified will output to a sub-directory with the name matching the version number specified via --specver flag.
  -p, --parameters string             List of key=value pairs separated by comma. These values are fed into the template as-is.
  -s, --specver string                Version of semantic convention to generate. Must be an existing version tag in the specification git repository.
  -t, --template string               Template filename (default "template.j2")```
````

[weaver]: https://github.com/open-telemetry/weaver
