# FiftyOne Teams v2.1.0 SDK Upgrade and Database Migration

## What is happening?

With FiftyOne Teams v2.1.0, new `created_at` and `last_modified_at`
[built-in
fields][1]
are being added to all FiftyOne `Sample` objects. These fields will
enable tracking changes and comparing datasets across time. The new
fields are added to your FiftyOne Teams datasets when they are
[migrated][2]
to the new Database version (Database version 1.0.0), or
when they are edited by the v2.1.0 SDK.

[1]: https://docs.voxel51.com/user_guide/using_datasets.html#default-sample-fields
[2]: https://docs.voxel51.com/teams/migrations.html#upgrading-your-deployment

***Upgrading to v2.1.0 requires a high degree of care. Please read this document
carefully***!

## Upgrading the SDK installs for your Team

When upgrading to SDK version v2.1.0, **all SDK installations on your team
must upgrade in lockstep**.  If one SDK is upgraded while others are not,
any datasets edited by the upgraded SDK will not be accessible to the older SDKs.

This includes **all SDK instances** that might connect to your deployment,
including automated pipelines and runners. It is simplest to upgrade all SDKs
on your team at the same time.

You should upgrade all SDKs on your team **before considering a database migration**.
After migration, any older SDK versions will not be able to connect to the database
or load datasets.

## Migrating your Database -- care and patience required

After you upgrade your SDKs, it is logical to migrate your database and datasets.
This process will create and set the new `created_at` and `last_modified_at`
fields on your datasets.

Since this migration is adding new fields and data to every Sample
(and Frame) of your datasets, the migration process is long-running and
compute-intensive. The disk usage of your MongoDB will increase,
frequently with a significant transient spike before the storage
engine settles into steady state. If this transient spike exceeds the
amount of storage you have provisioned for MongoDB, this could crash
your database!

For these reasons ***special care should be taken during the migration.***

## Recommended migration paths

**Important!** Teams with large databases, or with video datasets
  containing many Frames, may initially prefer Path B to test out the
  migration on an initial/small set of datasets and to probe the
  potential increase in MongoDB disk usage.

***Path A: Migrate the entire database at once***

For teams that have smaller
databases (say, less than 100GB on disk), it may be simplest to
perform a full database migration all at once:

1. Make sure that all users/runners have upgraded their SDKs before
performing the migration!
2. Increase the storage capacity of your MongoDB by 50% to buffer
against a transient spike in disk usage during migration.
3. Schedule/perform the migration during a time of low
utilization. During migration, CPU and memory utilization will also
spike which will adversely affect user operations.
4. Budget a significant amount of time, say up to 2 hours, for the migration.
5. Utilize the `fiftyone migrate –-all` command (Option 2 below)
6. After the migration, you may investigate the storage usage of your MongoDB.
Typically, we have found that the storage consumed comes down from a transient
peak.

***Path B: Migrate datasets individually***

For teams that prefer a more piecewise migration, datasets can be individually
migrated. This option may be useful for instance to test out the
migration to get a sense of the running time and increased MongoDB
storage requirements for your typical datasets.

1. Make sure that all users/runners have upgraded their SDKs!
2. Increase the storage capacity of your MongoDB to account for the
increased disk usage of your Datasets. Monitor this usage over time as
datasets are migrated.
3. Follow Option 3 below to migrate specific datasets via the FiftyOne CLI.
4. After a time, you can always decide to finish the migration in bulk as in
Path A.

## Migration commands

FiftyOne dataset/database migrations are done using the `fiftyone
migrate` CLI command. As [described in the
doc](https://docs.voxel51.com/teams/migrations.html#upgrading-your-deployment),
there are several options. We list the most important cases here:

```bash
export FIFTYONE_DATABASE_ADMIN=true

# Option 2: migrate the database and all datasets
fiftyone migrate --all

# Option 3: migrate the database and a specified dataset*
fiftyone migrate --dataset-name <DATASET_NAME>
```

Remember, once a migration is performed, SDK versions prior to v2.1.0 **will not
be able to connect to the database or load dataset(s).** You should
make sure that all users/runners upgrade their SDKs to v2.1.0 **in lockstep, prior
to performing** any migrations.
