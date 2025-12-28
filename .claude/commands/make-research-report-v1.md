# Make Research Report

## Introduction

You are a senior software engineer performing context discovery for a new feature. This is the first phase in a 3-phase context engineering pipeline:

```
Research → Plan → Implementation
```

Your research report will inform the planning phase, enabling detailed implementation design.

## Goal

Autonomously explore the codebase and produce a concise research report that identifies all relevant files, dependencies, data flows, and constraints needed to implement the feature.

## Input

### Feature Requirements

@$1/FEATURE.md

## Instructions

Explore the codebase systematically to understand everything needed to implement this feature. Work autonomously—do not ask for guidance.

### Discovery Process

1. **Identify entry points** — Find where this feature would integrate (routes, handlers, UI components, etc.)
2. **Trace data flow** — Follow how related data moves through the system (API → service → repository → database)
3. **Map dependencies** — Identify modules, utilities, and external services this feature will interact with
4. **Find patterns** — Look for similar existing features to understand conventions and reusable code
5. **Check constraints** — Note any validation rules, auth requirements, rate limits, or business logic constraints

### What to Examine

- Directory structure and file organization
- Existing similar features (as implementation reference)
- Shared utilities, helpers, and base classes
- Configuration files and environment variables
- Database schemas / models / migrations
- API contracts and type definitions
- Test patterns and fixtures

## Output

### Template

@doc/template/RESEARCH_REPORT_TEMPLATE.md

### Save As (Report Location)

@$1/RESEARCH_REPORT.md

---

**Keep the output concise.** Aim for a report an engineer can review in 5 minutes. Avoid verbose explanations—use bullet points and file paths.
