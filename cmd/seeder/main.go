package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/INOVA/DML/internal/config"
	"github.com/INOVA/DML/internal/db"
	"github.com/INOVA/DML/internal/logic/audit"
	"github.com/INOVA/DML/internal/logic/hr"
	"github.com/INOVA/DML/internal/logic/iam"
	"github.com/INOVA/DML/internal/logic/org"
	"github.com/INOVA/DML/internal/logic/tenancy"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

func parseUUID(s string) pgtype.UUID {
	var id pgtype.UUID
	uid, err := uuid.Parse(s)
	if err == nil {
		id.Bytes = uid
		id.Valid = true
	}
	return id
}

func parseUUIDPtr(s string) *pgtype.UUID {
	if s == "" {
		return nil
	}
	id := parseUUID(s)
	return &id
}

var firstNames = []string{
	"James", "Mary", "Robert", "Patricia", "John", "Jennifer", "Michael", "Linda", "David", "Elizabeth",
	"William", "Barbara", "Richard", "Susan", "Joseph", "Jessica", "Thomas", "Sarah", "Charles", "Karen",
	"Christopher", "Lisa", "Daniel", "Nancy", "Matthew", "Betty", "Anthony", "Margaret", "Mark", "Sandra",
	"Donald", "Ashley", "Steven", "Kimberly", "Paul", "Emily", "Andrew", "Donna", "Joshua", "Michelle",
	"Kenneth", "Carol", "Kevin", "Amanda", "Brian", "Dorothy", "George", "Melissa", "Timothy", "Deborah",
	"Ronald", "Stephanie", "Edward", "Rebecca", "Jason", "Sharon", "Jeffrey", "Laura", "Ryan", "Cynthia",
	"Jacob", "Kathleen", "Gary", "Amy", "Nicholas", "Angela", "Eric", "Shirley", "Jonathan", "Anna",
	"Stephen", "Brenda", "Larry", "Pamela", "Justin", "Emma", "Scott", "Nicole", "Brandon", "Helen",
	"Benjamin", "Samantha", "Samuel", "Katherine", "Gregory", "Christine", "Alexander", "Debra", "Frank", "Rachel",
	"Patrick", "Carolyn", "Raymond", "Janet", "Jack", "Catherine", "Dennis", "Maria", "Jerry", "Heather",
}
var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
	"Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin",
	"Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson",
	"Walker", "Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
	"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell", "Carter", "Roberts",
	"Gomez", "Phillips", "Evans", "Turner", "Diaz", "Parker", "Cruz", "Edwards", "Collins", "Reyes",
	"Stewart", "Morris", "Morales", "Murphy", "Cook", "Rogers", "Gutierrez", "Ortiz", "Morgan", "Cooper",
	"Peterson", "Bailey", "Reed", "Kelly", "Howard", "Ramos", "Kim", "Cox", "Ward", "Richardson",
	"Watson", "Brooks", "Chavez", "Wood", "James", "Bennet", "Gray", "Mendoza", "Ruiz", "Hughes",
	"Price", "Alvarez", "Castillo", "Sanders", "Patel", "Myers", "Long", "Ross", "Foster", "Jimenez",
}

