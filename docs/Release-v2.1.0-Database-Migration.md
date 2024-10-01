# FiftyOne Teams v2.1.0 Database Migration

## What is happening?

With FiftyOne Teams v2.1.0, new `created_at` and `last_modified_at`
[built-in
fields][1]
are being added to all FiftyOne `Sample` objects. These fields will
enable tracking changes and comparing datasets across time. The new
fields are added to your FiftyOne Teams datasets when they are
[migrated][2]
to the new Database version (this is Database version 1.0.0).

[1]: https://docs.voxel51.com/user_guide/using_datasets.html#default-sample-fields
[2]: https://docs.voxel51.com/teams/migrations.html#upgrading-your-deployment

## Why is this migration special?

Since this migration is adding new fields and data to every Sample
(and Frame) of your datasets, this migration is long-running and
compute-intensive. The disk usage of your MongoDB will increase,
frequently with a significant transient spike before the storage
engine settles into steady state. If this transient spike exceeds the
amount of storage you have provisioned for MongoDB, this could crash
your database!

For these reasons we ***highly recommend*** taking special care during
this migration. ***Please read this document carefully***!

## Recommended upgrade/migration paths

**Important!** Teams with large databases, or with video datasets
  containing many Frames, may initially prefer Path B to test out the
  migration on an initial/small set of datasets and to probe the
  potential increase in MongoDB disk usage.

***Path A: Migrate the entire database***

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
5. Utilize the `fiftyone migrate –all` command (Option 2 below)

***Path B: Migrate datasets on-demand***

For teams that prefer a more piecewise migration, datasets can be individually
migrated, either
lazily when they are next loaded, or via an explicit migration
command. This option may be useful for instance to test out the
migration to get a sense of the running time and increased MongoDB
storage requirements for your typical datasets.

1. Make sure that all users/runners have upgraded their SDKs!
2. Increase the storage capacity of your MongoDB to account for the
increased disk usage of your Datasets. Monitor this usage over time as
datasets are migrated.
3. Follow Option 1 or Option 3 below:
    * Option 1: Datasets are migrated when loaded in the SDK. Note,
larger datasets could take a significant time (many minutes) to
migrate at load-time.
    * Option 3: Specific datasets are explicitly
migrated via the FiftyOne CLI.  3. After a time, you can always decide
to finish the migration in bulk as in Path A.

## How do I perform the migration?

FiftyOne dataset/database migrations are done using the `fiftyone
migrate` CLI command. As [described in the
doc](https://docs.voxel51.com/teams/migrations.html#upgrading-your-deployment),
there are several options:

```bash
export FIFTYONE_DATABASE_ADMIN=true

# Option 1: update the database version only (datasets lazily migrated on load)
fiftyone migrate

# Option 2: migrate the database and all datasets
fiftyone migrate --all

# Option 3: migrate the database and a specified dataset*
fiftyone migrate --dataset-name <DATASET_NAME>
```

### Important, please note

Once a migration is performed, SDK versions prior to v2.1.0 **will not
be able to connect to the database or load dataset(s).** You should
make sure that all users/runners upgrade their SDKs to v2.1.0 **prior
to performing** any migrations.
