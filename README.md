# Company Asset Management Platform (Golang)

A robust, lightweight, and secure **Company Asset Management Platform** backend built with **Go (Golang)**, **Gin**, **GORM**, and **JWT-based Role-Based Access Control (RBAC)**.

The system tracks physical and digital company assets (e.g., laptops, monitors, accessories, security keys) across their entire lifecycle, managing inventory, employee assignments, return workflows, and equipment requests across three distinct roles: **IT**, **HR**, and **Employees**.

---

## Features & Highlights

- **Clean Layered Architecture:** Decoupled Handlers, Services, Repositories, Models, and Middlewares.
- **Role-Based Access Control (RBAC):** JWT middleware enforcing granular permissions for `it`, `hr`, and `employee` roles.
- **Dual-Database Support:**
  - **SQLite (Default):** Zero-configuration pure-Go driver (`github.com/glebarez/sqlite`) running locally without CGO or GCC.
  - **PostgreSQL:** Production-grade PostgreSQL support via GORM and `pgx`.
- **Automatic Migration & Auto-Seeding:** Automatically creates tables and seeds demo IT, HR, and Employee users alongside sample hardware assets on first run.
- **Lifecycle Asset Tracking:** Transitions between `available`, `assigned`, `maintenance`, and `retired`.
- **Complete Assignment Audit History:** Tracks historical and active asset allocations, assigners, return dates, and condition notes.
- **HR Onboarding & Offboarding Visibility:** HR-specific endpoints to audit active assets per employee and facilitate returns during offboarding.
- **Employee Self-Service:** Employees can submit hardware requests and view their assigned assets.
- **Centralized Error Handling:** Consistent JSON response envelopes with standardized HTTP status codes (`400`, `401`, `403`, `404`, `409`, `500`).
- **Input Validation:** Strict validation using Go struct tags (`go-playground/validator`).
- **Comprehensive Test Suite:** 100% passing integration tests covering authentication, RBAC boundaries, asset mutations, request approvals, and reporting.

---

## Architecture Overview

```
+-------------------------------------------------------------+
|                      Client Requests                        |
|        (Web Frontend, Postman, Mobile, Automation)          |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                  HTTP Router & Middlewares                  |
|    - CORS & Logger                                          |
|    - Centralized Error / Panic Recovery                     |
|    - JWT Authentication Middleware                          |
|    - Role-Based Access Control (RBAC)                       |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                     Handlers / Controllers                  |
|    (AuthHandler, AssetHandler, AssignmentHandler, ...)      |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                     Business Logic Services                 |
|    (AuthService, AssetService, AssignmentService, ...)      |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                    Data Repositories (GORM)                 |
|   (UserRepository, AssetRepository, RequestRepository, ...) |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                     Database Engine                         |
|     SQLite (Pure Go / Local)  <-->  PostgreSQL (Prod)       |
+-------------------------------------------------------------+
```

---

## Project Structure

