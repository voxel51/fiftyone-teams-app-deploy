# Releasing

> [!NOTE]
> These steps are to be performed by authorized Voxel51 engineers.

`main` is the trunk: every PR merges to `main`, and nothing originates on a
release branch. `main` carries the version in flight, so its `Chart.yaml`,
image tags and docs name the release being prepared; the previous release
lives on its `fiftyone-teams-app-X.Y.Z` tag.

## Minor release (X.Y.0)

1. Open a PR to `main` replacing every reference to the previous version
   with `X.Y.0` (`helm/fiftyone-teams-app/Chart.yaml` `version` and
   `appVersion`, the compose files, `skaffold.yaml`, the docs) and merge it.
1. Cut `release/vX.Y.0` from `main`.

   Every push to the release branch publishes an `X.Y.0-rc-<sha>` chart to
   the internal registry and dispatches the internal release-candidate
   deploy, which pushes its integration-fixture pins back to the branch.
1. Land fixes on `main` first. Label the `main` PR `cherry-pick-to-release`
   and the release-branch PR opens itself when it merges, or run the
   `Cherry-Pick to Release` workflow for an already merged PR. The
   `Release Gate` check admits only cherry-picks of `main` commits and PRs
   labeled `version-bump` or `release-only-fix`.
1. Release day: the
   [public release workflow](https://github.com/voxel51/cloud-build-and-deploy/blob/main/.github/workflows/public-release-build-and-deploy.yml)
   tags the release branch tip `fiftyone-teams-app-X.Y.0`. The tag triggers
   [Release Charts](./.github/workflows/release.yml), which checks the tag
   against `Chart.yaml`, verifies the released images are pullable,
   publishes the chart and its README to
   [helm.fiftyone.ai](https://helm.fiftyone.ai), dispatches the internal
   chart build, and opens a PR pinning the integration fixtures on `main`
   to the release. Merge that PR.

## Patch release (X.Y.Z)

1. Cut `release/vX.Y.Z` from the `fiftyone-teams-app-X.Y.Z-1` tag.
1. Bump the versions to `X.Y.Z`: on `main` first, cherry-picked to the
   branch, while `main` is still on `X.Y.Z-1`; otherwise as a PR to the
   branch labeled `version-bump`. Merge it before any other cherry-pick,
   since `Validate Chart Version` requires `Chart.yaml` to match the branch
   name.
1. Cherry-pick the fixes and release as above.

## Publishing by hand

Pushing a `fiftyone-teams-app-X.Y.Z` tag publishes the public chart; push
one only for a release the pipeline could not complete.
