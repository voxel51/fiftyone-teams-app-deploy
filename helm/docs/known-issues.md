# Known Issues

## Known Issues for FiftyOne Teams v1.6.0 and Above

### Invitations Disabled for Internal Authentication Mode

FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS), which
includes both
[`legacy` authentication mode][legacy-auth-mode]
and
[`internal` authentication mode][internal-auth-mode].

Inviting users to join your FiftyOne Teams instance is not currently supported
when `FIFTYONE_AUTH_MODE` is set to `internal`.

We publish the following FiftyOne Teams private images to Docker Hub:

- `voxel51/fiftyone-app`
- `voxel51/fiftyone-app-gpt`
- `voxel51/fiftyone-app-torch`
- `voxel51/fiftyone-teams-api`
- `voxel51/fiftyone-teams-app`
- `voxel51/fiftyone-teams-cas`

For Docker Hub credentials, please contact your Voxel51 support team.

[internal-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#internal-mode
[legacy-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#legacy-mode
