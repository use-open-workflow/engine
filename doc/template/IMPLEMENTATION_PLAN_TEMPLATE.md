# Implementation Plan

## 1. Implementation Summary

2-3 sentences describing the overall approach and key decisions.

## 2. Change Manifest

List every file that will be created or modified:

```
CREATE:
- path/to/new-file.ts — purpose

MODIFY:
- path/to/existing-file.ts — what changes
```

## 3. Step-by-Step Plan

For each logical unit of work, provide:

### Step N: [Descriptive Title]

**File:** `path/to/file.ts`

**Action:** CREATE | MODIFY

**Rationale:** Why this change is needed (1 sentence)

**Pseudocode:**

```
// Describe the implementation in detailed pseudocode
// Include function signatures with types
// Show control flow and logic
// Specify error handling
// Note any validation rules

function exampleFunction(param: Type): ReturnType {
    // 1. Validate input
    //    - Check param is not null
    //    - Validate format matches X pattern

    // 2. Fetch required data
    //    - Call existingService.getData(param)
    //    - Handle not-found case: throw NotFoundError

    // 3. Apply business logic
    //    - If condition A: do X
    //    - Else if condition B: do Y
    //    - Edge case: when Z happens, handle by...

    // 4. Persist changes
    //    - Call repository.save(entity)
    //    - Emit event: 'entity.updated'

    // 5. Return result
    //    - Transform to response DTO
    //    - Include fields: a, b, c
}
```

**Dependencies:** List any imports or modules this step requires

**Tests Required:**

- Test case 1: description
- Test case 2: description

---

## 4. Data Changes (if applicable)

**Schema/Model Updates:**

```
// New fields, tables, or model changes with types
```

**Migration Notes:**

- Migration strategy (if needed)
- Backward compatibility considerations

## 5. Integration Points

For each external integration:

- **Service:** Name
- **Interaction:** What this feature does with it
- **Error Handling:** How failures are handled

## 6. Edge Cases & Error Handling

| Scenario          | Handling          |
| ----------------- | ----------------- |
| Edge case 1       | How it's handled  |
| Error condition 1 | Response/recovery |

## 7. Testing Strategy

**Unit Tests:**

- List key unit test scenarios

**Integration Tests:**

- List integration test scenarios

**Manual Verification:**

- Steps to manually verify the feature works

## 8. Implementation Order

Recommended sequence for implementation:

1. Step X — reason for ordering
2. Step Y — reason for ordering
3. ...
