package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Employee struct {
	DOB        string   `json:"dob"`
	Name       Name     `json:"name"`
	Roles      []string `json:"roles"`
	Email      string   `json:"email"`
	Department string   `json:"department"`
	Username   string   `json:"username"`
}

type EmployeeClean struct {
	FullName   string `json:"full_name"`
	Department string `json:"department"`
	Email      string `json:"email"`
}

type Name struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

var (
	URL = "https://s3.eu-west-2.amazonaws.com/interview.thanskben.com/backend/employees_19-02-2024.zip"
)

func main() {
	resp, err := http.Get(URL)
	if err != nil {
		log.Printf("Error calling API: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error calling API: %v", resp.Status)
		return
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	zipFile, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		log.Printf("Error creating zip reader: %v", err)
		return
	}

	var employees []Employee
	for fileCount, file := range zipFile.File {
		zipContents, err := file.Open()
		if err != nil {
			log.Printf("Problem with zip: %v", err)
			continue
		}

		defer zipContents.Close()

		decodedZip, err := io.ReadAll(zipContents)
		if err != nil {
			log.Printf("Problem with zip: %v", err)
			continue
		}

		err = json.Unmarshal(decodedZip, &employees)
		if err != nil {
			log.Printf("Ignored invalid employees JSON: %v", file.Name)
			continue
		}

		filteredEmployees := filterEmployees(employees)

		csvData := "full_name,department,email\n"
		for _, employee := range filteredEmployees {
			csvData += fmt.Sprintf("%v,%v,%v\n", employee.FullName, employee.Department, employee.Email)
		}

		err = os.WriteFile(fmt.Sprintf("employees_%v.csv", fileCount), []byte(csvData), 0644)
		if err != nil {
			log.Printf("Error writing to file: %v", err)
			continue
		}
	}
}

func filterEmployees(employees []Employee) []EmployeeClean {
	var filteredEmployees []EmployeeClean
	for _, employee := range employees {
		salaried := salariedEmployee(employee)
		bornInSummer, err := bornInSummer(employee)

		if err != nil {
			log.Printf("Error checking if employee was born in summer: %v, %v", employee.Username, err)
		}
		if salaried && bornInSummer {
			cleanEmployee := cleanUpEmployeeData(employee)
			filteredEmployees = append(filteredEmployees, cleanEmployee)
		}
	}
	return filteredEmployees
}

func cleanUpEmployeeData(employee Employee) EmployeeClean {
	splitFirst := strings.Split(employee.Name.First, " ")

	var employeeClean EmployeeClean
	if (len(splitFirst[0]) + len(employee.Name.Last)) > 9 {
		employeeClean.FullName = fmt.Sprintf("%v. %v", string(splitFirst[0][0]), employee.Name.Last)
	} else {
		employeeClean.FullName = fmt.Sprintf("%v %v", employee.Name.First, employee.Name.Last)
	}
	employeeClean.Department = employee.Department
	employeeClean.Email = employee.Email
	return employeeClean
}

func salariedEmployee(employee Employee) bool {
	for _, role := range employee.Roles {
		if role == "salaried" {
			return true
		}
	}

	return false
}

func bornInSummer(employee Employee) (bool, error) {
	layout := time.RFC3339

	// Parse the time string to a time.Time object
	parsedTime, err := time.Parse(layout, employee.DOB)
	if err != nil {
		return false, err
	}

	summerMonths := []time.Month{time.June, time.July, time.August, time.September}
	for _, month := range summerMonths {
		if parsedTime.Month() == month {
			return true, nil
		}
	}
	return false, nil
}
