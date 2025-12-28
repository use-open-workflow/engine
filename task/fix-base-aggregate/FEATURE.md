# Goal

The base aggregate stores `createdAt` and `updatedAt` date time

# Background

The `createdAt` and `updatedAt` date time fields are only stored in database. However, these two value are not being deserialized to be part of aggregate

# Problem

Missing `createdAt` and `updatedAt` date time fields to base aggregate

# Solution

Add `createdAt` and `updatedAt` date time fields to base aggregate

# Proposal

The `createdAt` and `updatedAt` fields should be handled by the application itself. Do not rely on the SQL function.

# Acceptance Criteria

1. The API DTO contains `createdAt` and `updatedAt`
2. The `createdAt` and `updatedAt` are controlled by application not SQL function
3. The `createdAt` and `updatedAt` must be date time and in UTC timezone
