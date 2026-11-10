# Company Asset & Attendance Management Platform (Golang)

A robust, lightweight, and secure **Company Asset & Attendance Management Platform** backend built with **Go (Golang)**, **Gin**, **GORM**, and **JWT-based Role-Based Access Control (RBAC)**.

The platform tracks company-owned physical and digital assets across their entire lifecycle, manages equipment allocations and requests, and provides end-to-end employee attendance tracking (daily check-in/check-out, working hours calculation, and leave management) across three distinct user roles: **IT**, **HR**, and **Employees**.

---

## Features & Highlights

- **Clean Layered Architecture:** Decoupled Handlers -> Services -> Repositories -> Models -> Middlewares.
- **Role-Based Access Control (RBAC):** JWT middleware enforcing granular permissions for `it`, `hr`, and `employee` roles.
- **Dual-Database Support:**
  - **SQLite (Default):** Zero-configuration pure-Go driver (`github.com/glebarez/sqlite`) running locally without CGO or GCC.
  - **PostgreSQL:** Production-grade PostgreSQL support via GORM and `pgx`.
- **Automatic Migration & Auto-Seeding:** Automatically creates tables and seeds demo IT, HR, and Employee users alongside sample hardware assets, attendance records, and leave requests.
- **Asset Lifecycle Management:** Transitions between `available`, `assigned`, `maintenance`, and `retired`.
- **Equipment Allocation & Returns:** Complete audit history tracking assigner, timestamps, return dates, and condition notes.
- **Daily Attendance Tracking:**
  - One-click daily check-in (`present`, `remote`, `half_day`) and check-out.
  - Automatic work hours calculation upon check-out.
  - Protection against duplicate check-ins or check-outs within the same calendar day.
  - Scoped personal attendance logs for employees.
  - Company-wide attendance reports with date and department filtering for HR.
- **Leave Management:**
  - Employees submit leave requests (`sick`, `casual`, `annual`, `unpaid`).
  - Input validation ensuring `end_date` is on or after `start_date`.
  - HR reviews and approves/rejects leave requests with optional review notes.
- **Centralized Error Handling:** Standardized JSON response envelopes with HTTP status codes (`400`, `401`, `403`, `404`, `409`, `500`).
- **Comprehensive Test Suite:** 100% passing integration tests covering authentication, RBAC boundaries, asset mutations, request approvals, attendance check-in/out, and leave workflows.

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
|    - Role-Based Access Control (RBAC: it, hr, employee)     |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                     Handlers / Controllers                  |
|   (Auth, Asset, Assignment, Request, Attendance, Leave)     |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                     Business Logic Services                 |
|   (Auth, Asset, Assignment, Request, Attendance, Leave)     |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                    Data Repositories (GORM)                 |
|   (User, Asset, Assignment, Request, Attendance, Leave)     |
+-------------------------------------------------------------+
                               |
                               v
