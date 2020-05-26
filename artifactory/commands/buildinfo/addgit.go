package buildinfo

import (
	"errors"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/artifactory/utils"
	utilsconfig "github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/exec"
	"strconv"
)

const (
	GitLogLimit               = 100
	ConfigIssuesPrefix        = "issues."
	ConfigParseValueError     = "Failed parsing %s from configuration file: %s"
	MissingConfigurationError = "Configuration file must contain: %s"
)

type BuildAddGitCommand struct {
	buildConfiguration *utils.BuildConfiguration
	dotGitPath         string
	configFilePath     string
	serverId           string
	issuesConfig       *IssuesConfiguration
}

func NewBuildAddGitCommand() *BuildAddGitCommand {
	return &BuildAddGitCommand{}
}

func (config *BuildAddGitCommand) SetIssuesConfig(issuesConfig *IssuesConfiguration) *BuildAddGitCommand {
	config.issuesConfig = issuesConfig
	return config
}

func (config *BuildAddGitCommand) SetConfigFilePath(configFilePath string) *BuildAddGitCommand {
	config.configFilePath = configFilePath
	return config
}

func (config *BuildAddGitCommand) SetDotGitPath(dotGitPath string) *BuildAddGitCommand {
	config.dotGitPath = dotGitPath
	return config
}

func (config *BuildAddGitCommand) SetBuildConfiguration(buildConfiguration *utils.BuildConfiguration) *BuildAddGitCommand {
	config.buildConfiguration = buildConfiguration
	return config
}

func (config *BuildAddGitCommand) SetServerId(serverId string) *BuildAddGitCommand {
	config.serverId = serverId
	return config
}

func (config *BuildAddGitCommand) Run() error {
	log.Info("Collecting git revision and remote url...")
	err := utils.SaveBuildGeneralDetails(config.buildConfiguration.BuildName, config.buildConfiguration.BuildNumber)
	if err != nil {
		return err
	}

	// Find .git folder if it wasn't provided in the command.
	if config.dotGitPath == "" {
		var exists bool
		config.dotGitPath, exists, err = fileutils.FindUpstream(".git", fileutils.Dir)
		if err != nil {
			return err
		}
		if !exists {
			return errorutils.CheckError(errors.New("Could not find .git"))
		}
	}

	// Collect URL and Revision into GitManager.
	gitManager := clientutils.NewGitManager(config.dotGitPath)
	err = gitManager.ReadConfig()
	if err != nil {
		return err
	}

	// Collect issues if required.
	var issues []buildinfo.AffectedIssue
	if config.configFilePath != "" {
		issues, err = config.collectBuildIssues()
		if err != nil {
			return err
		}
	}

	// Populate partials with VCS info.
	populateFunc := func(partial *buildinfo.Partial) {
		partial.Vcs = &buildinfo.Vcs{
			Url:      gitManager.GetUrl(),
			Revision: gitManager.GetRevision(),
		}

		if config.configFilePath != "" {
			partial.Issues = &buildinfo.Issues{
				Tracker:                &buildinfo.Tracker{Name: config.issuesConfig.TrackerName, Version: ""},
				AggregateBuildIssues:   config.issuesConfig.Aggregate,
				AggregationBuildStatus: config.issuesConfig.AggregationStatus,
				AffectedIssues:         issues,
			}
		}
	}
	err = utils.SavePartialBuildInfo(config.buildConfiguration.BuildName, config.buildConfiguration.BuildNumber, populateFunc)
	if err != nil {
		return err
	}

	// Done.
	log.Info("Collected VCS details for", config.buildConfiguration.BuildName+"/"+config.buildConfiguration.BuildNumber+".")
	return nil
}

// Priorities for selecting server:
// 1. 'server-id' flag.
// 2. 'serverID' in config file.
// 3. Default server.
func (config *BuildAddGitCommand) RtDetails() (*utilsconfig.ArtifactoryDetails, error) {
	var serverId string
	if config.serverId != "" {
		serverId = config.serverId
	} else if config.configFilePath != "" {
		// Get the server ID from the conf file.
		var vConfig *viper.Viper
		vConfig, err := utils.ReadConfigFile(config.configFilePath, utils.YAML)
		if err != nil {
			return nil, err
		}
		serverId = vConfig.GetString(ConfigIssuesPrefix + "serverID")
	}
	return utilsconfig.GetArtifactorySpecificConfig(serverId)
}

func (config *BuildAddGitCommand) CommandName() string {
	return "rt_build_add_git"
}

