# Documentation checker

Tool that checks the components enabled in OpenTelemetry core and contrib repos have 
the proper documentations.

Usage:
```
checkdoc --project-path path/to/project \
	--component-rel-path service/defaultcomponents/defaults.go \
	--module-name go.opentelemetry.io/collector
```