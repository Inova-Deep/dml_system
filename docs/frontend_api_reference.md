# INOVA DML: Frontend API Integration Reference

Welcome to the INOVA DML API Documentation specifically tailored for the Frontend team. This document covers authentications, standard JSON workflows, REST capabilities, and the exact payloads required for smooth API consumption.

## 1. Authentication & Security Flow

All endpoints (except `POST /auth/login` and `GET /health`) require strict JWT Bearer authentication. Our Identity layer bounds every request to a specific `TenantID` isolated inside the Token.

### 1.1 Acquiring the Token

**Endpoint:** `POST /auth/login`

**Request Body:**
```json
{
  "email": "hemish.patel@inova.krd",
  "password": "Testing123!"
}
```

**Success Response (200 OK):**
```json
{
  "token": "eyJhbG...",
  "user": {
    "id": "e633d2ea-8b43-...-b1f4",
    "email": "hemish.patel@inova.krd",
    "displayName": "Hemish Director"
  },
  "roles": ["SYSTEM_ADMIN"],
  "tenantId": "c4d3..."
}
```

### 1.2 Using the Token

You **must** supply this token in the `Authorization` HTTP Header on all future calls:
```http
Authorization: Bearer <token>
```
*Note: The old `X-Tenant-ID` header has been entirely stripped and disabled. DO NOT send it. The system infers your Tenant boundaries cryptographically off your JWT Signature directly.*

---

## 2. API Conventions & Standard Responses

The backend utilizes standardized predictable struct responses to standardize error handling on Redux/Vuex contexts. 

### 2.1 Standardized Error Layouts

Whether receiving a `400 Bad Request` or `404 Not Found`, the format is always identically encapsulated:

```json
{
  "error": "Message describing issue explicitly",
  "details": [
     "Field 'password' validation failed: min 8 length required"
  ]
}
```

### 2.2 Pagination Bounds

When explicitly calling `.List` endpoints (like `GET /employees`), append standard URL Queries:
`GET /employees?page=1&size=50&search=Smith`

```json
{
  "data": [
    { /* entity objects here */ }
  ],
  "metadata": {
    "currentPage": 1,
    "pageSize": 50,
    "totalCount": 105,
    "totalPages": 3
  }
}
```

---

## 3. Core Operational Endpoints

### 3.1 Fetching Organizational Hierarchies

**Business Units**
- `GET /business-units`
- `POST /business-units` - Requires `ADMIN` Role.

**Departments**
- `GET /departments` - Use for dropdown fields natively.

**Job Titles**
- `GET /job-titles` - Includes native `grade` values (e.g. `GRADE-A`).

### 3.2 Reading Employee & Hierarchy Charts

**Fetch All Employees**
- `GET /employees?page=1`
- Includes native mapping pointers to Business Unit IDs natively.

**Fetch HR Relational Tree (Manager Mapping)**
- `GET /employees/{employeeID}/hierarchy`
- *Returns the employee mapped linearly downward with all reporting constraints explicitly shown natively rendering recursive chart trees cleanly.*

---

## 4. Complex Identity Flows: Onboarding (Phase 14)

Instead of the frontend making 4 sequential calls mapping Users, Employees, Roles, and Logs manually – the backend natively exposes an atomic transactional route to safely onboard users in exactly 1 call natively isolated.

**Endpoint:** `POST /onboard` (Requires `ADMIN` Role)

**Request Schema:**
```json
{
  "employeeNo": "UK-00001",
  "firstName": "Hemish",
  "lastName": "Patel",
  "displayName": "Hemish Director",
  "workEmail": "hemish.patel@inova.krd",
  "password": "Testing123!",
  "initialRoleId": "b2f6d2ea-...",
  "businessUnitId": "c499b...",
  "departmentId": "d027a...",
  "jobTitleId": "e11ba...",
  "managerId": "" 
}
```
*Tip: Any UUID mapping not known immediately can be securely submitted as an empty string `""` natively interpreting as a PostgreSQL NULL pointer locally preserving bounds constraints.*

**Response (201 Created):**
```json
{
  "employeeId": "...",
  "userId": "...",
  "email": "hemish.patel@inova.krd"
}
```
*Note: Upon successful creation, an Audit record is instantly streamed asynchronously tracing your actor bounds into the audit table.*

---

## 5. Audit Logistics

QMS environments strongly strictly mandate auditing capabilities directly isolated natively.

**Endpoint:** `GET /audit-logs?entityType=Employees` (Requires `ADMIN`)

**Response Schema:**
```json
{
  "data": [
    {
      "id": "...",
      "tenantId": "...",
      "actorId": "...",
      "action": "CREATE",
      "entityType": "Employees",
      "entityId": "...",
      "changes": {
        "employee_no": "UK-00001",
        "first_name": "Hemish",
        "last_name": "Patel"
      },
      "createdAt": "2026-02-20T21:00:00Z"
    }
  ]
}
```
*Enjoy interfacing with the API securely! Check the swagger JSON configuration natively inside `docs/swagger.json` if using Postman environments for mapping endpoints.*
