module go.opentelemetry.io/build-tools/multimod

go 1.15

require (
	github.com/kr/text v0.2.0 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	go.opentelemetry.io/build-tools v0.0.0-20210719163622-92017e64f35b
	golang.org/x/mod v0.4.2
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace go.opentelemetry.io/build-tools => ../
