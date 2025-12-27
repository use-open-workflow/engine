# Goal

Use the same UoW instance for throughout the request lifecycle.

# Background

The request lifecycle is using two UoW instances:

- Shared UoW: Created in the DI container
- Service-level UoW: Created when the service is invoked

# Problem

It can cause data inconsistency. For example, the first transaction commits the aggregate entity changes, but the second transaction fails to commit the domain event changes.

# Solution

The whole request lifecycle must use the same UoW instance.

# Proposal

Drop the shared UoW. Utilize factory pattern and create read/write repository factory that accept UoW instance.

# Acceptance Criteria

1. Single UoW
2. All changes within a request needs to be atomic
3. Data must always be accurate
