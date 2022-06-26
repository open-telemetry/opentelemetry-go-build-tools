module go.opentelemetry.io/build-tools/multimod

go 1.15

require (
	github.com/go-git/go-git/v5 v5.4.2
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.12.0
	github.com/stretchr/testify v1.7.2
	go.opentelemetry.io/build-tools v0.0.0-20220321164008-b8e03aff061a
	golang.org/x/mod v0.5.1
)

replace go.opentelemetry.io/build-tools => ../
