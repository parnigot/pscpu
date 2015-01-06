package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	VERSION                 string      = "0.1"
	CSV_FILE_MODE           os.FileMode = 0644
	DEFAULT_CSV_FOLDER_PATH string      = "" // current working directory
	DEFAULT_PID             uint        = 0
	DEFAULT_SECONDS         uint        = 5
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

func initFlags() (pid uint, csvFolderPath string, seconds uint) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "pscpu - v%v\n\n", VERSION)
		fmt.Fprintln(os.Stderr, "Monitor cpu usage in % of a process to a csv file.")
		fmt.Fprintln(os.Stderr, "Each line of the csv file will be in the format:\n")
		fmt.Fprintln(os.Stderr, "	RFC3339_TIMESTAMP,CPU_USAGE\n")
		fmt.Fprintln(os.Stderr, "For example: 2015-01-05T14:44:05+01:00,66.6\n")
		fmt.Fprintln(os.Stderr, "Usage:")
		flag.PrintDefaults()
	}
	// Create all the flags
	flag.UintVar(&pid, "pid", DEFAULT_PID,
		"REQUIRED, the pid of the process to monitor")
	flag.StringVar(&csvFolderPath, "f", DEFAULT_CSV_FOLDER_PATH,
		"output folder of the csv file, defaults to current working directory")
	flag.UintVar(&seconds, "s", DEFAULT_SECONDS,
		"collect stats of cpu usage every s seconds")
	// Parse the Flags
	flag.Parse()
	// Additional error checking
	if pid == DEFAULT_PID {
		log.Fatalf("Invalid pid: %d\n", pid)
	}
	return
}

func main() {
	// Process args
	pid, csvFolderPath, seconds := initFlags()
	// Open the output file
	csvFile, err := GetCsvFile(csvFolderPath, pid)
	if err != nil {
		log.Fatalln(err)
	}
	// Create the csv.Writer
	csvWriter := csv.NewWriter(csvFile)
	// Capture ctrl+c and other kill signals to clean up
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	signal.Notify(interruptChannel, syscall.SIGTERM)
	go func() {
		<-interruptChannel
		csvWriter.Flush()
		csvFile.Close()
		os.Exit(0)
	}()
	// Run the main program loop
	var cs *CpuStat
	for {
		// Collect the stats
		cs, err = ProcessCpuStat(pid)
		if err != nil {
			log.Fatalln(err)
		}
		// Write them to the csv file
		err = csvWriter.Write(cs.ToCsvRecord())
		if err != nil {
			log.Fatalln(err)
		}
		// Output human readable stats to stdout
		fmt.Println(cs.String())
		// Sleep for requested time and resume
		time.Sleep(time.Duration(seconds) * time.Second)
	}
}
