package main

import (
	"fmt"
	"log"
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
	csv := cs.ToCsvRecord()
	return fmt.Sprintf("%v - %v", csv[0], csv[1])
}

// Returns a CpuStat for a the process with the given PID
func ProcessCpuStat(pid uint) (cs *CpuStat) {
	// Run ps to get the % of CPU usage
	psTime := time.Now()
	ps := exec.Command("/bin/ps", "-p", fmt.Sprintf("%v", pid), "-o%cpu=")
	psOut, err := ps.Output()
	if err != nil {
		log.Fatalf("ps exited with an error. Are you sure PID %v is valid?\n", pid)
	}
	// Clean ps output
	psOutString := string(psOut[:])
	psOutString = strings.TrimSpace(psOutString)
	psOutString = strings.Replace(psOutString, ",", ".", 1)
	psPercent, err := strconv.ParseFloat(psOutString, 32)
	if err != nil {
		log.Fatalln(err)
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
	cs := ProcessCpuStat(1)
	fmt.Println(cs.String())
}
