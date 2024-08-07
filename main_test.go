package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func initTestServer(status int, responseFile string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)

		if responseFile != "" {
			http.ServeFile(w, r, responseFile)
		}
	}))

	return server
}

func TestMain(t *testing.T) {
	server := initTestServer(http.StatusOK, "test_files/employees_19-02-2024.zip")
	defer server.Close()
	URL = server.URL

	main()

	testFile, _ := os.ReadFile("test_files/test_employees.csv")
	generatedFile, _ := os.ReadFile("employees_0.csv")
	assert.Equal(t, testFile, generatedFile)
}

func TestMainBadResponse(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	log.SetOutput(logBuffer)

	server := initTestServer(http.StatusNotFound, "")
	defer server.Close()
	URL = server.URL

	main()
	log.Printf("Log buffer: %s", logBuffer.String())
	assert.Contains(t, logBuffer.String(), "Error calling API")
}

func TestMainBadServer(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	log.SetOutput(logBuffer)

	server := initTestServer(http.StatusNotFound, "")
	server.Close()
	URL = server.URL

	main()
	log.Printf("Log buffer: %s", logBuffer.String())
	assert.Contains(t, logBuffer.String(), "Error calling API")
}

func TestMainBadZip(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	log.SetOutput(logBuffer)

	server := initTestServer(http.StatusOK, "")
	defer server.Close()
	URL = server.URL

	main()
	log.Printf("Log buffer: %s", logBuffer.String())
	assert.Contains(t, logBuffer.String(), "not a valid zip file")
}

func TestFilterEmployees(t *testing.T) {
	employees := []Employee{
		{
			DOB:        "2006-01-02T15:04:05Z",
			Name:       Name{First: "John", Last: "Doe"},
			Roles:      []string{"salaried", "Manager"},
			Email:      "test@test.com",
			Department: "HR",
		},
		{
			DOB:        "2006-08-02T15:04:05Z",
			Name:       Name{First: "Jane", Last: "Doe"},
			Roles:      []string{"salaried", "Manager"},
			Email:      "test@test.com",
			Department: "HR",
		},
		{
			DOB:        "2006-08-02T15:04:05Z",
			Name:       Name{First: "Jane", Last: "Doe"},
			Roles:      []string{"Manager"},
			Email:      "test@test.com",
			Department: "HR",
		},
		{
			DOB:        "2006-08-02T15:04:05Z",
			Name:       Name{First: "Jane", Last: "VeryLongLastName"},
			Roles:      []string{"salaried", "Manager"},
			Email:      "test@test.com",
			Department: "IT",
		},
	}

	expectedEmployees := []EmployeeClean{
		{
			FullName:   "Jane Doe",
			Department: "HR",
			Email:      "test@test.com",
		},
		{
			FullName:   "J. VeryLongLastName",
			Department: "IT",
			Email:      "test@test.com",
		},
	}
	result := filterEmployees(employees)

	assert.Equal(t, expectedEmployees, result)
}

func TestFilterEmployeeBadDOB(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	log.SetOutput(logBuffer)

	employees := []Employee{
		{
			DOB:        "not date",
			Name:       Name{First: "John", Last: "Doe"},
			Roles:      []string{"salaried", "Manager"},
			Email:      "test@test.com",
			Department: "HR",
		},
		{
			DOB:        "2006-08-02T15:04:05Z",
			Name:       Name{First: "Jane", Last: "Doe"},
			Roles:      []string{"salaried", "Manager"},
			Email:      "test@test.com",
			Department: "HR",
		},
	}

	expectedEmployees := []EmployeeClean{
		{
			FullName:   "Jane Doe",
			Department: "HR",
			Email:      "test@test.com",
		},
	}
	result := filterEmployees(employees)

	// keeps good employee, ignores bad employee
	assert.Equal(t, expectedEmployees, result)
	// logs the bad employee
	assert.Contains(t, logBuffer.String(), "Error checking if employee was born in summer")
}
