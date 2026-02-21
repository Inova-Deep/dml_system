-- name: GetTenant :one
SELECT * FROM tenants WHERE id = $1 LIMIT 1;

-- name: ListTenants :many
SELECT * FROM tenants ORDER BY name;

-- name: CreateTenant :one
INSERT INTO
    tenants (id, code, name)
VALUES ($1, $2, $3)
RETURNING
    *;

-- name: GetUserForLogin :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE tenant_id = $1 AND email = $2 LIMIT 1;

-- name: GetUser :one
SELECT * FROM users WHERE tenant_id = $1 AND id = $2 LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM users
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR email ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR display_name ILIKE '%' || sqlc.arg ('search')::text || '%'
    )
ORDER BY email
LIMIT sqlc.arg ('limit')
OFFSET
    sqlc.arg ('offset');

-- name: CountUsers :one
SELECT count(*)
FROM users
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR email ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR display_name ILIKE '%' || sqlc.arg ('search')::text || '%'
    );

-- name: CreateUser :one
INSERT INTO
    users (
        id,
        tenant_id,
        employee_id,
        email,
        display_name,
        password_hash
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    *;

-- name: GetEmployee :one
SELECT * FROM employees WHERE tenant_id = $1 AND id = $2 LIMIT 1;

-- name: GetEmployeeWithDetails :one
SELECT
    e.id,
    e.tenant_id,
    e.employee_no,
    e.first_name,
    e.last_name,
    e.display_name,
    e.work_email,
    e.status,
    e.is_active,
    e.created_at,
    e.updated_at,
    e.business_unit_id,
    e.department_id,
    e.job_title_id,
    e.manager_id,
    bu.code AS business_unit_code,
    bu.name AS business_unit_name,
    d.code AS department_code,
    d.name AS department_name,
    jt.code AS job_title_code,
    jt.name AS job_title_name,
    jt.grade AS job_title_grade,
    m.employee_no AS manager_employee_no,
    m.first_name AS manager_first_name,
    m.last_name AS manager_last_name,
    m.display_name AS manager_display_name
FROM employees e
LEFT JOIN business_units bu ON e.business_unit_id = bu.id AND e.tenant_id = bu.tenant_id
LEFT JOIN departments d ON e.department_id = d.id AND e.tenant_id = d.tenant_id
LEFT JOIN job_titles jt ON e.job_title_id = jt.id AND e.tenant_id = jt.tenant_id
LEFT JOIN employees m ON e.manager_id = m.id AND e.tenant_id = m.tenant_id
WHERE e.tenant_id = $1 AND e.id = $2 LIMIT 1;

-- name: ListEmployees :many
SELECT *
FROM employees
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR first_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR last_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR display_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR work_email ILIKE '%' || sqlc.arg ('search')::text || '%'
    )
ORDER BY last_name, first_name
LIMIT sqlc.arg ('limit')
OFFSET
    sqlc.arg ('offset');

-- name: ListEmployeesWithDetails :many
SELECT
    e.id,
    e.tenant_id,
    e.employee_no,
    e.first_name,
    e.last_name,
    e.display_name,
    e.work_email,
    e.status,
    e.is_active,
    e.created_at,
    e.updated_at,
    e.business_unit_id,
    e.department_id,
    e.job_title_id,
    e.manager_id,
    bu.code AS business_unit_code,
    bu.name AS business_unit_name,
    d.code AS department_code,
    d.name AS department_name,
    jt.code AS job_title_code,
    jt.name AS job_title_name,
    jt.grade AS job_title_grade,
    m.employee_no AS manager_employee_no,
    m.first_name AS manager_first_name,
    m.last_name AS manager_last_name,
    m.display_name AS manager_display_name
FROM employees e
LEFT JOIN business_units bu ON e.business_unit_id = bu.id AND e.tenant_id = bu.tenant_id
LEFT JOIN departments d ON e.department_id = d.id AND e.tenant_id = d.tenant_id
LEFT JOIN job_titles jt ON e.job_title_id = jt.id AND e.tenant_id = jt.tenant_id
LEFT JOIN employees m ON e.manager_id = m.id AND e.tenant_id = m.tenant_id
WHERE
    e.tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR e.first_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR e.last_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR e.display_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR e.work_email ILIKE '%' || sqlc.arg ('search')::text || '%'
    )
ORDER BY e.last_name, e.first_name
LIMIT sqlc.arg ('limit')
OFFSET
    sqlc.arg ('offset');

-- name: CountEmployees :one
SELECT count(*)
FROM employees
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR first_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR last_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR display_name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR work_email ILIKE '%' || sqlc.arg ('search')::text || '%'
    );

-- name: CreateEmployee :one
INSERT INTO
    employees (
        id,
        tenant_id,
        employee_no,
        first_name,
        last_name,
        display_name,
        work_email,
        business_unit_id,
        department_id,
        job_title_id,
        manager_id
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11
    )
