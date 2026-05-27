# Grater

Tool that checks for regressions in our downstream dependents due to our changes.

Usage:

```sh
    # Adds one or more dependents to be tested.
    grater add foo/bar@v1.0.0 bar/foo --file <file_path>
```

```sh
    # Adds one or more replacements.
    grater replace foo/bar@v1.0.0 bar/foo@v1.0.1 --file <file_path>
```

```sh
    # Finds dependents for a remote module.
    grater replace foo/bar@v1.0.0
```