func (config *BuildAddGitCommand) collectBuildIssues() ([]buildinfo.AffectedIssue, error) {
	log.Info("Collecting build issues from VCS...")

	// Check that git exists in path.
	_, err := exec.LookPath("git")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	// Initialize issues-configuration.
	config.issuesConfig = new(IssuesConfiguration)

	// Create config's IssuesConfigurations from the provided spec file.
	err = config.createIssuesConfigs()
	if err != nil {
		return nil, err
	}

	// Get latest build's VCS revision from Artifactory.
	lastVcsRevision, err := config.getLatestVcsRevision()
	if err != nil {
		return nil, err
	}

	// Run issues collection.
	return config.DoCollect(config.issuesConfig, lastVcsRevision)
}

func (config *BuildAddGitCommand) DoCollect(issuesConfig *IssuesConfiguration, lastVcsRevision string) ([]buildinfo.AffectedIssue, error) {
	// Create regex pattern.
	issueRegexp, err := clientutils.GetRegExp(issuesConfig.Regexp)
	if err != nil {
		return nil, err
	}

	// Get log with limit, starting from the latest commit.
	logCmd := &LogCmd{logLimit: issuesConfig.LogLimit, lastVcsRevision: lastVcsRevision}
	var foundIssues []buildinfo.AffectedIssue
	logRegExp := gofrogcmd.CmdOutputPattern{
		RegExp: issueRegexp,
		ExecFunc: func(pattern *gofrogcmd.CmdOutputPattern) (string, error) {
			// Reached here - means no error occurred.

			// Check for out of bound results.
			if len(pattern.MatchedResults)-1 < issuesConfig.KeyGroupIndex || len(pattern.MatchedResults)-1 < issuesConfig.SummaryGroupIndex {
				return "", errors.New("Unexpected result while parsing issues from git log. Make sure that the regular expression used to find issues, includes two capturing groups, for the issue ID and the summary.")
			}
			// Create found Affected Issue.
			foundIssue := buildinfo.AffectedIssue{Key: pattern.MatchedResults[issuesConfig.KeyGroupIndex], Summary: pattern.MatchedResults[issuesConfig.SummaryGroupIndex], Aggregated: false}
			if issuesConfig.TrackerUrl != "" {
				foundIssue.Url = issuesConfig.TrackerUrl + pattern.MatchedResults[issuesConfig.KeyGroupIndex]
			}
			foundIssues = append(foundIssues, foundIssue)
			log.Debug("Found issue: " + pattern.MatchedResults[issuesConfig.KeyGroupIndex])
			return "", nil
		},
	}

	// Change working dir to where .git is.
	wd, err := os.Getwd()
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	defer os.Chdir(wd)
	err = os.Chdir(config.dotGitPath)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	// Run git command.
	_, _, exitOk, err := gofrogcmd.RunCmdWithOutputParser(logCmd, false, &logRegExp)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	if !exitOk {
		// May happen when trying to run git log for non-existing revision.
		return nil, errorutils.CheckError(errors.New("Failed executing git log command."))
	}

	// Return found issues.
	return foundIssues, nil
}

func (config *BuildAddGitCommand) createIssuesConfigs() (err error) {
	// Read file's data.
	err = config.issuesConfig.populateIssuesConfigsFromSpec(config.configFilePath)
	if err != nil {
		return
	}

	// Use 'server-id' flag if provided.
	if config.serverId != "" {
		config.issuesConfig.ServerID = config.serverId
	}

	// Build ArtifactoryDetails from provided serverID.
	err = config.issuesConfig.setArtifactoryDetails()
	if err != nil {
		return
	}

	// Add '/' suffix to URL if required.
	if config.issuesConfig.TrackerUrl != "" {
		// Url should end with '/'
		config.issuesConfig.TrackerUrl = clientutils.AddTrailingSlashIfNeeded(config.issuesConfig.TrackerUrl)
	}

	return
}

func (config *BuildAddGitCommand) getLatestVcsRevision() (string, error) {
	// Get latest build's build-info from Artifactory
	buildInfo, err := config.getLatestBuildInfo(config.issuesConfig)
	if err != nil {
		return "", err
	}

	// Get previous VCS Revision from BuildInfo.
	lastVcsRevision := ""
	if buildInfo.Vcs != nil {
		lastVcsRevision = buildInfo.Vcs.Revision
	}

	return lastVcsRevision, nil
}

func (config *BuildAddGitCommand) getLatestBuildInfo(issuesConfig *IssuesConfiguration) (*buildinfo.BuildInfo, error) {
	// Create services manager to get build-info from Artifactory.
	sm, err := utils.CreateServiceManager(issuesConfig.ArtDetails, false)
	if err != nil {
		return nil, err
	}

	// Get latest build-info from Artifactory.
	buildInfoParams := services.BuildInfoParams{BuildName: config.buildConfiguration.BuildName, BuildNumber: "LATEST"}
	buildInfo, err := sm.GetBuildInfo(buildInfoParams)
	if err != nil {
		return nil, err
	}

	return buildInfo, nil
}

