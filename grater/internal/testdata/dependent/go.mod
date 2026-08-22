module go.opentelemetry.io/build-tools/grater/internal/testdata/dependent

go 1.25.0

require (
	github.com/stretchr/testify v1.12.1
	go.opentelemetry.io/build-tools/grater/internal/testdata/module v0.30.0
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect

replace go.opentelemetry.io/build-tools/grater/internal/testdata/module => ../modulePass
