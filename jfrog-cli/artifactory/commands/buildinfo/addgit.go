package buildinfo

import (
	"errors"
	"fmt"
	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils/git"
	utilsconfig "github.com/jfrog/jfrog-cli-go/jfrog-cli/utils/config"
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

func AddGit(config *BuildAddGitConfiguration) error {
	log.Info("Collecting git revision and remote url...")
	err := utils.SaveBuildGeneralDetails(config.BuildName, config.BuildNumber)
	if err != nil {
		return err
	}

	// Find .git folder if it wasn't provided in the command.
	if config.DotGitPath == "" {
		config.DotGitPath, err = fileutils.FindUpstream(".git", fileutils.Dir)
		if err != nil {
			return err
		}
	}

	// Collect URL and Revision into GitManager.
	gitManager := git.NewManager(config.DotGitPath)
	err = gitManager.ReadConfig()
	if err != nil {
		return err
	}

	// Collect issues if required.
	var issues []buildinfo.AffectedIssue
	if config.ConfigFilePath != "" {
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

		if config.ConfigFilePath != "" {
			partial.Issues = &buildinfo.Issues{
				Tracker:                &buildinfo.Tracker{Name: config.IssuesConfig.TrackerName, Version: ""},
				AggregateBuildIssues:   config.IssuesConfig.Aggregate,
				AggregationBuildStatus: config.IssuesConfig.AggregationStatus,
				AffectedIssues:         issues,
			}
		}
	}
	err = utils.SavePartialBuildInfo(config.BuildName, config.BuildNumber, populateFunc)
	if err != nil {
		return err
	}

	// Done.
	log.Info("Collected VCS details for", config.BuildName+"/"+config.BuildNumber+".")
	return nil
}

func (config *BuildAddGitConfiguration) collectBuildIssues() ([]buildinfo.AffectedIssue, error) {
	log.Info("Collecting build issues from VCS...")

	// Check that git exists in path.
	_, err := exec.LookPath("git")
	if err != nil {
		return nil, errorutils.CheckError(err)
	}

	// Initialize issues-configuration.
	config.IssuesConfig = new(IssuesConfiguration)

	// Create config's IssuesConfigurations from the provided spec file.
	err = config.createIssuesConfigurations()
	if err != nil {
		return nil, err
	}

	// Get latest build's VCS revision from Artifactory.
	lastVcsRevision, err := config.getLatestVcsRevision()
	if err != nil {
		return nil, err
	}

	// Run issues collection.
	return config.DoCollect(config.IssuesConfig, lastVcsRevision)
}

func (config *BuildAddGitConfiguration) DoCollect(issuesConfig *IssuesConfiguration, lastVcsRevision string) ([]buildinfo.AffectedIssue, error) {
	// Create regex pattern.
	issueRegexp, err := clientutils.GetRegExp(issuesConfig.Regexp)
	if err != nil {
		return nil, err
	}

	// Get log with limit, starting from the latest commit.
	logCmd := &LogCmd{logLimit: issuesConfig.LogLimit, lastVcsRevision: lastVcsRevision}
	var foundIssues []buildinfo.AffectedIssue
	protocolRegExp := gofrogcmd.CmdOutputPattern{
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
	err = os.Chdir(config.DotGitPath)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}

	// Run git command.
	_, exitOk, err := gofrogcmd.RunCmdWithOutputParser(logCmd, false, &protocolRegExp)
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

func (config *BuildAddGitConfiguration) createIssuesConfigurations() (err error) {
	// Read file's data.
	err = config.IssuesConfig.populateIssuesConfigurationsFromSpec(config.ConfigFilePath)
	if err != nil {
		return
	}

	// Build ArtifactoryDetails from provided serverID.
	config.IssuesConfig.setArtifactoryDetails()
	if err != nil {
		return
	}

	// Add '/' suffix to URL if required.
	if config.IssuesConfig.TrackerUrl != "" {
		// Url should end with '/'
		config.IssuesConfig.TrackerUrl = clientutils.AddTrailingSlashIfNeeded(config.IssuesConfig.TrackerUrl)
	}

	return
}

func (config *BuildAddGitConfiguration) getLatestVcsRevision() (string, error) {
	// Get latest build's build-info from Artifactory
	buildInfo, err := config.getLatestBuildInfo(config.IssuesConfig)
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

func (config *BuildAddGitConfiguration) getLatestBuildInfo(issuesConfig *IssuesConfiguration) (*buildinfo.BuildInfo, error) {
	// Create services manager to get build-info from Artifactory.
	sm, err := utils.CreateServiceManager(issuesConfig.ArtDetails, false)
	if err != nil {
		return nil, err
	}

	// Get latest build-info from Artifactory.
	buildInfoParams := services.BuildInfoParams{BuildName: config.BuildName, BuildNumber: "LATEST"}
	buildInfo, err := sm.GetBuildInfo(buildInfoParams)
	if err != nil {
		return nil, err
	}

	return buildInfo, nil
}

func (ic *IssuesConfiguration) populateIssuesConfigurationsFromSpec(configFilePath string) (err error) {
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
	if !vConfig.IsSet(ConfigIssuesPrefix+"serverID") || vConfig.GetString(ConfigIssuesPrefix+"serverID") == "" {
		return errorutils.CheckError(errors.New(fmt.Sprintf(MissingConfigurationError, ConfigIssuesPrefix+"serverID")))
	}
	ic.ServerID = vConfig.GetString(ConfigIssuesPrefix + "serverID")

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
	artDetails, err := utilsconfig.GetArtifactoryConf(ic.ServerID)
	if err != nil {
		return err
	}
	ic.ArtDetails = artDetails
	return nil
}

type BuildAddGitConfiguration struct {
	BuildName      string
	BuildNumber    string
	DotGitPath     string
	ConfigFilePath string
	IssuesConfig   *IssuesConfiguration
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
