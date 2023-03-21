# gotmpl

`gotmpl` is a tool that uses Go's [`text/template`](https://pkg.go.dev/text/template)
for generting generating text files
based on file templates and input JSON data.

## Usage

```sh
Usage of gotmpl:
  -b, --body string   Template body filepath.
  -d, --data string   Data in JSON format.
  -o, --out string    Output filepath.
```

## Background

`gotmpl` has been created as one of the solutions
to avoid depending on internal packages of different modules.
Such dependencies could transitively lead to build failure
when an internal package API introduces a non-backward compatible change.

`gotmpl` can be used for sharing non-exported,
common code across multiple Go modules.
It is advised to keep the shared code in a separate internal Go module
e.g. under `internal/shared`.
The shared code should be consumed across modules using `go generate`,
for example:

```go
//go:generate gotmpl --body=../internal/shared/env.go.tmpl "--data={ \"pkg\": \"jaeger\" }" --out=env.go
```
