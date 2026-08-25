module go.opentelemetry.io/build-tools/dbotconf

go 1.25.0

require (
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.12.1
	go.opentelemetry.io/build-tools v0.30.0
	golang.org/x/mod v0.40.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
)

replace go.opentelemetry.io/build-tools => ../
