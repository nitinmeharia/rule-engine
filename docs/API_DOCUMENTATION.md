# Generic Rule Engine - API Documentation

**Version**: v1  
**Base URL**: `/v1`  
**Authentication**: JWT Bearer Token  
**Content-Type**: `application/json`

## Table of Contents

1. [Authentication](#authentication)
2. [Common Patterns](#common-patterns)
3. [Error Handling](#error-handling)
4. [Namespaces API](#namespaces-api)
5. [Fields API](#fields-api)
6. [Terminals API](#terminals-api)
7. [Functions API](#functions-api)
8. [Rules API](#rules-api)
9. [Workflows API](#workflows-api)
10. [Execution API](#execution-api)
11. [Admin API](#admin-api)
12. [Health & Metrics](#health--metrics)

---

## Authentication

All API endpoints require JWT authentication via the `Authorization` header.

### Headers
```http
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### JWT Requirements
- **Algorithm**: HS256
- **Required Claims**:
  - `clientId` (string): Unique client identifier
  - `role` (string or array): User roles (`admin`, `viewer`, `executor`)
  - `exp` (number): Token expiration timestamp

### Role Permissions
| Role | Permissions |
|------|-------------|
| `admin` | Full CRUD access, publish operations |
| `viewer` | Read-only access to all resources |
| `executor` | Execute workflows, read active configs |

---

## Common Patterns

### Versioned Entities
Functions, Rules, and Workflows follow a versioned lifecycle:
- **draft** → **active** → **inactive**
- Only one active and one draft version per entity
- Publish operations transition draft → active

### ETag Support
- GET requests return `ETag` header with content hash
- PUT requests require `If-Match` header with ETag value
- Prevents lost updates through optimistic locking

### Request IDs
All responses include a `X-Request-ID` header for tracing.

---

## Error Handling

### Standard Error Response
```json
{
  "error": "ERROR_CODE",
  "message": "Human-readable error description",
  "requestId": "req_12345",
  "details": [
    {
      "field": "fieldName",
      "message": "Field-specific error"
    }
  ]
}
```

### Common Error Codes
| HTTP Status | Error Code | Description |
|-------------|------------|-------------|
| 400 | `VALIDATION_ERROR` | Invalid request payload or parameters |
| 401 | `UNAUTHORIZED` | Missing or invalid JWT token |
| 403 | `FORBIDDEN` | Insufficient permissions for operation |
| 404 | `NOT_FOUND` | Resource does not exist |
| 409 | `DRAFT_EXISTS` | Draft already exists for entity |
| 409 | `PUBLISH_DEPENDENCY_INACTIVE` | Referenced dependency not active |
| 412 | `PRECONDITION_FAILED` | ETag mismatch on conditional update |
| 422 | `WORKFLOW_EXECUTION_FAILED` | Runtime execution error |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests |
| 500 | `INTERNAL_SERVER_ERROR` | Unexpected server error |

---

## Namespaces API

### List Namespaces
```http
GET /v1/namespaces
```
**Roles**: `viewer`, `admin`, `executor`

**Response 200**:
```json
[
  "lender_abc",
  "fraud_detection", 
  "credit_scoring"
]
```

### Create Namespace
```http
POST /v1/namespaces
```
**Roles**: `admin`

**Request Body**:
```json
{
  "id": "lender_abc",
  "description": "Rules for lender ABC operations"
}
```

**Response 201**:
```json
{
  "success": true,
  "namespace": {
    "id": "lender_abc",
    "description": "Rules for lender ABC operations",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  }
}
```

---

## Fields API

### List Fields
```http
GET /v1/namespaces/{namespace}/fields
```
**Roles**: `viewer`, `admin`, `executor`

**Response 200**:
```json
[
  {
    "fieldId": "salary",
    "type": "number",
    "description": "Monthly salary of applicant",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  },
  {
    "fieldId": "occupation",
    "type": "string",
    "description": "Job title or category",
    "createdAt": "2025-06-24T08:05:00Z",
    "createdBy": "client_12345"
  }
]
```

### Create Field
```http
POST /v1/namespaces/{namespace}/fields
```
**Roles**: `admin`

**Request Body**:
```json
{
  "fieldId": "salary",
  "type": "number",
  "description": "Monthly salary of applicant"
}
```

**Field Types**: `number`, `string`

**Response 201**:
```json
{
  "success": true,
  "field": {
    "fieldId": "salary",
    "type": "number",
    "description": "Monthly salary of applicant",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  }
}
```

---

## Terminals API

### List Terminals
```http
GET /v1/namespaces/{namespace}/terminals
```
**Roles**: `viewer`, `admin`, `executor`

**Response 200**:
```json
[
  {
    "id": "accept",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  },
  {
    "id": "reject",
    "createdAt": "2025-06-24T08:00:00Z", 
    "createdBy": "client_12345"
  }
]
```

### Create Terminal
```http
POST /v1/namespaces/{namespace}/terminals
```
**Roles**: `admin`

**Request Body**:
```json
{
  "id": "accept"
}
```

**Response 201**:
```json
{
  "success": true,
  "terminal": {
    "id": "accept",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  }
}
```

---

## Functions API

### Create Function (Draft)
```http
POST /v1/namespaces/{namespace}/functions
```
**Roles**: `admin`

**Request Body**:
```json
{
  "id": "max_income",
  "type": "max",
  "args": ["salary", "bonus"]
}
```

**Function Types** (returnType computed automatically):
- `max`: Maximum of numeric fields → returns `number`
- `sum`: Sum of numeric fields → returns `number`
- `avg`: Average of numeric fields → returns `number`
- `in`: Check if value exists in array → returns `bool`

**For `in` functions**:
```json
{
  "id": "valid_occupations",
  "type": "in", 
  "values": ["salaried", "self_employed", "business"]
}
```

**Response 201**:
```json
{
  "status": "draft",
  "function": {
    "id": "max_income",
    "version": 1,
    "status": "draft",
    "type": "max",
    "args": ["salary", "bonus"],
    "returnType": "number",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  }
}
```

### Update Function Draft
```http
PUT /v1/namespaces/{namespace}/functions/{functionId}/versions/draft
```
**Roles**: `admin`  
**Headers**: `If-Match: <etag>`

**Request Body**: Complete function object with modifications.

**Response 200**:
```json
{
  "function": {
    "id": "max_income",
    "version": 1,
    "status": "draft",
    "type": "max", 
    "args": ["salary", "bonus", "commission"],
    "returnType": "number",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345",
    "updatedAt": "2025-06-24T08:30:00Z"
  }
}
```

### Publish Function
```http
POST /v1/namespaces/{namespace}/functions/{functionId}/publish
```
**Roles**: `admin`

**Response 200**:
```json
{
  "status": "active",
  "function": {
    "id": "max_income",
    "version": 1,
    "status": "active",
    "type": "max",
    "args": ["salary", "bonus", "commission"],
    "returnType": "number",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345",
    "publishedAt": "2025-06-24T09:00:00Z",
    "publishedBy": "client_12345"
  }
}
```

### Get Function
```http
GET /v1/namespaces/{namespace}/functions/{functionId}
```
**Roles**: `viewer`, `admin`, `executor`

Returns active version if available, otherwise draft version.

**Response 200**: Function object with `ETag` header.

### Get Function History
```http
GET /v1/namespaces/{namespace}/functions/{functionId}/history
```
**Roles**: `viewer`, `admin`

**Response 200**:
```json
[
  {
    "version": 2,
    "status": "active",
    "createdAt": "2025-06-24T09:45:00Z",
    "createdBy": "client_12345",
    "publishedAt": "2025-06-24T10:00:00Z",
    "publishedBy": "client_12345"
  },
  {
    "version": 1, 
    "status": "inactive",
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345",
    "publishedAt": "2025-06-24T09:00:00Z",
    "publishedBy": "client_12345"
  }
]
```

---

## Rules API

### Create Rule (Draft)
```http
POST /v1/namespaces/{namespace}/rules
```
**Roles**: `admin`

**Request Body**:
```json
{
  "id": "income_check",
  "logic": "AND",
  "conditions": [
    {
      "type": "field",
      "fieldId": "salary",
      "operator": ">=",
      "value": 50000
    },
    {
      "type": "function", 
      "functionId": "max_income",
      "operator": ">=",
      "value": 60000
    },
    {
      "type": "rule",
      "ruleId": "occupation_check"
    }
  ]
}
```

**Logic Types**: `AND` (default), `OR`

**Condition Types**:
- `field`: Compare input field value
- `function`: Compare function result
- `rule`: Evaluate nested rule

**Operators**:
- Numbers: `==`, `!=`, `>`, `<`, `>=`, `<=`
- Strings: `==`, `!=`

**Response 201**:
```json
{
  "status": "draft",
  "rule": {
    "id": "income_check",
    "version": 1,
    "status": "draft",
    "logic": "AND",
    "conditions": [...],
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  }
}
```

### Update Rule Draft
```http
PUT /v1/namespaces/{namespace}/rules/{ruleId}/versions/draft
```
**Roles**: `admin`  
**Headers**: `If-Match: <etag>`

### Publish Rule  
```http
POST /v1/namespaces/{namespace}/rules/{ruleId}/publish
```
**Roles**: `admin`

### Get Rule
```http
GET /v1/namespaces/{namespace}/rules/{ruleId}
```
**Roles**: `viewer`, `admin`, `executor`

### Get Rule History
```http
GET /v1/namespaces/{namespace}/rules/{ruleId}/history
```
**Roles**: `viewer`, `admin`

---

## Workflows API

### Create Workflow (Draft)
```http
POST /v1/namespaces/{namespace}/workflows
```
**Roles**: `admin`

**Request Body**:
```json
{
  "id": "loan_eligibility",
  "startAt": "check_income",
  "steps": {
    "check_income": {
      "type": "rule",
      "ruleId": "income_check",
      "onSuccess": "check_occupation",
      "onFailure": "reject"
    },
    "check_occupation": {
      "type": "rule", 
      "ruleId": "occupation_check",
      "onSuccess": "accept",
      "onFailure": "reject"
    },
    "accept": {
      "type": "terminal"
    },
    "reject": {
      "type": "terminal"
    }
  }
}
```

**Step Types**:
- `rule`: Evaluate a rule, branch based on result
- `terminal`: End state of workflow

**Response 201**:
```json
{
  "status": "draft",
  "workflow": {
    "id": "loan_eligibility",
    "version": 1,
    "status": "draft",
    "startAt": "check_income",
    "steps": {...},
    "createdAt": "2025-06-24T08:00:00Z",
    "createdBy": "client_12345"
  }
}
```

### Update Workflow Draft
```http
PUT /v1/namespaces/{namespace}/workflows/{workflowId}/versions/draft
```
**Roles**: `admin`  
**Headers**: `If-Match: <etag>`

### Publish Workflow
```http
POST /v1/namespaces/{namespace}/workflows/{workflowId}/publish
```
**Roles**: `admin`

### Get Workflow
```http
GET /v1/namespaces/{namespace}/workflows/{workflowId}
```
**Roles**: `viewer`, `admin`, `executor`

### Get Workflow History
```http
GET /v1/namespaces/{namespace}/workflows/{workflowId}/history
```
**Roles**: `viewer`, `admin`

---

## Execution API

### Execute Workflow
```http
POST /v1/execute/namespaces/{namespace}/workflows/{workflowId}?trace=full
```
**Roles**: `executor`

**Query Parameters**:
- `trace` (optional): Trace verbosity level
  - `full` (default): Complete detailed trace with all values
  - `simple`: Minimal trace with only rule IDs and outcomes

**Request Body**:
```json
{
  "data": {
    "salary": 60000,
    "bonus": 5000,
    "occupation": "salaried"
  }
}
```

**Response 200**:
```json
{
  "terminal": "accept",
  "correlationId": "exec_12345",
  "executionTime": "2025-06-24T10:15:30Z",
  "trace": {
    "workflowId": "loan_eligibility",
    "workflowVersion": 2,
    "steps": [
      {
        "stepId": "check_income",
        "type": "rule",
        "ruleId": "income_check",
        "ruleVersion": 3,
        "result": true,
        "conditions": [
          {
            "type": "field",
            "fieldId": "salary",
            "operator": ">=",
            "value": 50000,
            "actualValue": 60000,
            "result": true
          },
          {
            "type": "function",
            "functionId": "max_income", 
            "functionVersion": 1,
            "operator": ">=",
            "value": 60000,
            "actualValue": 65000,
            "result": true
          }
        ],
        "nextStep": "check_occupation"
      },
      {
        "stepId": "check_occupation",
        "type": "rule",
        "ruleId": "occupation_check",
        "ruleVersion": 2,
        "result": true,
        "conditions": [
          {
            "type": "function",
            "functionId": "valid_occupations",
            "functionVersion": 1,
            "operator": "==",
            "value": true,
            "actualValue": true,
            "result": true
          }
        ],
        "nextStep": "accept"
      },
      {
        "stepId": "accept",
        "type": "terminal"
      }
    ]
  }
}
```

**Response 200** (Simple Trace):
```json
{
  "terminal": "accept",
  "correlationId": "exec_12345", 
  "executionTime": "2025-06-24T10:15:30Z",
  "trace": {
    "workflowId": "loan_eligibility",
    "workflowVersion": 2,
    "steps": [
      {
        "stepId": "check_income",
        "type": "rule",
        "ruleId": "income_check",
        "ruleVersion": 3,
        "result": true,
        "nextStep": "check_occupation"
      },
      {
        "stepId": "check_occupation",
        "type": "rule",
        "ruleId": "occupation_check",
        "ruleVersion": 2,
        "result": true,
        "nextStep": "accept"
      },
      {
        "stepId": "accept",
        "type": "terminal"
      }
    ]
  }
}
```

**Error Response 422** (Execution Failed):
```json
{
  "error": "WORKFLOW_EXECUTION_FAILED",
  "message": "Rule 'income_check' not found in active configuration",
  "requestId": "req_12345",
  "correlationId": "exec_12345",
  "trace": {
    "workflowId": "loan_eligibility",
    "failedAt": "check_income",
    "error": "Missing active rule: income_check"
  }
}
```

---

## Admin API

### Force Cache Reload
```http
POST /v1/admin/reload?namespace={namespace}
```
**Roles**: `admin`

Triggers immediate cache refresh for specified namespace (optional parameter).

**Response 200**:
```json
{
  "reloaded": true,
  "namespace": "lender_abc",
  "timestamp": "2025-06-24T10:30:00Z"
}
```

### Bulk Import
```http
POST /v1/namespaces/{namespace}/bulk-import
```
**Roles**: `admin`

**Request Body**:
```json
{
  "fields": [
    {
      "fieldId": "salary",
      "type": "number",
      "description": "Monthly salary"
    }
  ],
  "functions": [
    {
      "id": "max_income",
      "type": "max", 
      "args": ["salary", "bonus"]
    }
  ],
  "rules": [
    {
      "id": "income_check",
      "logic": "AND",
      "conditions": [...]
    }
  ],
  "workflows": [
    {
      "id": "loan_eligibility",
      "startAt": "check_income",
      "steps": {...}
    }
  ],
  "terminals": [
    {"id": "accept"},
    {"id": "reject"}
  ]
}
```

**Response 201**:
```json
{
  "success": true,
  "imported": {
    "fields": 5,
    "functions": 3,
    "rules": 8,
    "workflows": 2,
    "terminals": 2
  },
  "errors": []
}
```

---

## Health & Metrics

### Health Check
```http
GET /v1/healthz
```
**Roles**: None (public endpoint)

**Response 200**:
```json
{
  "status": "healthy",
  "timestamp": "2025-06-24T10:30:00Z",
  "version": "1.0.0",
  "checks": {
    "database": "healthy",
    "cache": "healthy"
  }
}
```

### Metrics (Prometheus)
```http
GET /v1/metrics
```
**Roles**: None (public endpoint)

Returns Prometheus-formatted metrics:
```
# HELP exec_duration_ms Execution duration in milliseconds
# TYPE exec_duration_ms histogram
exec_duration_ms_bucket{namespace="lender_abc",le="5"} 245
exec_duration_ms_bucket{namespace="lender_abc",le="10"} 891

# HELP cache_reload_ms Cache reload duration in milliseconds  
# TYPE cache_reload_ms histogram
cache_reload_ms_bucket{namespace="lender_abc",le="100"} 45

# HELP rule_evaluations_total Total number of rule evaluations
# TYPE rule_evaluations_total counter
rule_evaluations_total{namespace="lender_abc",outcome="true"} 1523
rule_evaluations_total{namespace="lender_abc",outcome="false"} 892
```

---

## Rate Limiting

All endpoints are subject to rate limiting:
- **Global**: 1000 requests per minute per IP
- **Per-user**: 100 requests per minute per JWT clientId
- **Execution**: 200 requests per minute per JWT clientId

Rate limit headers in responses:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

---

## API Versioning

- Current version: `v1`
- All endpoints prefixed with `/v1`
- Breaking changes will introduce new version (`v2`)
- Previous versions supported for 12 months minimum

---

## Examples

### Complete Workflow Setup
```bash
# 1. Create namespace
curl -X POST /v1/namespaces \
  -H "Authorization: Bearer $JWT" \
  -d '{"id":"demo","description":"Demo namespace"}'

# 2. Create fields
curl -X POST /v1/namespaces/demo/fields \
  -H "Authorization: Bearer $JWT" \
  -d '{"fieldId":"income","type":"number"}'

# 3. Create terminals  
curl -X POST /v1/namespaces/demo/terminals \
  -H "Authorization: Bearer $JWT" \
  -d '{"id":"approve"}'

# 4. Create and publish rule
curl -X POST /v1/namespaces/demo/rules \
  -H "Authorization: Bearer $JWT" \
  -d '{"id":"income_rule","conditions":[{"type":"field","fieldId":"income","operator":">=","value":50000}]}'

curl -X POST /v1/namespaces/demo/rules/income_rule/publish \
  -H "Authorization: Bearer $JWT"

# 5. Create and publish workflow  
curl -X POST /v1/namespaces/demo/workflows \
  -H "Authorization: Bearer $JWT" \
  -d '{"id":"simple_check","startAt":"check","steps":{"check":{"type":"rule","ruleId":"income_rule","onSuccess":"approve","onFailure":"reject"},"approve":{"type":"terminal"},"reject":{"type":"terminal"}}}'

curl -X POST /v1/namespaces/demo/workflows/simple_check/publish \
  -H "Authorization: Bearer $JWT"

# 6. Execute workflow
curl -X POST /v1/execute/namespaces/demo/workflows/simple_check \
  -H "Authorization: Bearer $JWT" \
  -d '{"data":{"income":60000}}'
```

---

**Document Version**: 1.0  
**Last Updated**: 2025-06-24  
**API Version**: v1 