RETURNING
    *;

-- name: GetBusinessUnit :one
SELECT *
FROM business_units
WHERE
    tenant_id = $1
    AND id = $2
LIMIT 1;

-- name: ListBusinessUnits :many
SELECT *
FROM business_units
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR code ILIKE '%' || sqlc.arg ('search')::text || '%'
    )
ORDER BY name
LIMIT sqlc.arg ('limit')
OFFSET
    sqlc.arg ('offset');

-- name: CountBusinessUnits :one
SELECT count(*)
FROM business_units
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR code ILIKE '%' || sqlc.arg ('search')::text || '%'
    );

-- name: CreateBusinessUnit :one
INSERT INTO
    business_units (id, tenant_id, code, name)
VALUES ($1, $2, $3, $4)
RETURNING
    *;

-- name: GetDepartment :one
SELECT * FROM departments WHERE tenant_id = $1 AND id = $2 LIMIT 1;

-- name: ListDepartments :many
SELECT *
FROM departments
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR code ILIKE '%' || sqlc.arg ('search')::text || '%'
    )
ORDER BY name
LIMIT sqlc.arg ('limit')
OFFSET
    sqlc.arg ('offset');

-- name: CountDepartments :one
SELECT count(*)
FROM departments
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR code ILIKE '%' || sqlc.arg ('search')::text || '%'
    );

-- name: CreateDepartment :one
INSERT INTO
    departments (
        id,
        tenant_id,
        parent_department_id,
        code,
        name
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetJobTitle :one
SELECT * FROM job_titles WHERE tenant_id = $1 AND id = $2 LIMIT 1;

-- name: ListJobTitles :many
SELECT *
FROM job_titles
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR code ILIKE '%' || sqlc.arg ('search')::text || '%'
    )
ORDER BY name
LIMIT sqlc.arg ('limit')
OFFSET
    sqlc.arg ('offset');

-- name: CountJobTitles :one
SELECT count(*)
FROM job_titles
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('search')::text = ''
        OR name ILIKE '%' || sqlc.arg ('search')::text || '%'
        OR code ILIKE '%' || sqlc.arg ('search')::text || '%'
    );

-- name: CreateJobTitle :one
INSERT INTO
    job_titles (
        id,
        tenant_id,
        code,
        name,
        grade
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetRole :one
SELECT * FROM rbac_roles WHERE tenant_id = $1 AND id = $2 LIMIT 1;

-- name: ListRoles :many
SELECT * FROM rbac_roles WHERE tenant_id = $1 ORDER BY name;

-- name: CreateRole :one
INSERT INTO
    rbac_roles (
        id,
        tenant_id,
        code,
        name,
        description
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING
    *;

-- name: GetUserRoles :many
SELECT r.code
FROM
    user_rbac_roles ur
    JOIN rbac_roles r ON ur.role_id = r.id
    AND ur.tenant_id = r.tenant_id
WHERE
    ur.tenant_id = $1
    AND ur.user_id = $2;

-- name: AssignUserRole :exec
INSERT INTO
    user_rbac_roles (
        tenant_id,
        user_id,
        role_id,
        business_unit_id,
        department_id,
        granted_by_user_id
    )
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (
    tenant_id,
    user_id,
    role_id,
    business_unit_id,
    department_id
) DO NOTHING;

-- name: RevokeUserRole :exec
DELETE FROM user_rbac_roles
WHERE
    tenant_id = $1
    AND user_id = $2
    AND role_id = $3;

-- name: GetEmployeeHierarchy :many
WITH RECURSIVE
    employee_tree AS (
        SELECT *
        FROM employees e1
        WHERE
            e1.tenant_id = $1
            AND e1.id = $2
        UNION ALL
        SELECT e2.*
        FROM
            employees e2
            INNER JOIN employee_tree et ON e2.manager_id = et.id
        WHERE
            e2.tenant_id = $1
    )
SELECT *
FROM employee_tree;

-- name: InsertAuditLog :one
INSERT INTO
    audit_logs (
        id,
        tenant_id,
        actor_id,
        action,
        entity_type,
        entity_id,
        changes
    )
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
    *;

-- name: ListAuditLogs :many
SELECT *
FROM audit_logs
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('entity_type')::text = ''
        OR entity_type = sqlc.arg ('entity_type')::text
    )
    AND (
        sqlc.arg ('action')::text = ''
        OR action = sqlc.arg ('action')::text
    )
ORDER BY created_at DESC
LIMIT sqlc.arg ('limit')
OFFSET
    sqlc.arg ('offset');

-- name: CountAuditLogs :one
SELECT count(*)
FROM audit_logs
WHERE
    tenant_id = $1
    AND (
        sqlc.arg ('entity_type')::text = ''
        OR entity_type = sqlc.arg ('entity_type')::text
    )
    AND (
        sqlc.arg ('action')::text = ''
        OR action = sqlc.arg ('action')::text
    );