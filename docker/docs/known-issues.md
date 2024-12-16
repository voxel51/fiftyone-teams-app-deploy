# Known Issues

## Known Issues for FiftyOne Teams v1.6.0 and Above

### Invitations Disabled for Internal Authentication Mode

FiftyOne Teams v1.6 introduces the Central Authentication Service (CAS), which
includes both
[`legacy` authentication mode][legacy-auth-mode]
and
[`internal` authentication mode][internal-auth-mode].

Prior to v2.2.0, inviting users to join your FiftyOne Teams instance was not supported
when `FIFTYONE_AUTH_MODE` is set to `internal`.
Starting in v2.2.0+, you can enable invitations for your organization through the
CAS SuperAdmin UI. To enable sending invitations as emails, you must also
configure an SMTP connection.

[internal-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#internal-mode
[legacy-auth-mode]: https://docs.voxel51.com/teams/pluggable_auth.html#legacy-mode
