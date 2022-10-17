
## developer

### running

Run `make run <args>`

### building

#### go releaser

Use `goreleaser` to build binaries, linux packages, and docker containers for all platforms without publishing.

Use `make release-snapshot` to build everything in `./dist/` and create local docker images.

#### go executable

Create a single platform binary in `./dist/` with `make build-go-linux-amd64` or `make build-go-linux-arm64`
or `make build-go-local` to auto-detect the host platform.

#### docker images

Use `make docker-linux-amd64` or `make docker-linux-arm64` to build with local docker images.

## releasing

### app

#### install

```shell
brew install goreleaser
```

#### release

To use `goreleaser` set a new git tag:

```shell
git tag -a v0.1.0 -m "New release"
git push origin v0.1.0
```

Set `GITHUB_TOKEN` in the environment with `repo` access.

Run `make release`.

Publish the draft release on the GitHub [Releases](https://github.com/superleaguegaming/archie/releases) page.

### helm chart

#### install

Use `make helm-install` to add the nexus helm plugin and repo.

#### release

Increment the chart `version` number in `helm/archie/Chart.yaml` and update the `appVersion` to the latest.

Use `make helm-release` to publish the helm chart to the repo.

Commit the changes.