```
asset-management-backend/
├── cmd/
│   └── server/
│       └── main.go                 # Application bootstrap & graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go               # Environment configuration loader
│   ├── models/
│   │   ├── user.go                 # User entity & role definitions
│   │   ├── asset.go                # Asset entity, categories & statuses
│   │   ├── assignment.go           # AssetAssignment entity
│   │   ├── request.go              # AssetRequest entity & workflow statuses
│   │   └── audit_log.go            # AuditLog compliance entity
│   ├── repository/
│   │   ├── user_repository.go      # User data access
│   │   ├── asset_repository.go     # Asset querying, filters & inventory aggregations
│   │   ├── assignment_repository.go# Assignment history & active checks
│   │   ├── request_repository.go   # Request queries
│   │   └── audit_repository.go     # Audit log storage
│   ├── services/
│   │   ├── auth_service.go         # Authentication & token issuance
│   │   ├── asset_service.go        # Asset lifecycle & CRUD logic
│   │   ├── assignment_service.go   # Asset allocation & return logic
│   │   ├── request_service.go      # Request submission & review logic
│   │   └── report_service.go       # Valuation metrics & HR allocation summaries
│   ├── handlers/
│   │   ├── auth_handler.go         # /api/v1/auth endpoints
│   │   ├── asset_handler.go        # /api/v1/assets endpoints
│   │   ├── assignment_handler.go   # /api/v1/assignments endpoints
│   │   ├── request_handler.go      # /api/v1/requests endpoints
│   │   ├── user_handler.go         # /api/v1/users endpoints
│   │   └── report_handler.go       # /api/v1/reports endpoints
│   ├── middleware/
│   │   ├── auth_middleware.go      # JWT token verification
│   │   ├── rbac_middleware.go      # Role restriction enforcement
│   │   ├── error_middleware.go     # Centralized error handler & panic recovery
│   │   └── cors_middleware.go      # Cross-origin headers
│   └── router/
│       └── router.go               # Gin route registrations & grouping
├── pkg/
│   ├── database/
│   │   ├── db.go                   # GORM DB connection (SQLite & PostgreSQL)
│   │   ├── migrate.go              # Schema AutoMigrate
│   │   └── seed.go                 # Demo users & initial hardware assets
│   └── utils/
│       ├── errors.go               # AppError standard constructors
│       ├── jwt.go                  # JWT claim generation & parsing
│       ├── password.go             # Bcrypt hashing & verification
│       └── response.go             # Standardized JSON envelopes
├── tests/
│   ├── test_helper.go              # In-memory DB setup & request simulator
│   ├── auth_test.go                # Auth & registration tests
│   ├── asset_test.go               # Asset CRUD, uniqueness & RBAC tests
│   ├── assignment_test.go          # Assignment & return workflow tests
│   ├── request_test.go             # Asset request & approval tests
│   └── report_test.go              # Inventory metrics & HR audit tests
├── postman_collection.json         # Ready-to-import Postman API collection
├── .env.example                    # Environment configuration template
├── go.mod                          # Go module dependencies
└── README.md                       # Complete documentation
```

---

## User Roles & Permissions Matrix

| Endpoint | Method | Description | IT | HR | Employee |
| :--- | :--- | :--- | :---: | :---: | :---: |
| `/api/v1/auth/register` | `POST` | Register a new user | Yes | Yes | Yes |
| `/api/v1/auth/login` | `POST` | Login & receive JWT | Yes | Yes | Yes |
| `/api/v1/auth/me` | `GET` | View authenticated profile | Yes | Yes | Yes |
| `/api/v1/assets` | `GET` | List all assets with search/filters | **Yes** | **Yes** | No (403) |
| `/api/v1/assets/:id` | `GET` | View asset details & assignment history | **Yes** | **Yes** | No (403) |
| `/api/v1/assets` | `POST` | Create new hardware/digital asset | **Yes** | No (403) | No (403) |
| `/api/v1/assets/:id` | `PUT` | Update asset details / status | **Yes** | No (403) | No (403) |
| `/api/v1/assets/:id` | `DELETE` | Retire / delete asset | **Yes** | No (403) | No (403) |
| `/api/v1/assignments` | `POST` | Assign available asset to employee | **Yes** | No (403) | No (403) |
| `/api/v1/assignments/:id/return` | `PUT` | Mark asset as returned | **Yes** | **Yes** | No (403) |
| `/api/v1/assignments/my-assets` | `GET` | View current and past assigned assets | Yes | Yes | **Yes** |
| `/api/v1/assignments` | `GET` | View all system assignments | **Yes** | **Yes** | No (403) |
| `/api/v1/requests` | `POST` | Submit hardware/accessory request | Yes | Yes | **Yes** |
| `/api/v1/requests` | `GET` | View requests (Employees see own only) | **All** | **All** | **Own Only** |
| `/api/v1/requests/:id/status`| `PUT` | Approve / Reject / Fulfill request | **Yes** | **Yes** | No (403) |
| `/api/v1/users` | `GET` | List employees with active asset counts | **Yes** | **Yes** | No (403) |
| `/api/v1/users/:id/assets` | `GET` | View employee's allocated asset history | **Yes** | **Yes** | No (403) |
| `/api/v1/reports/inventory` | `GET` | Inventory counts, category & valuation | **Yes** | **Yes** | No (403) |
| `/api/v1/reports/audit-logs` | `GET` | System-wide audit log trail | **Yes** | No (403) | No (403) |