+-------------------------------------------------------------+
|                     Database Engine                         |
|     SQLite (Pure Go / Local)  <-->  PostgreSQL (Prod)       |
+-------------------------------------------------------------+
```

---

## User Roles & Permissions Matrix

| Endpoint | Method | Description | IT | HR | Employee |
| :--- | :--- | :--- | :---: | :---: | :---: |
| **Authentication** | | | | | |
| `/api/v1/auth/register` | `POST` | Register a new user | Yes | Yes | Yes |
| `/api/v1/auth/login` | `POST` | Login & receive JWT | Yes | Yes | Yes |
| `/api/v1/auth/me` | `GET` | View authenticated profile | Yes | Yes | Yes |
| **Assets Module** | | | | | |
| `/api/v1/assets` | `GET` | List all assets with search/filters | **Yes** | **Yes** | No (403) |
| `/api/v1/assets/:id` | `GET` | View asset details & assignment history | **Yes** | **Yes** | No (403) |
| `/api/v1/assets` | `POST` | Create new hardware/digital asset | **Yes** | No (403) | No (403) |
| `/api/v1/assets/:id` | `PUT` | Update asset details / status | **Yes** | No (403) | No (403) |
| `/api/v1/assets/:id` | `DELETE` | Retire / delete asset | **Yes** | No (403) | No (403) |
| **Assignments & Requests** | | | | | |
| `/api/v1/assignments` | `POST` | Assign available asset to employee | **Yes** | No (403) | No (403) |
| `/api/v1/assignments/:id/return` | `PUT` | Mark asset as returned | **Yes** | **Yes** | No (403) |
| `/api/v1/assignments/my-assets` | `GET` | View personal assigned assets | Yes | Yes | **Yes** |
| `/api/v1/assignments` | `GET` | View all system assignments | **Yes** | **Yes** | No (403) |
| `/api/v1/requests` | `POST` | Submit hardware/accessory request | Yes | Yes | **Yes** |
| `/api/v1/requests` | `GET` | View requests (Employees see own only) | **All** | **All** | **Own Only** |
| `/api/v1/requests/:id/status`| `PUT` | Approve / Reject / Fulfill request | **Yes** | **Yes** | No (403) |
| **Attendance Module** | | | | | |
| `/api/v1/attendance/check-in` | `POST` | Daily check-in (present, remote, etc.) | Yes | Yes | **Yes** |
| `/api/v1/attendance/check-out` | `POST` | Daily check-out & compute work hours | Yes | Yes | **Yes** |
| `/api/v1/attendance/my-logs` | `GET` | Personal attendance history | Yes | Yes | **Yes** |
| `/api/v1/attendance/report` | `GET` | Company-wide report (date/dept filter) | No (403) | **Yes** | No (403) |
| **Leave Management** | | | | | |
| `/api/v1/leaves` | `POST` | Submit leave request (sick, casual, etc.) | Yes | Yes | **Yes** |
| `/api/v1/leaves` | `GET` | List leaves (HR: all, Employee: own) | No (403) | **All** | **Own Only** |
| `/api/v1/leaves/:id` | `GET` | View leave details | No (403) | **All** | **Own Only** |
| `/api/v1/leaves/:id/status` | `PUT` | Approve or reject leave request | No (403) | **Yes** | No (403) |
| **Reports & Organization** | | | | | |
| `/api/v1/users` | `GET` | List employees with active asset counts | **Yes** | **Yes** | No (403) |
| `/api/v1/users/:id/assets` | `GET` | View employee's allocated asset history | **Yes** | **Yes** | No (403) |
| `/api/v1/reports/inventory` | `GET` | Inventory counts, category & valuation | **Yes** | **Yes** | No (403) |
| `/api/v1/reports/audit-logs` | `GET` | System-wide audit log trail | **Yes** | No (403) | No (403) |

---

## Pre-seeded Demo Credentials

When running with `AUTO_SEED=true`, the following credentials are pre-configured:

| Role | Name | Email | Password | Pre-seeded State |
| :--- | :--- | :--- | :--- | :--- |
| **IT Admin** | Alex Turner | `it.admin@company.com` | `Admin@123` | Full IT administration |
| **HR Manager** | Sarah Connor | `hr.manager@company.com` | `HR@123` | Attendance reporting & leave approval |
| **Employee 1** | John Doe | `john.doe@company.com` | `Employee@123` | Assigned Laptop & Monitor; approved vacation leave |
| **Employee 2** | Jane Smith | `jane.smith@company.com` | `Employee@123` | Pending display request; pending sick leave |

---

## Getting Started

### 1. Configuration (`.env`)
```ini
SERVER_PORT=8080
ENVIRONMENT=development

# Database Settings (Default is pure-Go SQLite)
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

# Authentication
JWT_SECRET=super-secret-company-asset-jwt-key-2026!
JWT_EXPIRATION_HOURS=24

# Seed demo users, assets, attendance & leaves on startup
AUTO_SEED=true
```

### 2. Run Application
```bash
go run ./cmd/server
```
The server will start listening at `http://localhost:8080`.

---

## Running Automated Tests

Run the complete test suite:
```bash
go test -v ./tests/...
```

