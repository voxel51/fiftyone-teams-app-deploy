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

# Custom MongoDB Permissions

Generally, we recommend that FiftyOne Teams connect to MongoDB [with root access][1].

[1]: https://docs.voxel51.com/user_guide/config.html?highlight=mongodb%20uri#configuring-a-mongodb-connection

In some cases, more limited connection permissions are desired. The following
set of custom permissions may be used as of FiftyOne Teams v2.6.0:

* `clusterMonitor@admin`
* `read@admin`
* `readWrite@cas`
* `admin@fiftyone`
* `readWrite@fiftyone`