---

## Pre-seeded Demo Credentials

When running with `AUTO_SEED=true`, the following credentials are pre-configured:

| Role | Name | Email | Password | Pre-seeded State |
| :--- | :--- | :--- | :--- | :--- |
| **IT Admin** | Alex Turner | `it.admin@company.com` | `Admin@123` | Full admin privileges |
| **HR Manager** | Sarah Connor | `hr.manager@company.com` | `HR@123` | Employee offboarding & asset return access |
| **Employee 1** | John Doe | `john.doe@company.com` | `Employee@123` | Has 2 active assets (Dell XPS 15 & 4K Monitor) |
| **Employee 2** | Jane Smith | `jane.smith@company.com` | `Employee@123` | Has 1 pending request for Apple Studio Display |

---

## Getting Started

### Prerequisites
- **Go 1.22+** (tested with Go 1.26)
- Git

### 1. Clone & Setup Configuration
```bash
git clone <repository-url>
cd asset-management-backend

# Copy default environment settings
cp .env.example .env
```

### 2. Environment Variables (`.env`)
```ini
SERVER_PORT=8080
ENVIRONMENT=development

# Database Configuration (Default is pure-Go SQLite; no setup needed)
DB_DRIVER=sqlite
DB_PATH=./asset_management.db

# If switching to PostgreSQL:
# DB_DRIVER=postgres
# DB_HOST=localhost
# DB_PORT=5432
# DB_USER=postgres
# DB_PASSWORD=postgres
# DB_NAME=asset_management
# DB_SSLMODE=disable

# Security & JWT
JWT_SECRET=super-secret-company-asset-jwt-key-2026!
JWT_EXPIRATION_HOURS=24

# Seed demo users & assets on startup
AUTO_SEED=true
```

### 3. Run the Server
```bash
go run ./cmd/server
```
The server will start listening at `http://localhost:8080`.

---

## Running Automated Tests

Run the comprehensive test suite:
```bash
go test -v ./tests/...
```

All tests execute against an isolated in-memory SQLite database and test:
- User registration, duplicate detection, and login with bcrypt hashing.
- Role-Based Access Control (ensuring employees and HR cannot execute unauthorized actions).
- Asset creation, serial number uniqueness constraints, and soft deletion.
- Asset assignment lifecycle (state transitions from `available` -> `assigned` -> `available` on return).
- Employee request submissions, list scoping, and IT/HR approval transitions.
- Financial inventory valuations and audit trail logs.

---

## API Reference & Sample Requests

All API responses follow standard JSON structures:

**Success Response Envelope:**
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```

**Paginated Response Envelope:**
```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 45,
    "total_pages": 5
  }
}
```

**Error Response Envelope:**
```json
{
  "success": false,
  "message": "Error description",
  "error": {
    "code": 400,
    "details": "Validation details or context"
  }
}
```

---

### Workflow Examples with `curl`

#### 1. Authentication

**Login as IT Admin:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "it.admin@company.com",
    "password": "Admin@123"
  }'
```

**Save Token in environment variable:**
```bash
TOKEN="<JWT_TOKEN_FROM_LOGIN>"
```

**Get Current User Profile:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

---

#### 2. Asset Management (IT Only)

**Add a New Asset:**
```bash
curl -X POST http://localhost:8080/api/v1/assets \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MacBook Pro 16 M3 Max",
    "serial_number": "MBP-2026-9901",
    "category": "Laptop",
    "status": "available",
    "purchase_date": "2026-01-15T00:00:00Z",
    "purchase_cost": 3499.00,
    "specifications": "Apple M3 Max 16-core CPU, 40-core GPU, 64GB Unified RAM, 2TB SSD"
  }'
```

