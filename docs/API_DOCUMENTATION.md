# Generic Rule Engine - API Documentation

**Version**: v1  
**Base URL**: `/v1`  
**Authentication**: JWT Bearer Token  
**Content-Type**: `application/json`

## Table of Contents

1. [Quick Start Guide](#quick-start-guide)
2. [Authentication](#authentication)
3. [Common Patterns](#common-patterns)
4. [Error Handling](#error-handling)
5. [Resource Relationships & Dependencies](#resource-relationships--dependencies)
6. [State Management Patterns](#state-management-patterns)
7. [Performance & Scalability](#performance--scalability)
8. [Real-time Considerations](#real-time-considerations)
9. [Namespaces API](#namespaces-api)
10. [Fields API](#fields-api)
11. [Terminals API](#terminals-api)
12. [Functions API](#functions-api)
13. [Rules API](#rules-api)
14. [Workflows API](#workflows-api)
15. [Execution API](#execution-api)
16. [Admin API](#admin-api)
17. [Health & Metrics](#health--metrics)

---

## Quick Start Guide

This section provides the minimal set of API calls needed to get a basic UI working.

### 1. Generate JWT Token
```javascript
// Example JWT generation (client-side for demo only)
const jwt = generateJWT({
  clientId: "your-client-id",
  role: "admin",
  exp: Math.floor(Date.now() / 1000) + 3600 // 1 hour
});
```

### 2. Create Your First Namespace
```http
POST /v1/namespaces
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "id": "my-app",
  "description": "My application rules"
}
```

### 3. Add a Field
```http
POST /v1/namespaces/my-app/fields
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "fieldId": "score",
  "type": "number",
  "description": "User score"
}
```

### 4. Create and Publish a Function
```http
POST /v1/namespaces/my-app/functions
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "id": "high-score",
  "type": "max",
  "args": ["score"]
}
```

```http
POST /v1/namespaces/my-app/functions/high-score/publish
Authorization: Bearer <jwt_token>
```

### 5. Test Execution
```http
POST /v1/execute/namespaces/my-app/workflows/simple_check
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "data": {
    "score": 85
  }
}
```

**Expected Response Time**: < 100ms for simple operations, < 500ms for complex workflows.

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

### Token Refresh Pattern
```javascript
// Frontend pattern for token refresh
const refreshToken = async () => {
  const response = await fetch('/auth/refresh', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${currentToken}` }
  });
  const { token } = await response.json();
  localStorage.setItem('jwt_token', token);
  return token;
};
```

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

### Pagination
List endpoints support pagination:
```http
GET /v1/namespaces/{namespace}/functions?page=1&limit=20
```

**Response Headers**:
```http
X-Total-Count: 150
X-Page-Count: 8
Link: <https://api.example.com/v1/namespaces/ns/functions?page=2>; rel="next"
```

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
| HTTP Status | Error Code | Description | Frontend Action |
|-------------|------------|-------------|-----------------|
| 400 | `VALIDATION_ERROR` | Invalid request payload or parameters | Show field-specific errors |
| 401 | `UNAUTHORIZED` | Missing or invalid JWT token | Redirect to login |
| 403 | `FORBIDDEN` | Insufficient permissions for operation | Show permission denied message |
| 404 | `NOT_FOUND` | Resource does not exist | Show 404 page or create new |
| 409 | `DRAFT_EXISTS` | Draft already exists for entity | Show existing draft warning |
| 409 | `PUBLISH_DEPENDENCY_INACTIVE` | Referenced dependency not active | Highlight inactive dependencies |
| 412 | `PRECONDITION_FAILED` | ETag mismatch on conditional update | Refresh data and retry |
| 422 | `WORKFLOW_EXECUTION_FAILED` | Runtime execution error | Show execution error details |
| 429 | `RATE_LIMIT_EXCEEDED` | Too many requests | Show rate limit message, disable submit |
| 500 | `INTERNAL_SERVER_ERROR` | Unexpected server error | Show generic error, retry option |

### Error Handling Examples

#### Field Validation Error
```json
{
  "error": "VALIDATION_ERROR",
  "message": "Invalid field data",
  "details": [
    {
      "field": "type",
      "message": "Field type must be 'number' or 'string'"
    },
    {
      "field": "fieldId",
      "message": "Field ID is required"
    }
  ]
}
```

#### Dependency Error
```json
{
  "error": "PUBLISH_DEPENDENCY_INACTIVE",
  "message": "Cannot publish rule: referenced function 'max_income' is not active",
  "details": [
    {
      "resource": "function",
      "id": "max_income",
      "status": "draft"
    }
  ]
}
```

---

## Resource Relationships & Dependencies

### Hierarchy Structure
```
Namespace
├── Fields (independent)
├── Terminals (independent)
├── Functions (can reference Fields)
├── Rules (can reference Fields, Functions, other Rules)
└── Workflows (can reference Rules, Terminals)
```

### Required Order of Operations
1. **Namespace** must exist first
2. **Fields** can be created anytime (referenced by Functions/Rules)
3. **Terminals** can be created anytime (referenced by Workflows)
4. **Functions** require referenced Fields to exist
5. **Rules** require referenced Fields/Functions to be active
6. **Workflows** require referenced Rules/Terminals to be active

### Cross-Resource References
- **Functions** reference Fields via `args` array
- **Rules** reference Fields, Functions, and other Rules via conditions
- **Workflows** reference Rules and Terminals via steps

### Dependency Validation
```javascript
// Frontend pattern for dependency checking
const checkDependencies = async (namespace, resourceType, resourceId) => {
  const response = await fetch(`/v1/namespaces/${namespace}/${resourceType}/${resourceId}/dependencies`);
  const { dependencies, status } = await response.json();
  
  if (status === 'inactive') {
    showWarning(`This ${resourceType} has inactive dependencies`);
  }
  
  return dependencies;
};
```

---

## State Management Patterns

### Resource Lifecycle States

#### Draft State
- **Allowed Operations**: Update, Publish, Delete
- **UI Indicators**: "Draft" badge, edit buttons enabled
- **Validation**: Basic validation only

#### Active State
- **Allowed Operations**: View, Create new draft, Deactivate
- **UI Indicators**: "Active" badge, read-only view
- **Validation**: Full validation including dependencies

#### Inactive State
- **Allowed Operations**: View, Reactivate
- **UI Indicators**: "Inactive" badge, reactivate button
- **Validation**: No validation (historical state)

### State Transition Endpoints
```http
# Draft → Active
POST /v1/namespaces/{namespace}/functions/{id}/publish

# Active → Inactive
POST /v1/namespaces/{namespace}/functions/{id}/deactivate

# Inactive → Active
POST /v1/namespaces/{namespace}/functions/{id}/activate
```

### Frontend State Management
```javascript
// Example React state management
const [functionState, setFunctionState] = useState({
  status: 'draft',
  canEdit: true,
  canPublish: false,
  canDeactivate: false
});

const updateFunctionState = (status) => {
  setFunctionState({
    status,
    canEdit: status === 'draft',
    canPublish: status === 'draft',
    canDeactivate: status === 'active'
  });
};
```

---

## Performance & Scalability

### Response Time Expectations
| Operation Type | Expected Time | Notes |
|----------------|---------------|-------|
| Simple CRUD | < 100ms | Namespace, Field, Terminal operations |
| Complex CRUD | < 200ms | Function, Rule, Workflow operations |
| Publish | < 500ms | Includes validation and cache updates |
| Execution | < 1000ms | Simple workflows |
| Execution | < 5000ms | Complex workflows with many steps |
| Cache Refresh | < 2000ms | Background operation |

### Pagination Limits
- **Default**: 20 items per page
- **Maximum**: 100 items per page
- **Recommended**: 50 items per page for optimal performance

### Bulk Operations
```http
# Bulk create fields
POST /v1/namespaces/{namespace}/fields/bulk
Content-Type: application/json

{
  "fields": [
    {"fieldId": "field1", "type": "number"},
    {"fieldId": "field2", "type": "string"}
  ]
}
```

### Caching Strategy
- **GET requests**: Cached for 5 minutes
- **Cache invalidation**: Automatic on publish operations
- **ETags**: Used for conditional requests
- **Cache headers**: `Cache-Control: max-age=300`

---

## Real-time Considerations

### Cache Refresh Detection
The API uses checksum-based change detection for real-time updates:

```http
GET /admin/cache/stats/{namespace}
Authorization: Bearer <jwt_token>
```

**Response**:
```json
{
  "data": {
    "namespace": "my-app",
    "checksum": "abc123...",
    "lastRefresh": "2025-06-24T10:30:00Z",
    "stalenessSeconds": 0,
    "cacheStatus": "fresh",
    "isStale": false
  }
}
```

### Polling Pattern
```javascript
// Frontend polling pattern for cache updates
const pollForChanges = async (namespace, initialChecksum) => {
  const interval = setInterval(async () => {
    const response = await fetch(`/admin/cache/stats/${namespace}`);
    const { data } = await response.json();
    
    if (data.checksum !== initialChecksum) {
      clearInterval(interval);
      refreshData(); // Reload your UI data
    }
  }, 5000); // Poll every 5 seconds
  
  return interval;
};
```

### Webhook Support (Future)
```http
POST /v1/webhooks
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "url": "https://your-app.com/webhook",
  "events": ["function.published", "rule.published", "workflow.published"],
  "namespace": "my-app"
}
```

### Optimistic Updates
```javascript
// Frontend pattern for optimistic updates
const publishFunction = async (functionId) => {
  // Optimistically update UI
  setFunctionStatus(functionId, 'publishing');
  
  try {
    await fetch(`/functions/${functionId}/publish`, { method: 'POST' });
    setFunctionStatus(functionId, 'active');
  } catch (error) {
    // Revert on error
    setFunctionStatus(functionId, 'draft');
    showError('Publish failed');
  }
};
```

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

## Frontend Engineer Best Practices & Integration Tips

### What to Look for in This API Documentation

1. **Authentication Scheme**
   - All endpoints require JWT Bearer tokens in the `Authorization` header.
   - See [Authentication](#authentication) for header and token details.

2. **Endpoints, Resources, and HTTP Methods**
   - Each resource (namespaces, fields, functions, rules, workflows, execution) is documented with its HTTP method and path.
   - See the Table of Contents for a resource map.

3. **Data Contracts: Request & Response Payloads**
   - Every endpoint includes example request and response JSON.
   - Field names, types, and required/optional status are shown.

4. **Error Handling Contract**
   - Standard error response structure and error codes are documented.
   - See [Error Handling](#error-handling) for status codes, error codes, and frontend actions.

5. **Common Patterns & Headers**
   - ETag/If-Match for optimistic locking, pagination, and rate limiting are explained in [Common Patterns](#common-patterns).

6. **End-to-End Examples**
   - The [Complete Workflow Setup](#complete-workflow-setup) at the end of this document demonstrates a real-world "happy path" for frontend integration.

7. **Resource Relationships & State**
   - See [Resource Relationships & Dependencies](#resource-relationships--dependencies) and [State Management Patterns](#state-management-patterns) for how resources connect and transition.

8. **Performance & Real-time**
   - [Performance & Scalability](#performance--scalability) and [Real-time Considerations](#real-time-considerations) provide guidance for polling, caching, and expected response times.

---

## Complete Workflow Setup (Happy Path Example)

This example demonstrates the typical sequence of API calls for a frontend to create, configure, and execute a workflow:

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