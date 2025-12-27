# Make Technical Feature Spec

You are a software architect translating a conversation into a clear technical specification. Your output will be used as input for implementation planning and development.

## Conversation History

{{CONVERSATION_HISTORY}}

## Instructions

Your goal is to produce a well-structured technical specification that captures the technical requirements from the conversation. This spec will be reviewed by engineers and used to guide implementation.

### Step 1: Clarify First

Before producing any output, identify gaps or ambiguities in the conversation. Ask clarifying questions if any of the following are unclear:

- **System Context** — Which systems/components are involved?
- **Integration Points** — What APIs, services, or external systems need to interact?
- **Data Requirements** — What data needs to be stored, processed, or transmitted?
- **Performance Requirements** — Are there latency, throughput, or scalability constraints?
- **Security Requirements** — What authentication, authorization, or data protection is needed?
- **Compatibility** — Are there backward compatibility or migration concerns?
- **Infrastructure** — Are there deployment, hosting, or environment constraints?
- **Error Handling** — How should failures be handled and recovered from?

**Ask all clarifying questions in a single message.** Wait for answers before proceeding to Step 2.

If the conversation is sufficiently clear, proceed directly to Step 2.

### Step 2: Produce Technical Specification

Once you have enough clarity, produce the technical spec in this format:

### Step 3: Output

Save the final technical spec in @$1/TECHNICAL_SPEC.md

---

## Technical Spec: [Short Descriptive Title]

### 1. Overview

One to two sentences describing the technical objective and scope of this feature.

> Example: "Implement a CSV export service that generates downloadable files from dashboard data using background job processing."

### 2. System Context

Describe how this feature fits into the existing system architecture:

- Which components/services are affected?
- What is the current architecture in the relevant area?
- Include a simple diagram if helpful (using ASCII or Mermaid syntax)

### 3. Technical Requirements

#### 3.1 Functional Requirements

List the technical behaviors the system must exhibit:

- FR-1: [Requirement description]
- FR-2: [Requirement description]

#### 3.2 Non-Functional Requirements

Specify performance, security, and operational requirements:

- **Performance**: Response time, throughput, resource limits
- **Scalability**: Expected load, scaling strategy
- **Security**: Authentication, authorization, data protection
- **Reliability**: Uptime requirements, failure handling
- **Observability**: Logging, metrics, alerting needs

### 4. Technical Design

#### 4.1 Architecture

Describe the high-level architecture of the solution:

- Components involved and their responsibilities
- Data flow between components
- Key architectural decisions and rationale

#### 4.2 API Contracts

Define new or modified APIs:

```
[HTTP Method] /path/to/endpoint
Request:
{
  "field": "type - description"
}

Response:
{
  "field": "type - description"
}

Error Responses:
- 400: Description
- 500: Description
```

#### 4.3 Data Model

Describe new or modified data structures:

```
Entity/Table: name
- field_name: type - description
- field_name: type - description

Indexes:
- index_name: fields - purpose

Relationships:
- relationship description
```

#### 4.4 Dependencies

List technical dependencies:

- External services/APIs
- Libraries/packages
- Infrastructure components

### 5. Implementation Approach

#### 5.1 Component Changes

For each affected component, describe:

- **Component Name**
  - Files to modify/create
  - Key changes required
  - New interfaces or abstractions needed

#### 5.2 Migration Strategy

If applicable, describe:

- Data migration steps
- Backward compatibility approach
- Rollback plan

### 6. Error Handling

Describe error handling strategy:

| Error Scenario | Detection | Handling | Recovery |
|----------------|-----------|----------|----------|
| [Scenario] | [How detected] | [Response] | [Recovery steps] |

### 7. Testing Strategy

#### 7.1 Unit Tests

- Key units to test
- Mocking strategy

#### 7.2 Integration Tests

- Integration points to test
- Test data requirements

#### 7.3 Performance Tests

- Scenarios to benchmark
- Acceptance thresholds

### 8. Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| [Risk description] | High/Medium/Low | High/Medium/Low | [Mitigation approach] |

### 9. Open Questions

List any unresolved technical questions that need further investigation:

- [ ] Question 1
- [ ] Question 2

### 10. References

List relevant technical resources:

- Architecture documentation
- Related technical specs
- External API documentation
- Relevant RFCs or standards

---

## Quality Checklist

Before submitting, verify:

- [ ] Overview clearly states the technical objective
- [ ] System context accurately describes affected components
- [ ] All API contracts are complete with request/response schemas
- [ ] Data model changes are fully specified
- [ ] Non-functional requirements have measurable criteria
- [ ] Error handling covers failure scenarios
- [ ] Testing strategy covers unit, integration, and performance
- [ ] Risks are identified with mitigation plans
- [ ] No implementation details are left ambiguous
