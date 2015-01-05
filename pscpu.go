package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	CSV_FILE_MODE os.FileMode = 0644
)

// A CpuStat records the cpu usage, as % as returned by ps of a process
// at a given time
type CpuStat struct {
	time time.Time
	pcpu float32
}

// Convert a CpuStat to a string slice, suitable for a csv file
func (cs *CpuStat) ToCsvRecord() (csvRecord []string) {
	csvRecord = []string{
		cs.time.Format(time.RFC3339),
		fmt.Sprintf("%.1f", cs.pcpu),
	}
	return
}

// Return a string representation of the CpuStat
func (cs *CpuStat) String() string {
	return strings.Join(cs.ToCsvRecord(), " - ")
}

// Returns a CpuStat for a the process with the given PID
// An error will be raised if the ps returns with an error.
func ProcessCpuStat(pid uint) (cs *CpuStat, err error) {
	// Run ps to get the % of CPU usage
	psTime := time.Now()
	ps := exec.Command("/bin/ps", "-p", fmt.Sprintf("%v", pid), "-o%cpu=")
	psOut, err := ps.Output()
	if err != nil {
		errMsg := fmt.Sprintf("Error while launching ps (%v). "+
			"Are you sure PID %v is active?", err.Error(), pid)
		err = errors.New(errMsg)
		return
	}
	// Clean ps output
	psOutString := string(psOut[:])
	psOutString = strings.TrimSpace(psOutString)
	psOutString = strings.Replace(psOutString, ",", ".", 1)
	psPercent, err := strconv.ParseFloat(psOutString, 32)
	if err != nil {
		return
	}
	// Create and return the CpuStat
	cs = &CpuStat{
		time: psTime,
		pcpu: float32(psPercent),
	}
	return
}

// Return an *os.File that can be used to write csv records.
// folder is the destination where the csv file will be created. In that folder,
// the program will create a file called pscpu_<pid>.csv. If the file already
// exists, new lines will be appended.
// An error will be returned if there's an error opening the file.
func GetCsvFile(folder string, pid uint) (csvFile *os.File, err error) {
	// Create the final csv path
	csvFilePath := path.Join(folder, fmt.Sprintf("pscpu_%v.csv", pid))
	// Open/create the file
	csvFile, err = os.OpenFile(csvFilePath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, // append or create it in write only
		CSV_FILE_MODE)
	if err != nil {
		err = errors.New(fmt.Sprintf(
			"Error when creating/opening the csv file. %v", err.Error()))
	}
	return
}

func main() {
	fmt.Println("pscpu!")
}
