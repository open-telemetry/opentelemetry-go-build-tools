module go.opentelemetry.io/build-tools/dbotconf

go 1.16

require (
	github.com/spf13/cobra v1.3.0
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/build-tools v0.0.0-00010101000000-000000000000
	golang.org/x/mod v0.5.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace go.opentelemetry.io/build-tools => ../
