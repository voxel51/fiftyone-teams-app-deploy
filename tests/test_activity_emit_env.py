"""
Regression guard for the Activity Analytics emit path.

The incident: with ``activitySettings.enabled``, the annotate/review metrics
page showed only seeded data because **no real events were being captured**.
The workflows-plugin ``emit()`` runs in the ``teams-plugins`` and
delegated-operator pods, but those pods were not given
``FIFTYONE_MQ_REDIS_URL`` — so ``fiftyone.mq`` fell back to its default
``redis://localhost:6379/0`` (unreachable in-cluster) and every event was
**silently dropped** (emit is fire-and-forget by design). A beautiful metrics
page fed only by a seed script is worthless, so this test fails loudly if any
pod that emits activity events is missing the queue wiring.

Invariant under test: **every pod that runs the workflows-plugin emit must be
able to reach the activity queue** when Activity is enabled. Concretely, the
rendered ``teams-plugins`` Deployment and every delegated-operator Deployment
must set ``FIFTYONE_MQ_REDIS_URL`` (and ``FIFTYONE_ACTIVITY_ORG_ID``, without
which the org-scoped rollups never match the reader).

Runnable standalone (``python tests/test_activity_emit_env.py``) or via pytest.
Requires the ``helm`` CLI.

| Copyright 2017-2026, Voxel51, Inc.
| `voxel51.com <https://voxel51.com/>`_
|
"""

import os
import subprocess
import sys

import yaml

CHART = os.path.join(
    os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
    "helm",
    "fiftyone-teams-app",
)

# The env vars a pod needs before its emit() can deliver to the queue. The
# Redis URL is load-bearing: without it fiftyone.mq silently uses localhost and
# drops every event. The org id is required for the rollups to be readable.
REQUIRED_EMIT_ENV = ["FIFTYONE_MQ_REDIS_URL", "FIFTYONE_ACTIVITY_ORG_ID"]

# A delegated-operator deployment we inject so the DO env path is exercised.
# Its rendered Deployment name is the values key, so we know it exactly.
_DO_PROBE_KEY = "activityemitprobe"

# Deployments whose pods run the workflows-plugin / delegated-operator emit —
# the ones the incident showed were missing the queue wiring. Matched by exact
# rendered name (DO deployment name == the injected values key).
EMIT_POD_NAMES = {"teams-plugins", _DO_PROBE_KEY}


def _render():
    """Renders the chart with Activity + an example delegated operator on."""
    out = subprocess.run(
        [
            "helm",
            "template",
            "t",
            CHART,
            "--set",
            "activitySettings.enabled=true",
            "--set",
            "pluginsSettings.enabled=true",
            # A delegated-operator deployment so the DO env path is exercised.
            "--set",
            f"delegatedOperatorDeployments.deployments.{_DO_PROBE_KEY}.name=teams-do-probe",
        ],
        capture_output=True,
        text=True,
        check=True,
    )
    return [
        d
        for d in yaml.safe_load_all(out.stdout)
        if d and d.get("kind") == "Deployment"
    ]


def _env_names(deployment):
    container = deployment["spec"]["template"]["spec"]["containers"][0]
    return {e["name"] for e in (container.get("env") or [])}


def _emit_deployments(deployments):
    """The emit-running Deployments (teams-plugins + the delegated operator)."""
    return [d for d in deployments if d["metadata"]["name"] in EMIT_POD_NAMES]


def _check():
    deployments = _render()
    emit_pods = _emit_deployments(deployments)
    found = {d["metadata"]["name"] for d in emit_pods}
    missing_pods = EMIT_POD_NAMES - found
    assert not missing_pods, (
        f"expected emit deployments {sorted(EMIT_POD_NAMES)} to render, "
        f"missing {sorted(missing_pods)}"
    )

    failures = []
    for d in emit_pods:
        name = d["metadata"]["name"]
        env = _env_names(d)
        missing = [v for v in REQUIRED_EMIT_ENV if v not in env]
        if missing:
            failures.append(f"{name} is missing {missing}")

    assert not failures, (
        "Activity emit pods cannot reach the queue (events would be silently "
        "dropped):\n  " + "\n  ".join(failures)
    )
    return [d["metadata"]["name"] for d in emit_pods]


def test_activity_emit_pods_have_queue_env():
    """teams-plugins + delegated operators must carry the emit queue env."""
    _check()


if __name__ == "__main__":
    try:
        names = _check()
    except AssertionError as exc:
        print("FAIL:", exc)
        sys.exit(1)
    print("PASS: emit env present on:", ", ".join(names))
