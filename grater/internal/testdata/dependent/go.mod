module go.opentelemetry.io/build-tools/grater/internal/testdata/dependent

go 1.26.0

require (
	github.com/stretchr/testify v1.12.1
	go.opentelemetry.io/build-tools/grater/internal/testdata/module v0.30.0
)

require go.yaml.in/yaml/v4 v4.0.0 // indirect

replace go.opentelemetry.io/build-tools/grater/internal/testdata/module => ../modulePass
