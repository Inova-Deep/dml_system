# DML Application - Future Development Roadmap

This document outlines the architectural phases, API capabilities, and system modules that were explicitly scoped out of the V1 (Minimum Viable Product) release. These items should be prioritized in subsequent development sprints.

## 1. Full CRUD Write Implementation (`PUT` / `DELETE`)

During V1, the core entity domains were heavily optimized for `GET` (Read) and `POST` (Create) operations alongside complex hierarchical queries and massive atomic transactions (e.g., Staff Onboarding). Standard `PUT` (Update) and `DELETE` (Remove) handlers were deferred.

**Target Entities Requiring `PUT` and `DELETE` Handlers:**
*   Business Units (`/api/v1/business-units/{id}`)
*   Departments (`/api/v1/departments/{id}`)
*   Job Titles (`/api/v1/job-titles/{id}`)
*   Employees (`/api/v1/employees/{id}`)
*   Users (`/api/v1/users/{id}`)
*   Roles (`/api/v1/roles/{id}`) *(Note: Role Revocation `DELETE` is already implemented, but deleting the `Role` definition itself is deferred)*

**Implementation Steps (Next Sprint):**
1.  **SQL Layer:** Add `UPDATE` and `DELETE` commands to `sqlc/queries.sql`. Ensure all queries strictly enforce `tenant_id` boundaries. Implement soft-deletes (`deleted_at` timestamp) versus hard-deletes based on business compliance requirements.
2.  **Code Generation:** Run `sqlc generate` to map the new queries to standard Go structs in `internal/domain`.
3.  **Application Logic:** Expand the domain services (e.g., `EmployeeService`, `OrganizationService`) with robust update and removal logic, ensuring Audit Logs are emitted appropriately for these modified actions.
4.  **HTTP Handlers:** Introduce the `PUT` and `DELETE` routers securely behind the `RequireRole("ADMIN")` authorization middleware.
5.  **Documentation:** Annotate the new verbs with Swaggo declarative comments and re-run the OpenAPI compilation.
6.  **Postman UI:** Expand the `dml_postman_collection.json` asset with the new dynamic payload bodies for the `PUT` scenarios.

## 2. Document Control System (DCS)

The Document Control System was identified in the initial PRD analysis but skipped during V1 to accelerate the core QMS foundation rollout.

**Planned Capabilities:**
*   **Version Control:** Full document versioning (Draft, Published, Archived) explicitly tied to the PostgreSQL relational bindings.
*   **Approval Workflows:** Multi-stage hierarchical approval chaining (leveraging the existing `manager_id` mappings on the Employee table).
*   **Audit Trailing:** Deep integration with the currently running `audit_logs` system to track reads, signatures, and alterations.
*   **Storage Provider Integration:** Binding the Go backend to an S3-compatible object storage provider (e.g., AWS S3, MinIO) for physical file retention securely linked to the Postgres metadata rows.
