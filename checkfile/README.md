# File checker

Tool that checks the components enabled in OpenTelemetry core and contrib repos
contain the file provided in argument --file-name.

Usage:

```sh
checkfile --project-path path/to/project \
         --component-rel-path service/defaultcomponents/defaults.go \
         --module-name go.opentelemetry.io/collector
         --file-name README.md
```
