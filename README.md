Automation tool used for the job processing.

Supported jobs:

- Build and deploy a package (client or server).
- Build and deploy a release (client or server).
- Build and deploy an SDK.

Process:

- Each 30 seconds request a new job.
- Parse and validate a job metadata.
- Run job processing (processClientRelease, processServerRelease, processSdkRelease, processClientPackage, processServerPackage,
  processClientLauncher).
- Upload release/package files to the APIv2.

Requirements:

- Latest source build of Unreal Engine.
- Latest marketplace build of the Unreal Engine (for the SDK).
- Project source code.
- Launcher source code.

Environment variables:

- API_URL - URL of the API, e.g. "https://test.api2.veverse.com/v2"
- API_EMAIL - email of the builder user account
- API_PASSWORD - password of the builder user account
- EDITOR_PATH - path to the editor, e.g. "X:/UEV/UnrealEngine/Engine/Binaries/Win64/UnrealEditor-Cmd.exe"
- PLATFORMS - platforms supported by the builder, e.g. "Win64,Linux";
- PROJECT_DIR - path to the project directory where the Metaverse.uproject is located, e.g. "X:/UEV/UnrealEngine/Metaverse"
- PROJECT_NAME - project name, e.g. "Metaverse"
- UAT_PATH - path to the Unreal Automation Tool, e.g. "X:/UEV/UnrealEngine/Engine/Binaries/DotNET/AutomationTool/AutomationTool.exe"
