module go.opentelemetry.io/build-tools/multimod

go 1.15

require (
	github.com/go-git/go-git/v5 v5.4.2
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/build-tools v0.0.0-20210719163622-92017e64f35b
	golang.org/x/mod v0.4.2
)

replace go.opentelemetry.io/build-tools => ../
