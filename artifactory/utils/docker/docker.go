package docker

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path"
	"regexp"
	"strings"

	gofrogcmd "github.com/jfrog/gofrog/io"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/auth"
	clientConfig "github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/jfrog/jfrog-client-go/utils/version"
)

// Search for docker API version format pattern e.g. 1.40
var ApiVersionRegex = regexp.MustCompile(`^(\d+)\.(\d+)$`)

// Docker API version 1.31 is compatible with Docker version 17.07.0, according to https://docs.docker.com/engine/api/#api-version-matrix
const MinSupportedApiVersion string = "1.31"

// Docker login error message
const DockerLoginFailureMessage string = "Docker login failed for: %s.\nDocker image must be in the form: docker-registry-domain/path-in-repository/image-name:version."

func New(imageTag string) Image {
	return &image{tag: imageTag}
}

// Docker image
type Image interface {
	Push() error
	Id() (string, error)
	ParentId() (string, error)
	Manifest() (string, error)
	Tag() string
	Path() string
	Name() string
	Pull() error
}

// Internal implementation of docker image
type image struct {
	tag string
}

type DockerLoginConfig struct {
	ArtifactoryDetails *config.ArtifactoryDetails
}

// Push docker image
func (image *image) Push() error {
	cmd := &pushCmd{image: image}
	return gofrogcmd.RunCmd(cmd)
}

// Get docker image tag
func (image *image) Tag() string {
	return image.tag
}

// Get docker image ID
func (image *image) Id() (string, error) {
	cmd := &getImageIdCmd{image: image}
	content, err := gofrogcmd.RunCmdOutput(cmd)
	return strings.Trim(content, "\n"), err
}

// Get docker parent image ID
func (image *image) ParentId() (string, error) {
	cmd := &getParentId{image: image}
	content, err := gofrogcmd.RunCmdOutput(cmd)
	return strings.Trim(content, "\n"), err
}

// Get docker image relative path in Artifactory
func (image *image) Path() string {
	indexOfFirstSlash := strings.Index(image.tag, "/")
	indexOfLastColon := strings.LastIndex(image.tag, ":")

	if indexOfLastColon < 0 || indexOfLastColon < indexOfFirstSlash {
		return path.Join(image.tag[indexOfFirstSlash:], "latest")
	}
	return path.Join(image.tag[indexOfFirstSlash:indexOfLastColon], image.tag[indexOfLastColon+1:])
}

// Get docker image manifest
func (image *image) Manifest() (string, error) {
	cmd := &getImageManifestCmd{image: image}
	content, err := gofrogcmd.RunCmdOutput(cmd)
	return content, err
}

// Get docker image name
func (image *image) Name() string {
	indexOfLastSlash := strings.LastIndex(image.tag, "/")
	indexOfLastColon := strings.LastIndex(image.tag, ":")

	if indexOfLastColon < 0 || indexOfLastColon < indexOfLastSlash {
		return image.tag[indexOfLastSlash+1:] + ":latest"
	}
	return image.tag[indexOfLastSlash+1:]
}

// Pull docker image
func (image *image) Pull() error {
	cmd := &pullCmd{image: image}
	return gofrogcmd.RunCmd(cmd)
}

// Image push command
type pushCmd struct {
	image *image
}

