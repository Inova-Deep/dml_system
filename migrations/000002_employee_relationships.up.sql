ALTER TABLE employees
ADD COLUMN business_unit_id UUID REFERENCES business_units (id) ON DELETE SET NULL,
ADD COLUMN department_id UUID REFERENCES departments (id) ON DELETE SET NULL,
ADD COLUMN job_title_id UUID REFERENCES job_titles (id) ON DELETE SET NULL,
ADD COLUMN manager_id UUID REFERENCES employees (id) ON DELETE SET NULL;