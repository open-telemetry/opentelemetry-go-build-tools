# gocp

`gocp` copies Go source code.
It can be used for sharing non-exported,
common code across multiple Go modules.

It is advised to keep the shared code in a separate internal Go module
e.g. under `internal/shared`.
The shared code should be consumed aross modules using `go generate`
e.g. `//go:generate gocp --src=../../internal/shared/env.go --pkg=jaeger --dest=env_shared.go`

`gocp` has been created as one of the solutions
to avoid depending on internal packages of different modules.
Such dependencies could transitively lead to build breaking
when an internal package API introduces a non-backward compatible change.