**List Assets with Filter & Pagination (IT & HR):**
```bash
curl -X GET "http://localhost:8080/api/v1/assets?category=Laptop&status=available&page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Get Asset Details with Assignment History:**
```bash
curl -X GET http://localhost:8080/api/v1/assets/<ASSET_ID> \
  -H "Authorization: Bearer $TOKEN"
```

**Update Asset:**
```bash
curl -X PUT http://localhost:8080/api/v1/assets/<ASSET_ID> \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MacBook Pro 16 M3 Max (Engineering Stock)",
    "status": "available"
  }'
```

**Retire / Delete Asset:**
```bash
curl -X DELETE http://localhost:8080/api/v1/assets/<ASSET_ID> \
  -H "Authorization: Bearer $TOKEN"
```

---

#### 3. Asset Assignment Lifecycle

**Assign Asset to Employee (IT Only):**
```bash
curl -X POST http://localhost:8080/api/v1/assignments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "asset_id": "<ASSET_ID>",
    "user_id": "<EMPLOYEE_USER_ID>",
    "notes": "Issued during engineering onboarding; desk setup bundle"
  }'
```
*Note: This automatically transitions the asset's status from `available` to `assigned`.*

**Employee Viewing Their Assigned Assets:**
```bash
curl -X GET http://localhost:8080/api/v1/assignments/my-assets \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN"
```

**Mark Asset as Returned (IT & HR):**
```bash
curl -X PUT http://localhost:8080/api/v1/assignments/<ASSIGNMENT_ID>/return \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "returned_status": "available",
    "notes": "Hardware inspected, wiped, and placed back in main inventory"
  }'
```
*Note: This timestamps `returned_at` and automatically switches the asset status back to `available`.*

---

#### 4. Asset Requests Workflow

**Employee Submitting Request:**
```bash
curl -X POST http://localhost:8080/api/v1/requests \
  -H "Authorization: Bearer $EMPLOYEE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "asset_category": "Accessory",
    "item_requested": "Keychron Q1 Pro Mechanical Keyboard",
    "reason": "Ergonomic keyboard for coding workflow"
  }'
```

**List Requests (IT & HR see all; Employees see own):**
```bash
curl -X GET "http://localhost:8080/api/v1/requests?status=pending" \
  -H "Authorization: Bearer $TOKEN"
```

**IT/HR Approving Request:**
```bash
curl -X PUT http://localhost:8080/api/v1/requests/<REQUEST_ID>/status \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "review_notes": "Approved by IT Lead. Dispatched from IT storage."
  }'
```

---

#### 5. HR Offboarding & Employee Inventory

**HR List Employees with Current Allocated Asset Counts:**
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $HR_TOKEN"
```

**HR View All Allocated Assets for Specific Employee:**
```bash
curl -X GET http://localhost:8080/api/v1/users/<EMPLOYEE_USER_ID>/assets \
  -H "Authorization: Bearer $HR_TOKEN"
```

---

#### 6. Reports & Audit Trail

**Inventory Valuation & Breakdown (IT & HR):**
```bash
curl -X GET http://localhost:8080/api/v1/reports/inventory \
  -H "Authorization: Bearer $TOKEN"
```
Sample response:
```json
{
  "success": true,
  "message": "inventory summary retrieved successfully",
  "data": {
    "total_assets": 7,
    "total_valuation": 8229.00,
    "by_status": {
      "available": 4,
      "assigned": 2,
      "maintenance": 1
    },
    "by_category": {
      "Laptop": 2,
      "Monitor": 2,
      "Accessory": 2,
      "Security": 1
    }
  }
}
```

**System Audit Trail (IT Only):**
```bash
curl -X GET "http://localhost:8080/api/v1/reports/audit-logs?page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Postman Collection

Import `postman_collection.json` into Postman to immediately access pre-configured requests with environment variables for `baseUrl` and `token`.

---

## License
MIT License. Open for company internal and commercial use.
