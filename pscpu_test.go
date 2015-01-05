package main

import (
	"os"
	"testing"
	"time"
)

func TestCpuStatToCsvRecord(t *testing.T) {
	// Test with various percentances
	stat := &CpuStat{pcpu: 89.0}
	pString := stat.ToCsvRecord()[1]
	if pString != "89.0" {
		t.Errorf("Invalid %% formatting: %v", pString)
	}
	stat.pcpu = 0.01
	pString = stat.ToCsvRecord()[1]
	if pString != "0.0" {
		t.Errorf("Invalid %% formatting: %v", pString)
	}
	stat.pcpu = 100.0
	pString = stat.ToCsvRecord()[1]
	if pString != "100.0" {
		t.Errorf("Invalid %% formatting: %v", pString)
	}
	stat.pcpu = 10000.1 // In linux cpu usage can be > 100%
	pString = stat.ToCsvRecord()[1]
	if pString != "10000.1" {
		t.Errorf("Invalid %% formatting: %v", pString)
	}
	// Test time for timezones
	CETLocation, _ := time.LoadLocation("Europe/Rome")
	stat.time = time.Date(2014, time.January, 2, 18, 26, 56, 0, CETLocation)
	timestamp := stat.ToCsvRecord()[0]
	if timestamp != "2014-01-02T18:26:56+01:00" {
		t.Errorf("Invalid timestamp formatting: %v", timestamp)
	}
	stat.time = time.Date(2014, time.January, 2, 18, 26, 56, 0, time.UTC)
	timestamp = stat.ToCsvRecord()[0]
	if timestamp != "2014-01-02T18:26:56Z" {
		t.Errorf("Invalid timestamp formatting: %v", timestamp)
	}
}

func TestCpuStatstring(t *testing.T) {
	// Test with various percentances
	stat := &CpuStat{
		pcpu: 89.0,
		time: time.Date(2014, time.January, 2, 18, 26, 56, 0, time.UTC),
	}
	if stat.String() != "2014-01-02T18:26:56Z - 89.0" {
		t.Errorf("Invalid string: %v", stat.String())
	}
}

func TestGetCsvFileNew(t *testing.T) {
	// Test with a new file in a valid folder
	folder := "/tmp/"
	csvFile, err := GetCsvFile(folder, 1)
	if err != nil {
		t.Errorf("New file should be created without errors: %v", csvFile.Name())
	} else {
		// Clean tmp folder
		csvFile.Close()
		os.Remove(csvFile.Name())
	}
}

func TestGetCsvFileAppend(t *testing.T) {
	// Test with an existing file
	// First create the file and close it
	folder := "/tmp/"
	csvFile, _ := GetCsvFile(folder, 1)
	csvFile.Close()
	// Re-open it
	csvFile, err := GetCsvFile(folder, 1)
	if err != nil {
		t.Errorf("File should be opened in append mode without errors: %v", csvFile.Name())
	} else {
		// Clean tmp folder
		csvFile.Close()
		os.Remove(csvFile.Name())
	}
}

func TestGetCsvFileNonExistingFolder(t *testing.T) {
	// Test creating a file in a non existing folder
	folder := "/non-existing-folder/"
	_, err := GetCsvFile(folder, 1)
	if err == nil {
		t.Errorf("Error expected creating the file in a non-existing folder: %v", folder)
	}
}

func TestGetCsvFileNoPermissions(t *testing.T) {
	// Test creating a file in a folder where the user doesn't have permissions
	folder := "/root/"
	_, err := GetCsvFile(folder, 1)
	if err == nil {
		t.Errorf("Error expected creating the file in folder without privileges: %v", folder)
	}
}
