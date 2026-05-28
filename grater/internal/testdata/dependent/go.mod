module go.opentelemetry.io/build-tools/grater/internal/testdata/dependent

go 1.25.0

require (
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/build-tools/grater/internal/testdata/module v0.30.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace go.opentelemetry.io/build-tools/grater/internal/testdata/module => ../modulePass
