# githubgen

This executable is used to generate `.github/CODEOWNERS` files.

It reads status metadata from `metadata.yaml` files located throughout the
repository.

It checks that codeowners are known members of the OpenTelemetry organization.

## Usage

```shell
$> ./githubgen
```

The equivalent of:

```shell
$> githubgen --skipgithub --folder . --github-org "open-telemetry" \
  --default-codeowner open-telemetry/opentelemetry-collector-approvers \
  --allowlist cmd/githubgen/allowlist.txt
```

## Checking codeowners against OpenTelemetry membership via GitHub API

To authenticate, set the environment variable `GITHUB_TOKEN` to a PAT token.
If a PAT is not available you can use the `--skipgithub` flag to avoid checking
for membership in the GitHub organization.

Additionally, the GitHub organization name needs to be specified using
`--github-org <your-org>`, along with `--default-codeowner` to be used as the
default owner for general files.

For each codeowner, the script will check if the user is registered as a member
of the OpenTelemetry organization.

If any codeowner is missing, it will stop and print names of missing codeowners.

These can be added to allowlist.txt as a workaround.

If a codeowner is present in `allowlist.txt` and also a member of the
OpenTelemetry organization, the script will error out.