func (pushCmd *pushCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "push")
	cmd = append(cmd, pushCmd.image.tag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (pushCmd *pushCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (pushCmd *pushCmd) GetStdWriter() io.WriteCloser {
	return nil
}
func (pushCmd *pushCmd) GetErrWriter() io.WriteCloser {
	return nil
}

// Image get image id command
type getImageIdCmd struct {
	image *image
}

func (getImageId *getImageIdCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "images")
	cmd = append(cmd, "--format", "{{.ID}}")
	cmd = append(cmd, "--no-trunc")
	cmd = append(cmd, getImageId.image.tag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (getImageId *getImageIdCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (getImageId *getImageIdCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (getImageId *getImageIdCmd) GetErrWriter() io.WriteCloser {
	return nil
}

type Manifest struct {
	Descriptor       Descriptor       `json:"descriptor"`
	SchemaV2Manifest SchemaV2Manifest `json:"SchemaV2Manifest"`
}

type Descriptor struct {
	Digest *string `json:"digest"`
}

type SchemaV2Manifest struct {
	Config Config `json:"config"`
}

type Config struct {
	Digest *string `json:"digest"`
}

// Image get parent image id command
type getParentId struct {
	image *image
}

func (getImageId *getParentId) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "inspect")
	cmd = append(cmd, "--format", "{{.Parent}}")
	cmd = append(cmd, getImageId.image.tag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (getImageId *getParentId) GetEnv() map[string]string {
	return map[string]string{}
}

func (getImageId *getParentId) GetStdWriter() io.WriteCloser {
	return nil
}

func (getImageId *getParentId) GetErrWriter() io.WriteCloser {
	return nil
}

// Get image manifest command
type getImageManifestCmd struct {
	image *image
}

func (getImageManifest *getImageManifestCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "manifest")
	cmd = append(cmd, "inspect")
	cmd = append(cmd, getImageManifest.image.tag)
	cmd = append(cmd, "--verbose")
	return exec.Command(cmd[0], cmd[1:]...)
}

func (getImageManifest *getImageManifestCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (getImageManifest *getImageManifestCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (getImageManifest *getImageManifestCmd) GetErrWriter() io.WriteCloser {
	return nil
}

// Get docker registry from tag
func ResolveRegistryFromTag(imageTag string) (string, error) {
	indexOfFirstSlash := strings.Index(imageTag, "/")
	if indexOfFirstSlash < 0 {
		err := errorutils.CheckError(errors.New("Invalid image tag received for pushing to Artifactory - tag does not include a slash."))
		return "", err
	}

	indexOfSecondSlash := strings.Index(imageTag[indexOfFirstSlash+1:], "/")
	// Reverse proxy Artifactory
	if indexOfSecondSlash < 0 {
		return imageTag[:indexOfFirstSlash], nil
	}
	// Can be reverse proxy or proxy-less Artifactory
	indexOfSecondSlash += indexOfFirstSlash + 1
	return imageTag[:indexOfSecondSlash], nil
}

// Login command
type LoginCmd struct {
	DockerRegistry string
	Username       string
	Password       string
}

func (loginCmd *LoginCmd) GetCmd() *exec.Cmd {
	if cliutils.IsWindows() {
		return exec.Command("cmd", "/C", "echo", "%DOCKER_PASS%|", "docker", "login", loginCmd.DockerRegistry, "--username", loginCmd.Username, "--password-stdin")
	}
	cmd := "echo $DOCKER_PASS " + fmt.Sprintf(`| docker login %s --username="%s" --password-stdin`, loginCmd.DockerRegistry, loginCmd.Username)
	return exec.Command("sh", "-c", cmd)
}

func (loginCmd *LoginCmd) GetEnv() map[string]string {
	return map[string]string{"DOCKER_PASS": loginCmd.Password}
}

func (loginCmd *LoginCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (loginCmd *LoginCmd) GetErrWriter() io.WriteCloser {
	return nil
}

// Image pull command
type pullCmd struct {
	image *image
}

func (pullCmd *pullCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "pull")
	cmd = append(cmd, pullCmd.image.tag)
	return exec.Command(cmd[0], cmd[1:]...)
}

func (pullCmd *pullCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (pullCmd *pullCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (pullCmd *pullCmd) GetErrWriter() io.WriteCloser {
	return nil
}

func CreateServiceManager(artDetails *config.ArtifactoryDetails, threads int) (artifactory.ArtifactoryServicesManager, error) {
	certsPath, err := cliutils.GetJfrogCertsDir()
	if err != nil {
		return nil, err
	}
	artAuth, err := artDetails.CreateArtAuthConfig()
	if err != nil {
		return nil, err
	}

	configBuilder := clientConfig.NewConfigBuilder().
		SetServiceDetails(artAuth).
		SetCertificatesPath(certsPath).
		SetInsecureTls(artDetails.InsecureTls).
		SetThreads(threads)

	if threads != 0 {
		configBuilder.SetThreads(threads)
	}

	serviceConfig, err := configBuilder.Build()
	return artifactory.New(&artAuth, serviceConfig)
}

// First will try to login assuming a proxy-less tag (e.g. "registry-address/docker-repo/image:ver").
// If fails, we will try assuming a reverse proxy tag (e.g. "registry-address-docker-repo/image:ver").
func DockerLogin(imageTag string, config *DockerLoginConfig) error {
	imageRegistry, err := ResolveRegistryFromTag(imageTag)
	if err != nil {
		return err
	}

	username := config.ArtifactoryDetails.User
	password := config.ArtifactoryDetails.Password
	// If access-token exists, perform login with it.
	if config.ArtifactoryDetails.AccessToken != "" {
		log.Debug("Using access-token details in docker-login command.")
		username, err = auth.ExtractUsernameFromAccessToken(config.ArtifactoryDetails.AccessToken)
		if err != nil {
			return err
		}
		password = config.ArtifactoryDetails.AccessToken
	}

	// Perform login.
	cmd := &LoginCmd{DockerRegistry: imageRegistry, Username: username, Password: password}
	err = gofrogcmd.RunCmd(cmd)

	if exitCode := cliutils.GetExitCode(err, 0, 0, false); exitCode == cliutils.ExitCodeNoError {
		// Login succeeded
		return nil
	}
	log.Debug("Docker login while assuming proxy-less failed:", err)

	indexOfSlash := strings.Index(imageRegistry, "/")
	if indexOfSlash < 0 {
		return errorutils.CheckError(errors.New(fmt.Sprintf(DockerLoginFailureMessage, imageRegistry)))
	}

	cmd = &LoginCmd{DockerRegistry: imageRegistry[:indexOfSlash], Username: config.ArtifactoryDetails.User, Password: config.ArtifactoryDetails.Password}
	err = gofrogcmd.RunCmd(cmd)
	if err != nil {
		// Login failed for both attempts
		return errorutils.CheckError(errors.New(fmt.Sprintf(DockerLoginFailureMessage,
			fmt.Sprintf("%s, %s", imageRegistry, imageRegistry[:indexOfSlash])) + " " + err.Error()))
	}

	// Login succeeded
	return nil
}

// Version command
type VersionCmd struct{}

func (versionCmd *VersionCmd) GetCmd() *exec.Cmd {
	var cmd []string
	cmd = append(cmd, "docker")
	cmd = append(cmd, "version")
	cmd = append(cmd, "--format", "{{.Client.APIVersion}}")
	return exec.Command(cmd[0], cmd[1:]...)
}

func (versionCmd *VersionCmd) GetEnv() map[string]string {
	return map[string]string{}
}

func (versionCmd *VersionCmd) GetStdWriter() io.WriteCloser {
	return nil
}

func (versionCmd *VersionCmd) GetErrWriter() io.WriteCloser {
	return nil
}

func ValidateClientApiVersion() error {
	cmd := &VersionCmd{}
	// 'docker version' may return 1 in case of errors from daemon. We should ignore this kind of errors.
	content, err := gofrogcmd.RunCmdOutput(cmd)
	content = strings.TrimSpace(content)
	if !ApiVersionRegex.Match([]byte(content)) {
		// The Api version is expected to be 'major.minor'. Anything else should return an error.
		return errorutils.CheckError(err)
	}
	if !IsCompatibleApiVersion(content) {
		return errorutils.CheckError(errors.New("This operation requires Docker API version " + MinSupportedApiVersion + " or higher."))
	}
	return nil
}

func IsCompatibleApiVersion(dockerOutput string) bool {
	currentVersion := version.NewVersion(dockerOutput)
	return currentVersion.AtLeast(MinSupportedApiVersion)
}
