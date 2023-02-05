![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/switchboard-org/switchboard/ci.yaml)

# Switchboard
Switchboard is a configuration-based workflow automation tool that is built for developers. It is open-source,
highly-extensible, and configuration-first, and is designed to be human-readable while still meeting the most
essential needs of a developer.

## Key Features
 1. **Human-Readable** -> Author workflows in HCL (Hashicorp Configuration Language), a format optimized for readability and productivity.
 2. **Extensible** -> All integrations are written in their own isolated codebase called a Provider. Use the official providers or write your own.
 3. **Open-Source** -> Essential for transparency, free from vendor-lock, and guaranteed to be available.
 4. **Multi-Environment Support** -> Have as many workspaces as you want, depending on your needs.
 5. **CI/CD & Version-Control** -> Workflow automation that actually works with you and your team.
 6. **And Much More** -> Managed Authentication, Automated Trigger Registration, Effective Error and Retry Settings, just to name a few.

## Switchboard CLI
This repository specifically includes the Switchboard CLI, written in GO, and the main tool used by developers within the
Switchboard toolset ecosystem. In order to start authoring workflows, you will need to download the latest
version onto your box. You can find the latest release in the
[releases](https://github.com/switchboard-org/switchboard/releases) page.

### Primary CLI Commands
Below are the primary commands available in the CLI
 1. `switchboard init` - downloads any necessary dependencies specified in your workflow config. This must
be run first for any other command to work.
 2. `switchboard validate` - validates that your entire workflow configuration is valid.
 3. `switchboard deploy` - will first run validation, and then deploy all changes to the cloud environment, keeping
any unmodified workflows untouched. This also dynamically downloads + starts or stops + deletes providers in the cloud environment,
depending on the diff of the previous workflow state.
 4. `switchboard destroy` - terminates all workflows by deregistering any triggers (webhooks, event-listeners, etc.)
and deleting all providers on the cloud environment.

`switchboard deploy` and `switchboard destroy` both assume the cloud environment exists, is initialized, and is available
from where the CLI is running. The following commands are helpers for getting your cloud environment up and running

### Cloud CLI Commands
 1. `switchboard cloud init` - this will create and initialize the cloud environment for you. Depending on the cloud provider,
this command may not be supported (i.e. you have to setup the cloud environment on your own).
 2. `switchboard cloud up` - starts the coordinator service on a publicly exposed location. This service handles all incoming
events from integrations, coordinates all workflows, communicates with providers, and stores workflow state. It is also
primary endpoint the CLI communicates with, so it is essential this is running.
 3. `switchboard cloud down` - opposite of the above command. `switchboard destroy` must be run before this command, or it will fail.
 4. `switchboard cloud destroy` - opposite of the init command, but may not be available, depending on the cloud provider, and the
workflow configuration settings.