func (ic *IssuesConfiguration) populateIssuesConfigsFromSpec(configFilePath string) (err error) {
	var vConfig *viper.Viper
	vConfig, err = utils.ReadConfigFile(configFilePath, utils.YAML)
	if err != nil {
		return err
	}

	// Validate that the config contains issues.
	if !vConfig.IsSet("issues") {
		return errorutils.CheckError(errors.New(fmt.Sprintf(MissingConfigurationError, "issues")))
	}

	// Get server-id.
	if vConfig.IsSet(ConfigIssuesPrefix + "serverID") {
		ic.ServerID = vConfig.GetString(ConfigIssuesPrefix + "serverID")
	}

	// Set log limit.
	ic.LogLimit = GitLogLimit

	// Get tracker data
	if !vConfig.IsSet(ConfigIssuesPrefix + "trackerName") {
		return errorutils.CheckError(errors.New(fmt.Sprintf(MissingConfigurationError, ConfigIssuesPrefix+"trackerName")))
	}
	ic.TrackerName = vConfig.GetString(ConfigIssuesPrefix + "trackerName")

	// Get issues pattern
	if !vConfig.IsSet(ConfigIssuesPrefix + "regexp") {
		return errorutils.CheckError(errors.New(fmt.Sprintf(MissingConfigurationError, ConfigIssuesPrefix+"regexp")))
	}
	ic.Regexp = vConfig.GetString(ConfigIssuesPrefix + "regexp")

	// Get issues base url
	if vConfig.IsSet(ConfigIssuesPrefix + "trackerUrl") {
		ic.TrackerUrl = vConfig.GetString(ConfigIssuesPrefix + "trackerUrl")
	}

	// Get issues key group index
	if !vConfig.IsSet(ConfigIssuesPrefix + "keyGroupIndex") {
		return errorutils.CheckError(errors.New(fmt.Sprintf(MissingConfigurationError, ConfigIssuesPrefix+"keyGroupIndex")))
	}
	ic.KeyGroupIndex, err = strconv.Atoi(vConfig.GetString(ConfigIssuesPrefix + "keyGroupIndex"))
	if err != nil {
		return errorutils.CheckError(errors.New(fmt.Sprintf(ConfigParseValueError, ConfigIssuesPrefix+"keyGroupIndex", err.Error())))
	}

	// Get issues summary group index
	if !vConfig.IsSet(ConfigIssuesPrefix + "summaryGroupIndex") {
		return errorutils.CheckError(errors.New(fmt.Sprintf(MissingConfigurationError, ConfigIssuesPrefix+"summaryGroupIndex")))
	}
	ic.SummaryGroupIndex, err = strconv.Atoi(vConfig.GetString(ConfigIssuesPrefix + "summaryGroupIndex"))
	if err != nil {
		return errorutils.CheckError(errors.New(fmt.Sprintf(ConfigParseValueError, ConfigIssuesPrefix+"summaryGroupIndex", err.Error())))
	}

	// Get aggregation aggregate
	ic.Aggregate = false
	if vConfig.IsSet(ConfigIssuesPrefix + "aggregate") {
		ic.Aggregate, err = strconv.ParseBool(vConfig.GetString(ConfigIssuesPrefix + "aggregate"))
		if err != nil {
			return errorutils.CheckError(errors.New(fmt.Sprintf(ConfigParseValueError, ConfigIssuesPrefix+"aggregate", err.Error())))
		}
	}

	// Get aggregation status
	if vConfig.IsSet(ConfigIssuesPrefix + "aggregationStatus") {
		ic.AggregationStatus = vConfig.GetString(ConfigIssuesPrefix + "aggregationStatus")
	}

	return nil
}

func (ic *IssuesConfiguration) setArtifactoryDetails() error {
	// If no server-id provided, use default server.
	artDetails, err := utilsconfig.GetArtifactorySpecificConfig(ic.ServerID)
	if err != nil {
		return err
	}
	ic.ArtDetails = artDetails
	return nil
}

type IssuesConfiguration struct {
	ArtDetails        *utilsconfig.ArtifactoryDetails
	Regexp            string
	LogLimit          int
	TrackerUrl        string
	TrackerName       string
	KeyGroupIndex     int
	SummaryGroupIndex int
	Aggregate         bool
	AggregationStatus string
	ServerID          string
}

type LogCmd struct {
	logLimit        int
	lastVcsRevision string
}

func (logCmd *LogCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "git")
	cmd = append(cmd, "log", "--pretty=format:%s", "-"+strconv.Itoa(logCmd.logLimit))
	if logCmd.lastVcsRevision != "" {
		cmd = append(cmd, logCmd.lastVcsRevision+"..")
	}
	return exec.Command(cmd[0], cmd[1:]...)
}

func (logCmd *LogCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (logCmd *LogCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (logCmd *LogCmd) GetErrWriter() io.WriteCloser {
	return nil
}
