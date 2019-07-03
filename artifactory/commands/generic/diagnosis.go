package generic

import (
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"regexp"
	"io/ioutil"
	"path/filepath"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	//"errors"
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/progressbar"
	ioUtils "github.com/jfrog/jfrog-client-go/utils/io"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	//"github.com/jfrog/jfrog-client-go/utils/log"
	"text/tabwriter"
	"os"
	"time"
	"strconv"
	//"bufio"
	"fmt"
	"encoding/json"
)

type DiagCommand struct {
	GenericCommand
	logFile       *os.File
	thread 		 int
}

func NewDiagCommand() *DiagCommand {
	return &DiagCommand{GenericCommand: *NewGenericCommand()}
}

func (dc *DiagCommand) LogFile() *os.File {
	return dc.logFile
}

func (dc *DiagCommand) SetThread(thr int) {
	dc.thread = thr
}

func (dc *DiagCommand) GetThread() int {
	return dc.thread
}

func (dc *DiagCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return dc.rtDetails, nil
}

func (dc *DiagCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *DiagCommand {
	dc.rtDetails = rtDetails
	return dc
}

func (dc *DiagCommand) CommandName() string {
	return "rt_diagnosis"
}

func (dc *DiagCommand) ExecSystemInfo() (interface{}, error) {
	return "",nil
}

func (dc *DiagCommand) Run() error {
	var err error
	var progressBar ioUtils.Progress
	progressBar, dc.logFile, err = progressbar.InitProgressBarIfPossible()
	if err != nil {
		return err
	}
	if progressBar != nil {
		defer progressBar.Quit()
	}
	servicesManager, err := utils.CreateDiagnosisServiceManager(dc.rtDetails, dc.thread, false, progressBar)  //utils.CreateServiceManager(dc.rtDetails, false)
	if err != nil {
		return err
	}
	oldResult,err := dc.readOldResults()
	if err != nil {
		fmt.Println(err)
		//return err
	}
	result,err := servicesManager.Diagnostics()
	if err != nil {
		return err
	}
	dc.compareResultsAndDisplay(oldResult,result)

	err = dc.saveCurrentResults(result)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (dc *DiagCommand) readOldResults() (*services.DiagnosisResult,error){
	logDir,err := config.CreateDirInJfrogHome("logs")
	if err != nil {
		return nil,err
	}
	jsonFiles,err := ioutil.ReadDir(logDir)
	if err != nil {
		return nil,err
	}
	var modTime time.Time
	var jsonFileName string

	for _,fi := range jsonFiles {
		//if found,err := filepath.Match("perf*json",fi.Name()); err != nil && found  {
		if ok,_ := regexp.MatchString("perf.*.json",fi.Name()); ok {
			if fi.Mode().IsRegular() {
				if !fi.ModTime().Before(modTime) {
					if fi.ModTime().After(modTime) {
						modTime = fi.ModTime()
						jsonFileName = fi.Name()
					}
				}
			}
		}
	}
	jsonFile, err := os.Open(logDir+"/"+jsonFileName)
	if err != nil {
		return nil,nil
	}
	//defer jsonFile.Close()
	bytevalue,err := ioutil.ReadAll(jsonFile)
	if err != nil {
		
		return nil,err
	}
	var data *services.DiagnosisResult
	json.Unmarshal(bytevalue,&data)
	return data,nil

}

func (dc *DiagCommand) saveCurrentResults(tasks *services.DiagnosisResult) error {
	logDir,err := config.CreateDirInJfrogHome("logs")
	if err != nil {
		return err
	}
	currentTime := time.Now().Format("2006-01-02.15-04-05")
	pid := os.Getpid()
	fileName := filepath.Join(logDir, "perf."+currentTime+"."+strconv.Itoa(pid)+".json")
	jsonFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return errorutils.CheckError(err)
	}
	defer jsonFile.Close() 
	jsonData,err := json.Marshal(tasks)
	if err != nil {
		return errorutils.CheckError(err)
	}
	jsonFile.Write(jsonData)
	return nil
}

func (dc *DiagCommand) compareResultsAndDisplay(oldResult, newResult *services.DiagnosisResult) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.Debug|tabwriter.AlignRight)
	fmt.Fprintf(w, "Task\tLast Run\tThis Run\tStatus\t\n")
	oldVersion := "x"
	oldUrl := "x"
	if oldResult != nil {
		oldVersion = oldResult.Version
		oldUrl = oldResult.Url
	}
	newVersion := "x"
	newUrl := "x"
	if newResult != nil {
		newVersion = newResult.Version
		newUrl = newResult.Url
	}
	fmt.Fprintf(w, "Version\t%s\t%s\t✓\t\n",oldVersion,newVersion)
	fmt.Fprintf(w, "URL\t%s\t%s\t✓\t\n",oldUrl,newUrl)
	i := 0
	if newResult != nil {
		for taskName, taskResult := range newResult.Tasks {
			status := "✓"
			if taskResult.Err != nil {
				status = "x"
			}
			if oldResult != nil {
				if oldTaskResult,ok := oldResult.Tasks[taskName]; ok {
					fmt.Fprintf(w, "%s\t%v\t%v\t%s\t\n",taskName,oldTaskResult.ElapsedTime,taskResult.ElapsedTime,status)
				}
			} else {
				fmt.Fprintf(w, "%s\tx\t%v\t%s\t\n",taskName,taskResult.ElapsedTime,status)
			}
			i++
		}
	}
	fmt.Fprintln(w)
	w.Flush()
	fmt.Println("Total: ",i)
}
