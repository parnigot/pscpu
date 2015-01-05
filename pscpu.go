package main

import (
	"errors"
	"fmt"
	//"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

func main() {
	fmt.Println("pscpu!")
	cs, err := ProcessCpuStat(123123)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(cs.String())
	}
}
