# Goal

Initialized a new domain called `workflow`

# Background

User-defined workflow is the core feature in the platform. The `workflow` aggregate will provide the functionality to allow user-defined workflow.

# Problem

Integration between services is hard and required coding.

# Solution

Allow user to define `workflow`. Within the `workflow`, it will contain a list of `node-definition` and a list of `edge`. The `node-definition` is an entity that is created based on the `node-template`. The `edge` is an entity that connects `node-definition` to form the `workflow`.

# Proposal

Create APIs, services, and repositories for these 3 aggregates/entities:

1. `workflow` (aggregate) - contains a list of `node-definition` and a list of `edge`
2. `node-definition` (entity) - contains a reference ID to `node-template`
3. `edge` (entity) - contains the from/to `node-definition`

These aggregates/entities will stay under `workflow` domain.

# Acceptance Criteria

1. APIs to retrieve `workflow`, `node-definition` and `edge`
2. Services to perform logic for `workflow`, `node-definition` and `edge`
3. Postgres repositories to store `workflow`, `node-definition` and `edge`
