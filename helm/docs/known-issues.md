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

## Known Issues for FiftyOne Enterprise v2.26.0 and Above

### Invitations After Upgrading From Legacy Authentication Mode

FiftyOne Enterprise v2.26.0 removes `legacy` authentication mode and moves
existing `legacy` deployments to `internal` authentication mode on startup.
In `legacy` mode, Auth0 sent invitations to join your organization.
After the upgrade, FiftyOne Enterprise sends them instead, and only if
invitations are enabled for your organization in the CAS SuperAdmin UI.
Check that setting after upgrading. To send invitations as emails, you must
also configure an SMTP connection in the CAS SuperAdmin UI.
