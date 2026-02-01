# Directory Structure

## General Structure
The project is designed to comply with the standard Go project layout as described in the [Golang Standards project layout repository](https://github.com/golang-standards/project-layout). However, there are some situations where the project deviates from the guidelines presented within that repository. In the case that a deviation is made, it is documented below.

## Deviations from Standard Layout

### Directory Nesting
Nested directories are commonly used to group related functionality together. For example, the `internal/fflogs` directory might contain all code related to interacting with the FFLogs API, including data models, service clients, and utility functions. However, there comes a point where nesting is more of a hinderance than a help.

When nesting directories, the depth of the directory structure should not exceed three levels. For example, `internal/fflogs/client` is acceptable, but `internal/fflogs/client/http` is not. If a directory structure exceeds three levels of nesting, it should be flattened by either merging directories or moving code to a more appropriate location.

In addition, nested directories should only be created when the code within them is closely related to the parent directory. As an example, the bot config loading functionality is located within the `internal/clearingway/config` directory. This is because no other code outside the `internal/clearingway` directory requires access to the config loading functionality, and it is closely related to the core bot functionality. This is in contrast to things like environment variable management, which is located within the `internal/env` directory, as it may be accessed by other modules within the project.

Directories should only be created when there is a clear need to group related functionality together. If a directory contains only a single file or a small number of files that are not closely related, it may be better to move those files to another location within the project. The exception to this is when the parent directory is an organisational directory such as `cmd/`, `internal/`, or `pkg/`.