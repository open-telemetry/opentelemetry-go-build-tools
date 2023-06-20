# Release Process

## Pre-Release

1. Create a new release branch. I.e. `git checkout -b release-X.X.X main`.

1. Update the version in [`versions.yaml`](versions.yaml) and commit the change.

1. Run the pre-release steps which updates `go.mod` and `version.go` files
   in modules for the new release.

    ```sh
    make multimod-prerelease
    ```

1. Merge the branch created by `multimod` into your release branch.

1. Update `go.mod` and `go.sum` files.

    ```sh
    make tidy
    ```

1. Update [CHANGELOG.md](CHANGELOG.md) with new the new release.

    ```sh
    make chlog-update VERSION=<vX.X.X>
    ```

1. Push the changes and create a Pull Request on GitHub.

## Tag

Once the Pull Request with all the version changes has been approved
and merged it is time to tag the merged commit.

***IMPORTANT***: It is critical you use the crrect commit
that was pushed in the Pre-Release step!
Failure to do so will leave things in a broken state.

***IMPORTANT***:
[There is currently no way to remove an incorrectly tagged version of a Go module](https://github.com/golang/go/issues/34189).
It is critical you make sure the version you push upstream is correct.
[Failure to do so will lead to minor emergencies and tough to work around](https://github.com/open-telemetry/opentelemetry-go/issues/331).

```sh
make push-tags REMOTE=<upstream> COMMIT=<hash>
```

## Release

Create a Release for the new `<new tag>` on GitHub.
The release body should include all the release notes
for this release taken from [CHANGELOG.md](CHANGELOG.md).
