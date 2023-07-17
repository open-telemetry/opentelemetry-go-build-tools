# Documentation checker

Tool that checks the components enabled in OpenTelemetry core and contrib repos
have the proper documentations.

Usage:

```sh
checkdoc --project-path path/to/project \
         --component-rel-path service/defaultcomponents/defaults.go \
         --module-name go.opentelemetry.io/collector
```

## Deprecation note

This has been deprecated in favor of `checkfile`. Please use `checkfile` with
argument `--file-name README.md` instead.
