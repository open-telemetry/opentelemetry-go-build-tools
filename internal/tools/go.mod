module go.opentelemetry.io/build-tools/internal/tools

go 1.15

require (
	github.com/client9/misspell v0.3.4
	github.com/gogo/protobuf v1.3.2
	github.com/golangci/golangci-lint v1.45.2
	github.com/itchyny/gojq v0.12.9
	go.opentelemetry.io/build-tools/dbotconf v0.0.0-20220321164008-b8e03aff061a
	golang.org/x/tools v0.1.12
)

replace (
	go.opentelemetry.io/build-tools => ../../
	go.opentelemetry.io/build-tools/dbotconf => ../../dbotconf
)
