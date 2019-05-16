package utils

import (
	"github.com/jfrog/jfrog-cli-go/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/utils/cliutils"
	"github.com/jfrog/jfrog-cli-go/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/reportusage"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type Command interface {
	// Runs the command
	Run() error
	// Returns the Artifactory details. The report usage will be send to this Artifactory server.
	RtDetails() (*config.ArtifactoryDetails, error)
	// The command name for the usage report.
	CommandName() string
}

func Exec(command Command) error {
	channel := make(chan bool)
	// Triggers the report usage.
	go reportUsage(command, channel)
	// Invoke the command interface
	err := command.Run()
	// Waits for the signal from the report usage to be done.
	<-channel
	return err
}

func reportUsage(command Command, channel chan<- bool) {
	defer signalReportUsageFinished(channel)
	reportUsage, err := clientutils.GetBoolEnvValue(cliutils.ReportUsage, true)
	if err != nil {
		log.Debug(err)
		return
	}
	if reportUsage {
		rtDetails, err := command.RtDetails()
		if err != nil {
			log.Debug(err)
			return
		}
		if rtDetails != nil {
			log.Debug("Sending report usage information...")
			serviceManager, err := utils.CreateServiceManager(rtDetails, false)
			if err != nil {
				log.Debug(err)
				return
			}
			productId := cliutils.ClientAgent + "/" + cliutils.CliVersion
			err = reportusage.SendReportUsage(productId, command.CommandName(), serviceManager)
			if err != nil {
				log.Debug(err)
				return
			}
			log.Debug("Report usage information finished successfully.")
		}
	} else {
		log.Debug("Sending report usage is disabled.")
	}
}

// Set to true when the report usage func exits
func signalReportUsageFinished(ch chan<- bool) {
	ch <- true
}