Test coverage includes:
- User registration, duplicate checks, login, and JWT claims validation.
- RBAC boundary enforcement across Employee, HR, and IT roles.
- Asset CRUD, serial number uniqueness, and lifecycle transitions.
- Asset assignment and offboarding return flows.
- Attendance check-in, duplicate checks, check-out, and work hour calculation.
- Leave request submission, date validation, and HR review workflows.
- Financial inventory valuations and audit trail logs.

---

## API Reference & Sample Requests

All responses follow standard envelopes:

```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### 1. Attendance Workflow

#### Daily Check-In (Employee)
```bash
curl -X POST http://localhost:8080/api/v1/attendance/check-in \
  -H "Authorization: Bearer <EMPLOYEE_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "present",
    "notes": "Checked in at office workstation"
  }'
```

#### Daily Check-Out (Employee)
```bash
curl -X POST http://localhost:8080/api/v1/attendance/check-out \
  -H "Authorization: Bearer <EMPLOYEE_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "notes": "Completed tasks for sprint 12"
  }'
```
*Note: The platform automatically computes `work_hours` as the difference between `check_out_time` and `check_in_time`.*

#### View Personal Attendance Logs (Employee)
```bash
curl -X GET "http://localhost:8080/api/v1/attendance/my-logs?page=1&limit=20" \
  -H "Authorization: Bearer <EMPLOYEE_TOKEN>"
```

#### Company-Wide Attendance Report (HR Only)
```bash
curl -X GET "http://localhost:8080/api/v1/attendance/report?department=Engineering&start_date=2026-11-01&end_date=2026-11-30" \
  -H "Authorization: Bearer <HR_TOKEN>"
```

---

### 2. Leave Management Workflow

#### Submit Leave Request (Employee)
```bash
curl -X POST http://localhost:8080/api/v1/leaves \
  -H "Authorization: Bearer <EMPLOYEE_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "leave_type": "casual",
    "start_date": "2026-11-20T00:00:00Z",
    "end_date": "2026-11-22T00:00:00Z",
    "reason": "Personal family travel"
  }'
```

#### List Leave Requests (HR: All / Employee: Own)
```bash
curl -X GET "http://localhost:8080/api/v1/leaves?status=pending" \
  -H "Authorization: Bearer <HR_TOKEN>"
```

#### Approve / Reject Leave Request (HR Only)
```bash
curl -X PUT http://localhost:8080/api/v1/leaves/<LEAVE_ID>/status \
  -H "Authorization: Bearer <HR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "review_notes": "Approved by HR. Coverage confirmed."
  }'
```

---

### 3. Asset Management Workflow

#### Create Asset (IT Only)
```bash
curl -X POST http://localhost:8080/api/v1/assets \
  -H "Authorization: Bearer <IT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "MacBook Pro 16 M3 Max",
    "serial_number": "MBP-2026-9901",
    "category": "Laptop",
    "status": "available",
    "purchase_date": "2026-01-15T00:00:00Z",
    "purchase_cost": 3499.00,
    "specifications": "Apple M3 Max, 64GB RAM, 2TB SSD"
  }'
```

#### Assign Asset to Employee (IT Only)
```bash
curl -X POST http://localhost:8080/api/v1/assignments \
  -H "Authorization: Bearer <IT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "asset_id": "<ASSET_ID>",
    "user_id": "<EMPLOYEE_ID>",
    "notes": "Issued during engineering onboarding"
  }'
```

#### Return Asset (IT / HR Offboarding)
```bash
curl -X PUT http://localhost:8080/api/v1/assignments/<ASSIGNMENT_ID>/return \
  -H "Authorization: Bearer <HR_OR_IT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "returned_status": "available",
    "notes": "Hardware inspected and returned to storage"
  }'
```

---

## Postman Collection

Import [`postman_collection.json`](file:///c:/asset-management-backend/postman_collection.json) into Postman to immediately access all pre-configured endpoints across Authentication, Assets, Assignments, Requests, Attendance, and Leave Management.

---

## License
MIT License. Open for company internal and commercial use.
