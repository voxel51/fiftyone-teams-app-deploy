<!-- markdownlint-disable no-inline-html line-length -->
<!-- markdownlint-disable-next-line first-line-heading -->
<div align="center">
<p align="center">

<img alt="Voxel51 Logo" src="https://user-images.githubusercontent.com/25985824/106288517-2422e000-6216-11eb-871d-26ad2e7b1e59.png" height="55px"> &nbsp;
<img alt="Voxel51 FiftyOne" src="https://user-images.githubusercontent.com/25985824/106288518-24bb7680-6216-11eb-8f10-60052c519586.png" height="50px">

</p>
</div>
<!-- markdownlint-enable no-inline-html line-length -->

---

# Known Issues

## Known Issues for FiftyOne Enterprise v1.6.0 and Above

### Invitations Disabled for Internal Authentication Mode

FiftyOne Enterprise v1.6 introduces the Central Authentication Service (CAS), which
includes both
[`legacy` authentication mode][legacy-auth-mode]
and
[`internal` authentication mode][internal-auth-mode].

Prior to v2.2.0, inviting users to join your FiftyOne Enterprise instance was
not supported when `FIFTYONE_AUTH_MODE` is set to `internal`.
After v2.2.0+, you can enable invitations for your organization through the
CAS SuperAdmin UI. To enable sending invitations as emails, you must also
configure an SMTP connection.

[internal-auth-mode]: https://docs.voxel51.com/enterprise/pluggable_auth.html#internal-mode
[legacy-auth-mode]: https://docs.voxel51.com/enterprise/pluggable_auth.html#legacy-mode
