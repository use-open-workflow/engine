# Execute Implementation Plan

## Introduction

You are a senior software engineer implementing a feature. This is the third phase in a 3-phase context engineering pipeline:

```
Research → Plan → Implementation
```

Your implementation will translate the approved plan into production-quality code, following the pseudocode step-by-step.

## Goal

Using the approved implementation plan, produce working code that implements the feature exactly as designed.

## Input

### Feature Requirements

@$1/FEATURE.md

### Research Report

@$1/RESEARCH_REPORT.md

### Approved Plan

@$1/IMPLEMENTATION_PLAN.md

## Instructions

Execute the implementation plan step-by-step. The plan has been reviewed and approved—your job is to translate it into production-quality code. Work autonomously—do not ask for guidance.

### Execution Guidelines

1. **Follow the plan** — Implement exactly what the plan specifies. If you discover an issue with the plan, flag it explicitly before deviating.

2. **Follow the implementation order** — Execute steps in the sequence defined in the "Implementation Order" section.

3. **Match existing patterns** — Use the reference implementations and coding patterns identified in the research phase.

4. **One step at a time** — Complete each step fully before moving to the next:

   - Write the code
   - Ensure it compiles/parses
   - Write the tests specified for that step
   - Verify tests pass

5. **Handle deviations explicitly** — If something in the plan doesn't work as expected:
   - Stop and explain the issue
   - Propose the minimal adjustment needed
   - Note the deviation clearly in the implementation summary

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

## Output

### Template

@doc/template/IMPLEMENTATION_SUMMARY_TEMPLATE.md

### Save As (Location)

@$1/IMPLEMENTATION_SUMMARY.md

---

**Execute in one shot.** The plan has been approved—implement all steps completely without stopping for feedback. Document any deviations in the implementation summary rather than pausing for approval.
