module go.opentelemetry.io/build-tools/dbotconf

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/spf13/cobra v1.4.0
	github.com/stretchr/testify v1.7.1
	go.opentelemetry.io/build-tools v0.0.0-20220304161722-feb5ff5848f1
	golang.org/x/mod v0.5.1
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace go.opentelemetry.io/build-tools => ../
