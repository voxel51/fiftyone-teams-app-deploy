# FiftyOne Agent

The FiftyOne Agent is an AI-powered assistant built into the FiftyOne Enterprise
App. It lets you work with your datasets using natural language. You can import
data, run model inference, find duplicates, evaluate predictions, and more, all
from a conversational interface.

[![FiftyOne Agent Demo](https://cdn.voxel51.com/voxel-agent/enterprise/voxel_agent_demo_1.webp)](https://your-video-url.mp4)

## Central Deployment

### Enable the Agent

Contact your Customer Success representative to enable the FiftyOne Agent for
your deployment via feature flag.

Once enabled, open any dataset in the FiftyOne Enterprise App. You will see a
new **Agent** button in the upper-right corner of the App.

![Agent button location](https://cdn.voxel51.com/voxel-agent/enterprise/location_agent.webp)

### Configure model providers

The first time you open the Agent, you will be prompted to configure a model
provider. The Agent supports over 100 providers, including Anthropic, OpenAI,
Google, and more.

![Agent settings](https://cdn.voxel51.com/voxel-agent/enterprise/agent_settings.webp)

To add a provider, fill in the following fields:

- **Name**: a label for this provider configuration
- **Provider**: select from the list of supported providers
- **Endpoint** (optional): use this if your model is hosted at a custom URL
- **API key**: your provider's API key
- **Models**: select one or more models to make available
- **Default**: mark this provider as the default

![Provider details](https://cdn.voxel51.com/voxel-agent/enterprise/provider_more_details.webp)

Click **Test connection** to verify your credentials before saving.

### Using the Agent

Once a provider is configured, type any task in plain language and the Agent
will execute it against your dataset.

![Agent chat](https://cdn.voxel51.com/voxel-agent/enterprise/agent_chat.webp)

Some examples of what you can ask:

- *"Find and remove duplicate images from this dataset"*
- *"Run object detection and show me low-confidence predictions"*
- *"Export this dataset to COCO format"*

To start a new conversation, click the **+** button. To return to a previous
conversation, click **History**.

### Skills

The Agent ships with a set of built-in skills that cover the most common
computer vision workflows. Skills are structured instructions that tell the
agent exactly how to perform a task, step by step.

![Skills](https://cdn.voxel51.com/voxel-agent/enterprise/skills.webp)

## Local Development

Connect your local FiftyOne installation to your development environment. This
lets you point at your local plugin folder instead of the shared one, so your
changes don't affect other users.

### Prerequisites

You need the agent plugin `.zip` file.

> **Note:** This is a temporary step. After the next release the Agent will be
> available as a built-in feature and this step will no longer be needed.

To install the plugin locally, find your plugins directory:

```bash
python -c "import fiftyone; print(fiftyone.config.plugins_dir)"
```

Extract the `.zip` file into that directory.

### One-time setup

**Step 1: Create a conda environment**

```bash
conda create --name teams python=3.11
conda activate teams
```

**Step 2: Install FiftyOne Teams**

Get the exact install command from your **API keys settings page** in the App.

![api_settings](https://cdn.voxel51.com/voxel-agent/enterprise/api_settings.webp)


Then run it:

```bash
pip install --index-url "https://<token>@pypi.dev.fiftyone.ai/simple/" \
  fiftyone==<version>
```

**Step 3: Install agent dependencies**

```bash
pip install litellm==1.83.0 fiftyone-mcp-server==0.1.12
```

**Step 4: Generate an API key**

Go to your environment's settings page and create a new API key.

**Step 5: Export credentials**

![export_credentials](https://cdn.voxel51.com/voxel-agent/enterprise/export_credentials.webp)


```bash
export FIFTYONE_API_URI=https://<your-env>-api.fiftyone.ai
export FIFTYONE_API_KEY=<your-api-key>
```

**Step 6: Configure a model provider**

In local mode, provider API keys are read from environment variables rather
than the UI. Open the Agent using the button in the upper-right corner of the
App:

![Agent button location](https://cdn.voxel51.com/voxel-agent/enterprise/location_agent.webp)

The Agent will display the exact variable name it needs for your provider.
Export it:

```bash
export VOXEL_AGENT_CUSTOM_PROVIDER_1=<your-key>
```

### Run

```bash
fiftyone app debug
```

Once the App is open, click the **Agent** button in the upper-right corner to
start a conversation.

![Agent chat](https://cdn.voxel51.com/voxel-agent/enterprise/agent_chat.webp)
