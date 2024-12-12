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

[internal-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#internal-mode
[legacy-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#legacy-mode
