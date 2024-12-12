# githubgen

This executable is used to generate `.github/CODEOWNERS` and
`.github/ALLOWLIST` files.

It reads status metadata from `metadata.yaml` files located throughout the
repository.

It checks that codeowners are known members of the OpenTelemetry organization.

## Usage

```shell
$> ./githubgen
```

The equivalent of:

```shell
$> GITHUB_TOKEN=<mypattoken> githubgen --folder . [--allowlist cmd/githubgen/allowlist.txt] 
```

## Checking codeowners against OpenTelemetry membership via GitHub API

To authenticate, set the environment variable `GITHUB_TOKEN` to a PAT token.
If a PAT is not available you can use the `--skipgithub` flag to avoid checking
for membership in the GitHub organization.

For each codeowner, the script will check if the user is registered as a member
of the OpenTelemetry organization.

If any codeowner is missing, it will stop and print names of missing codeowners.

These can be added to allowlist.txt as a workaround.

If a codeowner is present in allowlist.txt and also a member of the
OpenTelemetry organization, the script will error out.
