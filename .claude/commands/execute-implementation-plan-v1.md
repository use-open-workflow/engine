# Execute Implementation Plan

You are a senior software engineer implementing a feature. Follow the approved plan exactly, translating the pseudocode into working code.

## Feature Requirements

@$1/FEATURE.md

## Research Output

@$1/RESEARCH_REPORT.md

## Approved Plan

@$1/IMPLEMENTATION_PLAN.md

## Instructions

Execute the implementation plan step-by-step. The plan has been reviewed and approved—your job is to translate it into production-quality code.

### Implementation Guidelines

1. **Follow the plan** — Implement exactly what the plan specifies. If you discover an issue with the plan, flag it explicitly before deviating.

2. **Follow the implementation order** — Execute steps in the sequence defined in "Implementation Order" section.

3. **Match existing patterns** — Use the reference implementations and coding patterns identified in the research phase.

4. **One step at a time** — Complete each step fully before moving to the next:

   - Write the code
   - Ensure it compiles/parses
   - Write the tests specified for that step
   - Verify tests pass

5. **Handle deviations explicitly** — If something in the plan doesn't work as expected:
   - Stop and explain the issue
   - Propose the minimal adjustment needed
   - Wait for approval before proceeding (or note the deviation clearly)

### Code Quality Standards

- Follow the coding patterns from the reference implementations
- Include appropriate error handling as specified in the plan
- Add comments only where logic is non-obvious
- Ensure type safety (if applicable)
- Write tests as specified in the plan

### Progress Reporting

After completing each step, briefly report:

```
✓ Step N: [Title]
  - Files changed: list
  - Tests added: count
  - Notes: any observations (optional)
```

If you encounter a blocker:

```
⚠ Step N: [Title] — BLOCKED
  - Issue: description
  - Proposed resolution: suggestion
```

## Execution

Begin with Step 1 from the Implementation Order. Work through each step sequentially until the feature is complete.

After all steps are complete, provide a final summary:

```
## Implementation Complete

**Files Created:** count
**Files Modified:** count
**Tests Added:** count

**Verification:**
- [ ] All tests pass
- [ ] Feature works as specified
- [ ] No unresolved deviations from plan

**Manual Verification Steps:**
(List from the plan)
```
