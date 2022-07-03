module go.opentelemetry.io/build-tools/dbotconf

go 1.16

require (
	github.com/spf13/cobra v1.4.0
	github.com/stretchr/testify v1.8.0
	go.opentelemetry.io/build-tools v0.0.0-20220321164008-b8e03aff061a
	golang.org/x/mod v0.5.1
	gopkg.in/yaml.v3 v3.0.1
)

replace go.opentelemetry.io/build-tools => ../
