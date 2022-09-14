module go.opentelemetry.io/build-tools/semconvgen

go 1.18

require (
	github.com/spf13/pflag v1.0.5
	go.opentelemetry.io/build-tools v0.0.0-20220321164008-b8e03aff061a
	golang.org/x/mod v0.5.1
	golang.org/x/text v0.3.7
)

require golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898 // indirect

replace go.opentelemetry.io/build-tools => ../
