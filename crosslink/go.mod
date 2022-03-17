module go.opentelemetry.io/build-tools/crosslink

go 1.16

require (
	github.com/google/go-cmp v0.5.7
	github.com/spf13/cobra v1.3.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.20.0
)

require (
	github.com/otiai10/copy v1.7.0
	go.opentelemetry.io/build-tools v0.0.0-20220315162859-0719a88c076c
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	golang.org/x/mod v0.5.1
)

replace go.opentelemetry.io/build-tools => ../
