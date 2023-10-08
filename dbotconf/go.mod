module go.opentelemetry.io/build-tools/dbotconf

go 1.19

require (
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/build-tools v0.12.0
	golang.org/x/mod v0.13.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace go.opentelemetry.io/build-tools => ../
