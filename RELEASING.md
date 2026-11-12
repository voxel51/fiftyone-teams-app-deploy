# Releasing the FiftyOne Enterprise Helm chart

> [!NOTE]
> These steps are to be performed by authorized Voxel51 engineers.

This repository is trunk based with the `main` branch as the trunk.
Follow these rules for a release:

- Merge changes to `main` first.
- Never commit directly to a release branch.
  Cherry-pick from `main` instead.
  The only exceptions are version bumps and release-only fixes.
  See [Cherry-picking fixes](#cherry-picking-fixes).
- Set the release version on the release branch, not on `main`.
- Keep `main` on the latest released version.
  After each release,
  the Release Charts workflow opens a PR that moves `main` forward.
- Patch only the latest minor release.
- Use the `fiftyone-teams-app-X.Y.Z` tag
  to find the code for a past release.

## Minor release (X.Y.0)

1. Create the branch `release/vX.Y.0` from `main`:

    ```shell
    git checkout main
    git pull
    git checkout -b release/vX.Y.0
    git push --set-upstream origin release/vX.Y.0
    ```

1. Bump the version on the release branch to `X.Y.0`.
   Follow [Bumping the version](#bumping-the-version)
   with version `X.Y.0`.
1. Test the release candidate charts.
   See [Release candidates](#release-candidates).
1. Merge each fix to `main`.
   Cherry-pick the fix to the release branch.
   See [Cherry-picking fixes](#cherry-picking-fixes).
1. On release day,
   follow the steps in [Release day](#release-day)
   with version `X.Y.0`.

## Patch release (X.Y.Z)

1. Merge the fixes to `main`.
1. Create the branch `release/vX.Y.Z`
   from the tag of the release you are patching, `fiftyone-teams-app-X.Y.Z-1`:

    ```shell
    git fetch origin --tags
    git checkout -b release/vX.Y.Z fiftyone-teams-app-X.Y.Z-1
    git push --set-upstream origin release/vX.Y.Z
    ```

1. Bump the version on the release branch to `X.Y.Z`.
   Follow [Bumping the version](#bumping-the-version)
   with version `X.Y.Z`.
1. Cherry-pick the fixes to the release branch.
   See [Cherry-picking fixes](#cherry-picking-fixes).
1. On release day,
   follow the steps in [Release day](#release-day)
   with version `X.Y.Z`.

## Bumping the version

Use these steps for both minor and patch releases.
`X.Y.Z` is the version of the release branch.
Merge the version bump before you cherry-pick any fixes.
The `Validate Chart Version` check fails
until `Chart.yaml` matches the branch name.

1. Create a branch from the release branch:

    ```shell
    git fetch origin
    git checkout -b version-bump/vX.Y.Z origin/release/vX.Y.Z
    ```

1. Replace every reference to the previous version with `X.Y.Z`.
   The script updates the chart, Docker Compose files, `skaffold.yaml`, and docs.
   It doesn't change the integration test fixtures,
   or text between `<!-- bump-version: off -->` and `<!-- bump-version: on -->`.
   Wrap upgrade notes and rollback steps for a specific release in those markers,
   so later releases keep the versions they name.

    ```shell
    ./utils/bump-version.sh --to X.Y.Z
    ```

1. Review the changes.
   Confirm that only version references changed:

    ```shell
    git diff
    ```

1. Commit the changes and push the branch:

    ```shell
    git add --update
    git commit -m "chore(release): bump version to vX.Y.Z"
    git push --set-upstream origin version-bump/vX.Y.Z
    ```

    > [!NOTE]
    > If a pre-commit hook changes files and stops the commit,
    > run `git add --update` and `git commit` again.

1. Open a PR to the release branch with the `version-bump` label:

    ```shell
    gh pr create \
      --base release/vX.Y.Z \
      --head version-bump/vX.Y.Z \
      --title "chore(release): bump version to vX.Y.Z" \
      --body "Bumps the versions on release/vX.Y.Z to X.Y.Z." \
      --label version-bump
    ```

1. Get an approval,
   wait for the checks to pass,
   and merge the PR.

## Release day

Use these steps for both minor and patch releases.
`X.Y.Z` is the version you are releasing.

1. The release coordinator runs the
   [public release workflow](https://github.com/voxel51/cloud-build-and-deploy/blob/main/.github/workflows/public-release-build-and-deploy.yml)
   as part of the full FiftyOne Enterprise release.
   See the
   [release process](https://github.com/voxel51/infra-docs/blob/main/docs/release-processes/releasing.md).
   The workflow tags the tip of the release branch `fiftyone-teams-app-X.Y.Z`.
1. Find the
   [Release Charts workflow](./.github/workflows/release.yml)
   run that the tag push started,
   and watch it:

    ```shell
    gh run list \
      --workflow release.yml \
      --branch fiftyone-teams-app-X.Y.Z
    gh run watch <RUN_ID> --exit-status
    ```

    If it fails,
    fix the cause and re-run the failed jobs until the run succeeds:

    ```shell
    gh run rerun <RUN_ID> --failed
    ```

    > [!NOTE]
    > The workflow:
    >
    > - checks that the tag matches `Chart.yaml`
    > - checks that you can pull the released images
    > - publishes the chart and its README to
    >   [helm.fiftyone.ai](https://helm.fiftyone.ai)
    > - starts the internal chart build
    > - opens a PR that moves `main` to the release.
    >   It bumps the versions on `main` to `X.Y.Z`
    >   and pins the integration fixtures to the released images.

1. Confirm that the public Helm repository lists version `X.Y.Z`:

    ```shell
    helm repo add voxel51 https://helm.fiftyone.ai
    helm repo update voxel51
    helm search repo voxel51/fiftyone-teams-app --version X.Y.Z
    ```

1. Find the `chore(release): update main to vX.Y.Z` PR
   that the workflow opened on `main`:

    ```shell
    gh pr list \
      --base main \
      --head release-sync/vX.Y.Z
    ```

1. Review the PR,
   get an approval,
   and merge it.

## Cherry-picking fixes

Get every fix onto a release branch by cherry-picking it from `main`.
Pick one of these methods:

- Label the PR (preferred)
  - Before you merge the `main` PR,
    label it `cherry-pick-to-release`.
    When it merges,
      a bot opens the release-branch PR for you.
- Run the workflow
  - If the `main` PR already merged,
    run the
    [Cherry-Pick to Release workflow](./.github/workflows/cherry-pick.yml).
     Enter the PR number and the release branch.
- Cherry-pick by hand
  - Use this when the bot reports a conflict:

    ```shell
    git fetch origin
    git checkout -b cherry-pick/vX.Y.Z/pr-<PR_NUMBER> origin/release/vX.Y.Z
    git cherry-pick -x <COMMIT_SHA>
    ```

    If `<COMMIT_SHA>` is a merge commit,
    add `-m 1` to the `git cherry-pick` command.
    Resolve any conflicts,
    then run `git cherry-pick --continue`.
    Push the branch and open a PR to the release branch:

    ```shell
    git push --set-upstream origin cherry-pick/vX.Y.Z/pr-<PR_NUMBER>
    gh pr create \
      --base release/vX.Y.Z \
      --head cherry-pick/vX.Y.Z/pr-<PR_NUMBER> \
      --title "[release/vX.Y.Z] <PR_TITLE>" \
      --body "Cherry-pick of #<PR_NUMBER>."
    ```

  - Always use the `-x` flag.
    It records the source commit,
    which the `Release Gate` check uses to confirm the commit is on `main`.
    Rebase cherry-pick branches.
    Don't merge into them.
    The gate rejects merge commits.

The `Release Gate` check blocks any other PR to a release branch.
Use a label only in these cases:

- `version-bump`:
  the PR only bumps the versions on a release branch.
- `release-only-fix`:
  the fix applies only to the release branch
  and can't land on `main` first.

## Release candidates

Every push to a release branch publishes an `X.Y.Z-rc-<sha>` chart
to the internal registry.
A new release branch skips this step
until its version-bump PR merges.
It also starts the internal release-candidate deploy
(which pushes its integration-fixture pins back to the branch).
Test the release candidate before release day.

## Publishing Manually

The public release workflow creates the tag.
Only push a tag manually when the pipeline can't finish a release.
The tag push publishes the public chart.
