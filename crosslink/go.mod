module go.opentelemetry.io/build-tools/crosslink

go 1.19

require (
	github.com/google/go-cmp v0.5.9
	github.com/otiai10/copy v1.12.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/build-tools v0.12.0
	go.uber.org/zap v1.26.0
	golang.org/x/mod v0.12.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace go.opentelemetry.io/build-tools => ../
