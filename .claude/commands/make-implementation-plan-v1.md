# Make Implementation Plan

## Introduction

You are a senior software engineer creating a detailed implementation plan. This is the second phase in a 3-phase context engineering pipeline:

```
Research → Plan → Implementation
```

Your implementation plan will guide the implementation phase, enabling any engineer to implement the feature by following the plan step-by-step.

## Goal

Using the research report, produce a comprehensive plan with pseudocode that an engineer can validate before any code is written.

## Input

### Feature Requirements

@$1/FEATURE.md

### Research Report

@$1/RESEARCH_REPORT.md

## Instructions

Create a detailed, step-by-step implementation plan. Work autonomously—do not ask for guidance.

### Plan Requirements

The plan should be specific enough that:

- An engineer can validate the approach before any code is written
- The implementation phase can follow it mechanically
- Edge cases and error handling are addressed upfront

### What to Include

- Every file to be created or modified
- Detailed pseudocode with types and signatures
- Control flow and business logic
- Error handling for each step
- Test cases for each component
- Integration points and failure modes
- Recommended implementation order

## Output

### Template

@doc/template/IMPLEMENTATION_PLAN_TEMPLATE.md

### Save As (Location)

@$1/IMPLEMENTATION_PLAN.md

---

**Be thorough.** The engineer reviewing this plan should be able to identify any design issues before implementation starts. Include enough detail that implementation becomes primarily a translation exercise.
