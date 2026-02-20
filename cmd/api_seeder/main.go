package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const APIBase = "http://localhost:8081/api/v1"

var firstNames = []string{"Oliver", "George", "Arthur", "Noah", "Leo", "Oscar", "Archie", "Jack", "Amelia", "Isla", "Ava", "Ivy", "Lily", "Mia", "Evie", "Florence", "Sarah", "John", "Liam", "Grace"}
var lastNames = []string{"Smith", "Jones", "Williams", "Taylor", "Davies", "Evans", "Thomas", "Johnson", "Roberts", "Walker", "Wright", "Robinson", "Thompson", "White", "Hughes"}

func randomName() (string, string) {
	return firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))]
}

type AuthResponse struct {
	Token string `json:"token"`
}

type PaginatedResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func doJSONReq(method, url string, token string, body interface{}) (map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonStr, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonStr)
	}

	req, _ := http.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API Error %d: %s", resp.StatusCode, string(respBody))
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	return res, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Println("Authenticating with Production API...")

	loginBody := map[string]string{
		"email":    "hemish.patel@inova.krd",
		"password": "Testing123!",
	}
	authRes, err := doJSONReq("POST", APIBase+"/auth/login", "", loginBody)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	token := authRes["token"].(string)

	log.Println("Token acquired successfully.")

	// Fetch Job Titles and create them if they don't exist
	jobTitlesToCreate := []map[string]string{
		{"code": "MFG-MGR", "name": "Manufacturing Manager", "grade": "GRADE-3"},
		{"code": "MFG-OP", "name": "Machine Operator", "grade": "GRADE-5"},
		{"code": "MNT-SUP", "name": "Maintenance Supervisor", "grade": "GRADE-4"},
		{"code": "MNT-TECH", "name": "Maintenance Technician", "grade": "GRADE-5"},
		{"code": "QA-MGR", "name": "Quality Assurance Manager", "grade": "GRADE-3"},
		{"code": "QA-INSP", "name": "QA Inspector", "grade": "GRADE-5"},
		{"code": "HSE-DIR", "name": "HSE Director", "grade": "GRADE-2"},
		{"code": "HSE-COORD", "name": "Safety Coordinator", "grade": "GRADE-4"},
	}

	for _, jt := range jobTitlesToCreate {
		_, err := doJSONReq("POST", APIBase+"/job-titles", token, jt)
		if err != nil {
			log.Printf("Create JobTitle %s: %v (May already exist)", jt["code"], err)
		} else {
			log.Printf("Created JobTitle %s", jt["code"])
		}
	}

	// Fetch IDs mappings
	log.Println("Fetching mappings mapped out on VPS...")
	buList, _ := doJSONReq("GET", APIBase+"/business-units?pageSize=50", token, nil)
	deptList, _ := doJSONReq("GET", APIBase+"/departments?pageSize=50", token, nil)
	jtList, _ := doJSONReq("GET", APIBase+"/job-titles?pageSize=50", token, nil)

	bus := buList["data"].([]interface{})
	depts := deptList["data"].([]interface{})
	jts := jtList["data"].([]interface{})

	var manOpsID, lonHqID string
	for _, b := range bus {
		bu := b.(map[string]interface{})
		if bu["code"] == "MAN-OPS" {
			manOpsID = bu["id"].(string)
		}
		if bu["code"] == "LON-HQ" {
			lonHqID = bu["id"].(string)
		}
	}
	if manOpsID == "" {
		manOpsID = lonHqID
	} // Fallback

	getDeptID := func(code string) string {
		for _, d := range depts {
			dept := d.(map[string]interface{})
			if dept["code"] == code {
				return dept["id"].(string)
			}
		}
		return ""
	}
	getJtID := func(code string) string {
		for _, j := range jts {
			jtb := j.(map[string]interface{})
			if jtb["code"] == code {
				return jtb["id"].(string)
			}
		}
		return ""
	}

	deptMFG := getDeptID("MFG")
	deptMNT := getDeptID("MNT")
	deptQA := getDeptID("QA")
	deptHSE := getDeptID("HSE")

	if deptMFG == "" || deptMNT == "" || deptQA == "" || deptHSE == "" {
		log.Fatal("Could not resolve new functional departments! Run the previous creation step.")
	}

	empCounter := 800 // High ID to avoid conflict with initial seeder

	createEmp := func(bu, dept, jt, manager, fname, lname string) string {
		email := fmt.Sprintf("%s.%s%d@inova.krd", fname, lname, empCounter)
		payload := map[string]interface{}{
			"employeeNo":     fmt.Sprintf("UK-N%03d", empCounter),
			"firstName":      fname,
			"lastName":       lname,
			"workEmail":      email,
			"jobTitleId":     jt,
			"departmentId":   dept,
			"businessUnitId": bu,
		}
		if manager != "" {
			payload["managerId"] = manager
		}

		res, err := doJSONReq("POST", APIBase+"/employees", token, payload)
		empCounter++
		if err != nil {
			log.Printf("Failed to create %s: %v", fname, err)
			return ""
		}
		return res["id"].(string)
	}

	log.Println("Creating Managers...")
	mfgMgrID := createEmp(manOpsID, deptMFG, getJtID("MFG-MGR"), "", "Robert", "Manufacturer")
	mntSupID := createEmp(manOpsID, deptMNT, getJtID("MNT-SUP"), "", "Alice", "Maintenance")
	qaMgrID := createEmp(manOpsID, deptQA, getJtID("QA-MGR"), "", "William", "Quality")
	hseDirID := createEmp(manOpsID, deptHSE, getJtID("HSE-DIR"), "", "Emily", "Safety")

	log.Println("Creating Individual Contributors...")

	// 20 Manufacturing
	for i := 0; i < 20; i++ {
		f, l := randomName()
		createEmp(manOpsID, deptMFG, getJtID("MFG-OP"), mfgMgrID, f, l)
	}

	// 10 Maintenance
	for i := 0; i < 10; i++ {
		f, l := randomName()
		createEmp(manOpsID, deptMNT, getJtID("MNT-TECH"), mntSupID, f, l)
	}

	// 4 Quality
	for i := 0; i < 4; i++ {
		f, l := randomName()
		createEmp(manOpsID, deptQA, getJtID("QA-INSP"), qaMgrID, f, l)
	}

	// 4 Safety
	for i := 0; i < 4; i++ {
		f, l := randomName()
		createEmp(manOpsID, deptHSE, getJtID("HSE-COORD"), hseDirID, f, l)
	}

	log.Println("Successfully provisioned ~40 hierarchical operators natively through Production API!")
	os.Exit(0)
}
