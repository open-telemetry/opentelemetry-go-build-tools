module go.opentelemetry.io/build-tools/crosslink

go 1.16

require (
	github.com/google/go-cmp v0.5.7
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.1
	go.uber.org/zap v1.20.0
)

require (
	github.com/kr/pretty v0.2.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace go.opentelemetry.io/build-tools => ../
