-- Tenancy (future-proof; demo can be a single tenant row)
create table tenants (
    id uuid primary key,
    code text not null unique,
    name text not null,
    created_at timestamptz not null default now()
);

-- 1) Business Units = Sites
create table business_units (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    code text,
    name text not null, -- HQ Office / Factory London / Factory Bristol
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, code),
    unique (tenant_id, name)
);

-- 2) Business Lines (global list)
create table business_lines (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    code text,
    name text not null,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, code),
    unique (tenant_id, name)
);

-- 3) Departments (global list reused across sites)
create table departments (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    parent_department_id uuid null references departments (id),
    code text,
    name text not null,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, code),
    unique (tenant_id, name)
);

-- 4) Job Titles
create table job_titles (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    code text,
    name text not null,
    grade text,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, code),
    unique (tenant_id, name)
);

-- 5) Employees (HR identity)
create table employees (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    employee_no text not null, -- unique per tenant
    first_name text not null,
    last_name text not null,
    display_name text,
    work_email text, -- optional (users.email is required)
    status text not null default 'active',
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, employee_no),
    unique (tenant_id, work_email)
);

-- 6) Employee Assignments (one current primary only)
create table employee_assignments (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    employee_id uuid not null references employees (id),
    business_unit_id uuid not null references business_units (id),
    department_id uuid not null references departments (id),
    business_line_id uuid null references business_lines (id),
    job_title_id uuid null references job_titles (id),
    manager_employee_id uuid null references employees (id),
    effective_from date not null,
    effective_to date null, -- null = current
    is_primary boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- Enforce: exactly one CURRENT primary assignment per employee
create unique index uq_employee_one_current_assignment on employee_assignments (tenant_id, employee_id)
where
    effective_to is null
    and is_primary = true;

create index idx_employee_assignments_employee on employee_assignments (tenant_id, employee_id);

-- 7) Users (mandatory 1:1 with employees; login = email)
create table users (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    employee_id uuid not null references employees (id) on delete restrict,
    email text not null,
    display_name text,
    password_hash text,
    is_active boolean not null default true,
    last_login_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, email),
    unique (tenant_id, employee_id)
);

-- 8) RBAC roles (App roles)
create table rbac_roles (
    id uuid primary key,
    tenant_id uuid not null references tenants (id),
    code text not null, -- ADMIN, QA_MANAGER, VIEWER, ...
    name text not null,
    description text,
    is_active boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (tenant_id, code)
);

create table user_rbac_roles (
  tenant_id uuid not null references tenants(id),
  user_id uuid not null references users(id),
  role_id uuid not null references rbac_roles(id),

-- keep nullable; useful later without changing the model

business_unit_id uuid null references business_units(id),
  department_id uuid null references departments(id),

  granted_at timestamptz not null default now(),
  granted_by_user_id uuid null references users(id),

  primary key (tenant_id, user_id, role_id, business_unit_id, department_id)
);

create index idx_user_roles_user on user_rbac_roles (tenant_id, user_id);