func randomName() (string, string) {
	return firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))]
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; executing with system environments.")
	}

	cfg := config.Load()

	ctx := context.Background()
	database, err := db.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed resolving PSQL driver: %v", err)
	}
	defer database.Close()

	log.Println("Starting Massive Database Seeder for UK Operations...")

	auditSvc := audit.NewAuditService(database)
	_ = tenancy.NewService(database)
	orgSvc := org.NewBusinessUnitService(database)
	deptSvc := org.NewDepartmentService(database)
	jobSvc := org.NewJobTitleService(database)
	roleSvc := iam.NewRoleService(database, auditSvc)
	onboardSvc := hr.NewOnboardingService(database, auditSvc)

	// --- 1. Tenants & System Account (Get or Create) ---
	var tenant1ID, sysUserUUID pgtype.UUID

	// Check if tenant exists
	err = database.Pool.QueryRow(ctx, "SELECT id FROM tenants WHERE code = 'TEN-UK-001'").Scan(&tenant1ID)
	if err != nil {
		log.Println(">> Tenant 'TEN-UK-001' not found. Initializing from scratch...")
		tenantSvc := tenancy.NewService(database)
		tempSysUUID := parseUUID(uuid.New().String())
		t1, err := tenantSvc.CreateTenant(ctx, tempSysUUID, "TEN-UK-001", "Nova Systems UK Ltd")
		if err != nil {
			log.Fatalf("Fatal: Failed to create base tenant: %v", err)
		}
		tenant1ID = t1.ID

		// Create secondary isolation tenant
		_, _ = tenantSvc.CreateTenant(ctx, tempSysUUID, "TEN-UK-002", "Nova Consulting Services UK")
	}

	// Check if system user exists
	err = database.Pool.QueryRow(ctx, "SELECT id FROM users WHERE email = 'system@inova.krd' OR email = 'system@nova.local'").Scan(&sysUserUUID)
	if err != nil {
		log.Println(">> System account not found. Initializing...")
		sysUserUUID = parseUUID(uuid.New().String())
		// We'll insert it later inside the Business Unit block as per original script flow
	}

	// Wrapper to fix downstream code
	type mockEntity struct{ ID pgtype.UUID }
	tenant1 := mockEntity{ID: tenant1ID}
	tenant2 := mockEntity{ID: parseUUID(uuid.New().String())} // Placeholder for compiler

	log.Printf(">> Using Primary Tenant: %s", uuid.UUID(tenant1.ID.Bytes).String())

	// --- 3. Business Units ---
	buLondon, _ := orgSvc.CreateBusinessUnit(ctx, parseUUID(uuid.New().String()), tenant1.ID, "LON-HQ", "London Headquarters")
	buMan, _ := orgSvc.CreateBusinessUnit(ctx, parseUUID(uuid.New().String()), tenant1.ID, "MAN-OPS", "Manchester Operations")
	buEdin, _ := orgSvc.CreateBusinessUnit(ctx, parseUUID(uuid.New().String()), tenant1.ID, "EDI-TECH", "Edinburgh Tech Hub")

	database.Pool.Exec(ctx, "INSERT INTO employees (id, tenant_id, employee_no, first_name, last_name, business_unit_id) VALUES ($1, $2, 'SYS', 'System', 'Admin', $3)", sysUserUUID, tenant1.ID, buLondon.ID)
	database.Pool.Exec(ctx, "INSERT INTO users (id, tenant_id, employee_id, email, display_name) VALUES ($1, $2, $1, 'system@nova.local', 'SYSTEM_ACCOUNT')", sysUserUUID, tenant1.ID)

	// Super Admin Role limits executing seeder queries natively mapping the system ID seamlessly
	// --- 2. Roles ---
	adminRole, _ := roleSvc.CreateRole(ctx, parseUUID(uuid.New().String()), tenant1.ID, sysUserUUID, "SYSTEM_ADMIN", "Super Administrator", nil)
	hrRole, _ := roleSvc.CreateRole(ctx, parseUUID(uuid.New().String()), tenant1.ID, sysUserUUID, "HR_ADMIN", "Human Resources Admin", nil)
	mgrRole, _ := roleSvc.CreateRole(ctx, parseUUID(uuid.New().String()), tenant1.ID, sysUserUUID, "DEPT_MANAGER", "Departmental Manager", nil)
	empRole, _ := roleSvc.CreateRole(ctx, parseUUID(uuid.New().String()), tenant1.ID, sysUserUUID, "EMPLOYEE", "Standard Employee", nil)

	// Create roles for Tenant 2 to ensure no overlap violations
	roleSvc.CreateRole(ctx, parseUUID(uuid.New().String()), tenant2.ID, sysUserUUID, "SYSTEM_ADMIN", "Super Administrator", nil)

	log.Printf(">> Loaded 4 Base Roles.")

	// --- 4. Departments ---
	deptExe, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "EXE", "Executive Board")
	deptIT, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "IT", "Information Technology")
	deptHR, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "HR", "Human Resources")
	deptFin, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "FIN", "Finance & Accounting")
	_, _ = deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "OPS", "Operations")
	deptMFG, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "MFG", "Manufacturing")
	deptMNT, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "MNT", "Maintenance")
	deptQA, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "QA", "Quality")
	deptHSE, _ := deptSvc.CreateDepartment(ctx, parseUUID(uuid.New().String()), tenant1.ID, nil, "HSE", "Safety")

	// --- 5. Job Titles ---
	jobCEO, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "EXEC-CEO", "Chief Executive Officer", "GRADE-1")
	jobCTO, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "EXEC-CTO", "Chief Technology Officer", "GRADE-1")
	jobCHRO, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "EXEC-CHRO", "Chief HR Officer", "GRADE-1")

	jobITDir, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "IT-DIR", "Director of IT", "GRADE-2")
	jobFinDir, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "FIN-DIR", "Director of Finance", "GRADE-2")

	jobEngMgr, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "IT-MGR", "Engineering Manager", "GRADE-3")
	jobHRMgr, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "HR-MGR", "HR Manager", "GRADE-3")

	jobSrEng, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "IT-SENG", "Senior Software Engineer", "GRADE-4")
	jobEng, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "IT-ENG", "Software Engineer", "GRADE-5")
	jobHRBP, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "HR-BP", "HR Business Partner", "GRADE-4")
	jobAcc, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "FIN-ACC", "Accountant", "GRADE-5")

	jobMfgMgr, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "MFG-MGR", "Manufacturing Manager", "GRADE-3")
	jobMfgOp, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "MFG-OP", "Machine Operator", "GRADE-5")
	jobMntSup, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "MNT-SUP", "Maintenance Supervisor", "GRADE-4")
	jobMntTech, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "MNT-TECH", "Maintenance Technician", "GRADE-5")
	jobQaMgr, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "QA-MGR", "Quality Assurance Manager", "GRADE-3")
	jobQaInsp, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "QA-INSP", "QA Inspector", "GRADE-5")
	jobHseDir, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "HSE-DIR", "HSE Director", "GRADE-2")
	jobHseCoord, _ := jobSvc.CreateJobTitle(ctx, parseUUID(uuid.New().String()), tenant1.ID, "HSE-COORD", "Safety Coordinator", "GRADE-4")

	// --- 6. Executive Layer Onboarding ---
	var existingItDirID string
	err = database.Pool.QueryRow(ctx, "SELECT id FROM employees WHERE employee_no = 'UK-00004' AND tenant_id = $1", tenant1.ID).Scan(&existingItDirID)

	employeeCounter := 10

	if err == nil && existingItDirID != "" {
		log.Println(">> Base Executive and IC Layers already exist. Skipping sections 6-9...")
		var count int
		err = database.Pool.QueryRow(ctx, "SELECT count(*) FROM employees WHERE tenant_id = $1", tenant1.ID).Scan(&count)
		if err == nil {
			employeeCounter = count + 10
		} else {
			employeeCounter = 150
		}
	} else {
		ceoDisp := "Hemish Patel (CEO)"
		ceo, err := onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, "UK-00001", "Hemish", "Patel", &ceoDisp, "hemish.patel@inova.krd", "Testing123!", adminRole.ID,
			buLondon.ID, deptExe.ID, jobCEO.ID, parseUUID(""),
		)
		if err != nil {
			log.Fatalf("FAILED CEO ONBOARDING: %v", err)
		}

		ctoDisp := "Emily T. (CTO)"
		cto, _ := onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, "UK-00002", "Emily", "Taylor", &ctoDisp, "emily.taylor@inova.krd", "Testing123!", adminRole.ID,
			buLondon.ID, deptExe.ID, jobCTO.ID, parseUUID(ceo.EmployeeID),
		)

		chroDisp := "David W. (CHRO)"
		chro, _ := onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, "UK-00003", "David", "Williams", &chroDisp, "david.williams@inova.krd", "Testing123!", adminRole.ID,
			buLondon.ID, deptExe.ID, jobCHRO.ID, parseUUID(ceo.EmployeeID),
		)

		// --- 7. Director Layer Onboarding ---
		itDir, _ := onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, "UK-00004", "James", "Davies", nil, "james.davies@inova.krd", "Testing123!", adminRole.ID,
			buMan.ID, deptIT.ID, jobITDir.ID, parseUUID(cto.EmployeeID),
		)

		finDir, _ := onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, "UK-00005", "Sarah", "Evans", nil, "sarah.evans@inova.krd", "Testing123!", mgrRole.ID,
			buLondon.ID, deptFin.ID, jobFinDir.ID, parseUUID(ceo.EmployeeID),
		)

		// --- 8. Manager Layer ---
		var hrManagers []string
		for i := 1; i <= 3; i++ {
			f, l := randomName()
			email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, i))
			mgr, _ := onboardSvc.ExecuteOnboarding(
				ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-H%03d", i), f, l, nil, email, "Testing123!", hrRole.ID,
				buLondon.ID, deptHR.ID, jobHRMgr.ID, parseUUID(chro.EmployeeID),
			)
			hrManagers = append(hrManagers, mgr.EmployeeID)
		}

		var engManagers []string
		for i := 1; i <= 4; i++ {
			f, l := randomName()
			email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, i))
			mgr, _ := onboardSvc.ExecuteOnboarding(
				ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-E%03d", i), f, l, nil, email, "Testing123!", mgrRole.ID,
				buEdin.ID, deptIT.ID, jobEngMgr.ID, parseUUID(itDir.EmployeeID),
			)
			engManagers = append(engManagers, mgr.EmployeeID)
		}

		var finManagers []string
		for i := 1; i <= 2; i++ {
			f, l := randomName()
			email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, i))
			mgr, _ := onboardSvc.ExecuteOnboarding(
				ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-F%03d", i), f, l, nil, email, "Testing123!", mgrRole.ID,
				buLondon.ID, deptFin.ID, jobAcc.ID, parseUUID(finDir.EmployeeID),
			)
			finManagers = append(finManagers, mgr.EmployeeID)
		}

		// --- 9. Generating 105 Individual Contributors Programmatically ---
		// Distribute ~60 to Engineering, ~20 to Finance, ~25 to HR
		log.Println(">> Generating 105 Standard Employees and Binding to Hierarchy...")

		employeeCounter := 10

		// ENGINEERING
		for i := 0; i < 60; i++ {
			f, l := randomName()
			email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, employeeCounter))

			// Map IC to a random Engineering Manager natively
			assignedMgr := engManagers[rand.Intn(len(engManagers))]
			selectedJob := jobEng.ID
			if rand.Float32() > 0.7 {
				selectedJob = jobSrEng.ID
			} // 30% Senior

			onboardSvc.ExecuteOnboarding(
				ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-%05d", employeeCounter), f, l, nil, email, "Testing123!", empRole.ID,
				buEdin.ID, deptIT.ID, selectedJob, parseUUID(assignedMgr),
			)
			employeeCounter++
		}

		// HR STAFF
		for i := 0; i < 25; i++ {
			f, l := randomName()
			email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, employeeCounter))

			assignedMgr := hrManagers[rand.Intn(len(hrManagers))]

			onboardSvc.ExecuteOnboarding(
				ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-%05d", employeeCounter), f, l, nil, email, "Testing123!", empRole.ID,
				buLondon.ID, deptHR.ID, jobHRBP.ID, parseUUID(assignedMgr),
			)
			employeeCounter++
		}

		// FINANCE STAFF
		for i := 0; i < 20; i++ {
			f, l := randomName()
			email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, employeeCounter))

			assignedMgr := finManagers[rand.Intn(len(finManagers))]

			onboardSvc.ExecuteOnboarding(
				ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-%05d", employeeCounter), f, l, nil, email, "Testing123!", empRole.ID,
				buMan.ID, deptFin.ID, jobAcc.ID, parseUUID(assignedMgr),
			)
			employeeCounter++
		}
		existingItDirID = itDir.EmployeeID
	}

	// --- 10. New Functional Departments Layer ---
	log.Println(">> Generating New Functional Departments Hierarchy...")

	var f, l, email string

	f, l = randomName()
	email = strings.ToLower(fmt.Sprintf("%s.%s@inova.krd", f, l))
	mfgMgr, _ := onboardSvc.ExecuteOnboarding(
		ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", mgrRole.ID,
		buMan.ID, deptMFG.ID, jobMfgMgr.ID, parseUUID(existingItDirID),
	)
	employeeCounter++

	f, l = randomName()
	email = strings.ToLower(fmt.Sprintf("%s.%s@inova.krd", f, l))
	mntSup, _ := onboardSvc.ExecuteOnboarding(
		ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", mgrRole.ID,
		buMan.ID, deptMNT.ID, jobMntSup.ID, parseUUID(existingItDirID),
	)
	employeeCounter++

	f, l = randomName()
	email = strings.ToLower(fmt.Sprintf("%s.%s@inova.krd", f, l))
	qaMgr, _ := onboardSvc.ExecuteOnboarding(
		ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", mgrRole.ID,
		buMan.ID, deptQA.ID, jobQaMgr.ID, parseUUID(existingItDirID),
	)
	employeeCounter++

	f, l = randomName()
	email = strings.ToLower(fmt.Sprintf("%s.%s@inova.krd", f, l))
	hseDir, _ := onboardSvc.ExecuteOnboarding(
		ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", mgrRole.ID,
		buMan.ID, deptHSE.ID, jobHseDir.ID, parseUUID(existingItDirID),
	)
	employeeCounter++

	// 20 Manufacturing
	for i := 0; i < 20; i++ {
		f, l := randomName()
		email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, employeeCounter))
		onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", empRole.ID,
			buMan.ID, deptMFG.ID, jobMfgOp.ID, parseUUID(mfgMgr.EmployeeID),
		)
		employeeCounter++
	}

	// 10 Maintenance
	for i := 0; i < 10; i++ {
		f, l := randomName()
		email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, employeeCounter))
		onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", empRole.ID,
			buMan.ID, deptMNT.ID, jobMntTech.ID, parseUUID(mntSup.EmployeeID),
		)
		employeeCounter++
	}

	// 4 Quality
	for i := 0; i < 4; i++ {
		f, l := randomName()
		email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, employeeCounter))
		onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", empRole.ID,
			buMan.ID, deptQA.ID, jobQaInsp.ID, parseUUID(qaMgr.EmployeeID),
		)
		employeeCounter++
	}

	// 4 Safety
	for i := 0; i < 4; i++ {
		f, l := randomName()
		email := strings.ToLower(fmt.Sprintf("%s.%s%d@inova.krd", f, l, employeeCounter))
		onboardSvc.ExecuteOnboarding(
			ctx, tenant1.ID, sysUserUUID, fmt.Sprintf("UK-N%03d", employeeCounter), f, l, nil, email, "Testing123!", empRole.ID,
			buMan.ID, deptHSE.ID, jobHseCoord.ID, parseUUID(hseDir.EmployeeID),
		)
		employeeCounter++
	}

	// Verify Seeding
	var queryCount int64
	err = database.Pool.QueryRow(ctx, "SELECT count(*) FROM employees WHERE tenant_id = $1", tenant1.ID).Scan(&queryCount)
	if err == nil {
		log.Printf(">> SUCCESSFULLY SEEDED ~%d Employees natively distributed across exact Corporate boundaries!", queryCount)
	} else {
		log.Printf(">> Successfully Seeded data! You can login with hemish.patel@inova.krd and password 'Testing123!'")
	}

	os.Exit(0)
}
