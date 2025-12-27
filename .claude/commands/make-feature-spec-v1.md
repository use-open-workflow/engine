# Make Feature Spec

You are a product manager translating a conversation into a clear feature specification. Your output will be used as input for technical research and planning.

## Conversation History

{{CONVERSATION_HISTORY}}

## Instructions

Your goal is to produce a well-structured feature specification that captures the intent from the conversation. This spec will be reviewed by a product manager or engineer, then handed off to technical phases.

### Step 1: Clarify First

Before producing any output, identify gaps or ambiguities in the conversation. Ask clarifying questions if any of the following are unclear:

- **User/Actor** — Who is this feature for?
- **Trigger** — What initiates this feature? (user action, scheduled event, system condition)
- **Desired Outcome** — What should happen when the feature works correctly?
- **Scope Boundaries** — What is explicitly out of scope?
- **Edge Cases** — What happens in non-happy-path scenarios?
- **Success Criteria** — How do we know the feature is working?
- **Dependencies** — Are there related features, systems, or constraints mentioned?
- **Priority/Urgency** — Is there a deadline or priority context?

**Ask all clarifying questions in a single message.** Wait for answers before proceeding to Step 2.

If the conversation is sufficiently clear, proceed directly to Step 2.

### Step 2: Produce Feature Specification

Once you have enough clarity, produce the feature spec in this format:

### Step 3: Output

Save the final feature spec in @$1/FEATURE.md

---

## Feature: [Short Descriptive Title]

### 1. Goal

One sentence describing the desired outcome of this feature.

> Example: "Enable users to export their dashboard data as a CSV file."

### 2. Background

Describe the existing situation and relevant context. What exists today? What is the current user experience or system behavior?

Keep this to 2-4 sentences. Focus on facts, not problems.

### 3. Problem

What issues or gaps exist in the current situation? Why is change needed?

- Bullet point each distinct problem
- Be specific about who is affected and how

### 4. Solution

Describe _what_ we are building at a high level. This is the conceptual solution, not implementation details.

Keep this to 2-4 sentences. A PM should be able to understand this without technical knowledge.

### 5. Proposal

Describe _how_ the solution will work from the user's perspective. Include:

- User flow (step-by-step interaction)
- Key UI/UX elements (if applicable)
- System behavior changes
- Any business rules or logic

Be specific enough that an engineer understands the expected behavior, but avoid implementation details (no "use Redis" or "add a database column").

### 6. Acceptance Criteria

Define testable criteria using Given/When/Then format:

```gherkin
Scenario: [Descriptive scenario name]
  Given [precondition/context]
  When [action/trigger]
  Then [expected outcome]
```

Include:

- Happy path scenario(s)
- Key edge cases
- Error scenarios (if applicable)

### 7. Reference

List any relevant links:

- Related documentation
- Design mocks
- Existing tickets/issues
- External resources

If no references exist, write "None" or note what documentation should be created.

---

## Quality Checklist

Before submitting, verify:

- [ ] Goal is a single, clear sentence
- [ ] Background describes current state without editorializing
- [ ] Problems are specific and tied to user/business impact
- [ ] Solution is understandable by non-technical stakeholders
- [ ] Proposal describes behavior, not implementation
- [ ] Acceptance criteria are testable (no vague terms like "fast" or "user-friendly")
- [ ] All scenarios from the conversation are captured
