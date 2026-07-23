"""Seeds a deployment's orchestrator registrations in Mongo, so they are
versioned in the deployment's values instead of hand-created.

Runs as a helm post-install/post-upgrade hook (see
`../templates/seed-orchestrators-job.yaml`).
Helm provides the `ORCHESTRATORS` env var
containing a JSON list of orchestrators (key names under
`delegatedOperatorJobTemplates.jobs` and `.serviceOrchestrators` with
`registerOrchestrator=true`).

Upserts by instance_id: config. The fields `description`, `environment`,
and `secrets` are re-applied on every run.
The `created_at` field is only written when the document is first created.
Service orchestrators with `available_operators` are reset on every run.
Job orchestrators (that never contain `available_operators`) are never modified.
The app's Refresh action owns the discovered list for job targets.
Connects to Mongo using the deployment's existing teams secrets.
"""

import datetime
import json
import os

import pymongo

orchestrators = json.loads(os.environ["ORCHESTRATORS"])

client = pymongo.MongoClient(os.environ["FIFTYONE_DATABASE_URI"])
coll = client[os.environ["FIFTYONE_DATABASE_NAME"]]["orchestrators"]
now = datetime.datetime.now(datetime.timezone.utc)

for orc in orchestrators:
    set_fields = {
        "description": orc["description"],
        "environment": orc["environment"],
        "config": orc["config"],
        "secrets": orc.get("secrets", {}),
        "updated_at": now,
    }
    if "available_operators" in orc:
        set_fields["available_operators"] = orc["available_operators"]
    insert_fields = {
        "instance_id": orc["instance_id"],
        "created_at": now,
    }
    result = coll.update_one(
        {"instance_id": orc["instance_id"]},
        {"$set": set_fields, "$setOnInsert": insert_fields},
        upsert=True,
    )
    action = "created" if result.upserted_id else "updated"
    print(f"{action} orchestrator {orc['instance_id']}")

print("orchestrator seeding complete")
