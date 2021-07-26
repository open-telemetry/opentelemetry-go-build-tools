# Go Module Releaser

This Go Cobra app adds versioning support for repos containing multiple Go
Modules. Specifically, this app allows for repo maintainers to specify versions
for sets of Go Modules, so that different sets may be versioned separately.

## Specify Module Sets and Versions

To specify sets of modules whose versions will be incremented in lockstep, the
`versions.yaml` file must be manually edited between versions. When creating a
`versions.yaml` file...

* For each module set, give it a name and specify its version and modules it
  includes.
* Specify the Go module's import path (rather than its file path)
* Specify modules which should be excluded from versioning
* Ensure versions are consistent with the semver versioning requirements.
* Update version numbers or module groupings as needed for new releases.

An example versioning file is given in [the versions-example.yaml
file](./docs/versions-example.yaml).

## Creating the app binary

TODO: switch to automatically pulling newest version of `releaser` app binary.

This section describes the current process used for building the Cobra app
binary, but the process will be updated soon to automatically fetch the most
recent version of the app.

To build the binary, simply use the following commands:

```sh
# from the opentelemetry-go-build-tools repo root
cd releaser
go build -o releaser main.go
```

## Verify Module Versioning

Once changes have been completed to the `versions.yaml` file, and the `releaser`
file is within the target repo, verify that the versioning is correct by running
the `verify` subcommand:

```sh
# within the target repo
./releaser verify
```

* The script is called with the following parameters:
  * **versioning-file (optional):** Path to versioning file that contains
    definitions of all module sets. If unspecified, will default to
    \<RepoRoot\>/versions.yaml.
* The following verifications are performed:
  * `verifyAllModulesInSet` checks that every module (as defined by a `go.mod`
      file) is contained in exactly one module set.
  * `verifyVersions` checks that module set version conform to semver semantics
      and checks that no more than one module set exists for any given non-zero
      major version.
  * `verifyDependencies` checks if any stable modules depend on unstable
    modules.
    * The stability of a given module is defined by its version in the
      `version.yaml` file (`v1` and above is stable, `v0` is unstable).
    * A dependency is defined by the "require" section of the module's `go.mod`
      file (in the current branch).
    * A warning will be printed for each dependency of a stable module on an
      unstable module.

## Prepare a pre-release commit

Update `go.mod` for all modules to depend on the specified module set's new
release.

1. Run the pre-release script. It creates a branch
   `pre_release_<module set name>_<new version>` that will contain all release changes.

    ```sh
    ./releaser prerelease --module-set-name <name>
    ```

    * The script is called with the following parameters:
        * **module-set-name (required):** Name of module set whose version is
          being changed. Must be listed in the module set versioning YAML.
        * **versioning-file (optional):** Path to versioning file that contains
          definitions of all module sets. If unspecified will default to
          (RepoRoot)/versions.yaml.
        * **from-existing-branch (optional):** Name of existing branch from
          which to base the pre-release branch. If unspecified, defaults to
          current branch.
        * **skip-make (boolean flag):** Specify this flag to skip the 'make
          lint' and 'make ci' steps. To be used for debugging purposes. Should
          not be skipped during actual release.

2. Verify the changes.

    ```sh
    git diff main
    ```

   This should have changed the version for all modules listed in `go.mod` files
   to be `<new version>`.

3. Update the repo's Changelog to move new changes to the new version's section.

4. Push the changes to upstream and create a Pull Request on GitHub. Be sure to
   include the curated changes from the Changelog in the description.

## Tag the new release commit

Once the Pull Request with all the version changes has been approved and merged,
it is time to tag the merged commit.

Make sure that you have pulled all changes from the upstream remote so that the
commit will be found in your currently checked out branch.

1. Run the tag subcommand using the `<commit-hash>` of the commit on the main
   branch for the merged Pull Request. In other words, once a Pull Request has
   been merged, the commit hash of the single commit containing all merged
   pre-release changes should be used. The tag can be given as its abbreviation,
   in which case the script will attempt to find and print the full SHA1 hash.

    ```sh
    ./releaser tag --module-set-name <name> --commit-hash <hash>
    ```

2. Push tags for ALL updated modules to the upstream remote (not your fork),
   which are printed by the `tag` script. This step must be done manually,
   specifying one tag at a time to push.

    ```sh
    git push upstream <new tag 1>
    git push upstream <new tag 2>
    ...
    ```

In the case that you made a mistake in creating Git Tags (e.g. you used the
wrong commit hash), you can run the following command to delete all of a
specified module set's tags for the version specified in `versions.yaml`. This
command will not allow you to delete tags that already exist in the upstream
remote.

```sh
./releaser tag --module-set-name <name> --delete-module-set-tags
```

## Release

Finally create a Release for the new `<new tag>` on GitHub. The release body
should include all the release notes from the Changelog for this release.

The standard releasing process for the repo should then be followed.
