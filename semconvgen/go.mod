module go.opentelemetry.io/build-tools/semconvgen

go 1.15

require (
	github.com/spf13/pflag v1.0.5
	go.opentelemetry.io/build-tools v0.0.0-20210716171533-9af21479e912
	golang.org/x/mod v0.4.2
)

// This should be removed once the first commit lands in this module
replace go.opentelemetry.io/build-tools => ../
