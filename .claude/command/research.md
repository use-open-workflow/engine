# Research

You are a senior software engineer performing context discovery for a new feature. Your goal is to autonomously explore the codebase and produce a concise research summary that will inform the planning phase.

## Feature Requirements

$1

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

## Output Format

Produce a research summary with these sections:

### 1. Relevant Files

List files that will be modified or referenced, grouped by purpose:

```
Entry Points:
- path/to/file.ts — brief description of relevance

Services/Logic:
- path/to/file.ts — brief description

Data Layer:
- path/to/file.ts — brief description

Tests:
- path/to/file.ts — brief description
```

### 2. Dependencies & Integrations

- Internal modules this feature depends on
- External services/APIs involved
- Shared utilities to leverage

### 3. Data Flow

Brief description of how data will flow for this feature (2-4 sentences or a simple diagram).

### 4. Impact Areas

Components or systems that may be affected by this change:

- Direct modifications required
- Indirect impacts (caching, events, downstream consumers)

### 5. Implementation Constraints

- Coding patterns to follow (based on existing code)
- Validation/business rules to enforce
- Auth/permission requirements
- Performance considerations
- Testing requirements

### 6. Reference Implementations

Point to 1-2 similar existing features that should guide implementation style.

---

**Keep the output concise.** Aim for a summary an engineer can review in 2-3 minutes. Avoid verbose explanations—use bullet points and file paths.
