package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	buildutils "github.com/jfrog/jfrog-cli-go/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"io/ioutil"
	"path"
	"strings"
)

const (
	Pull CommandType = "pull"
	Push CommandType = "push"
)

const ImageNotFoundErrorMessage string = "Could not find docker image in Artifactory, expecting image ID: %s"

// Docker image build info builder
type Builder interface {
	Build() (*buildinfo.BuildInfo, error)
}

// Create instance of docker build info builder
func BuildInfoBuilder(image Image, repository, buildName, buildNumber string, serviceManager *artifactory.ArtifactoryServicesManager, commandType CommandType) Builder {
	builder := &buildInfoBuilder{}
	builder.image = image
	builder.repository = repository
	builder.buildName = buildName
	builder.buildNumber = buildNumber
	builder.serviceManager = serviceManager
	builder.commandType = commandType
	return builder
}

type buildInfoBuilder struct {
	image          Image
	repository     string
	buildName      string
	buildNumber    string
	serviceManager *artifactory.ArtifactoryServicesManager

	// internal fields
	imageId      string
	layers       []utils.ResultItem
	artifacts    []buildinfo.Artifact
	dependencies []buildinfo.Dependency
	commandType  CommandType
}

// Create build info for docker image
func (builder *buildInfoBuilder) Build() (*buildinfo.BuildInfo, error) {
	var err error
	builder.imageId, err = builder.image.Id()
	if err != nil {
		return nil, err
	}

	err = builder.updateArtifactsAndDependencies()
	if err != nil {
		return nil, err
	}

	// Set build properties only when pushing image
	if builder.commandType == Push {
		_, err = builder.setBuildProperties()
		if err != nil {
			return nil, err
		}
	}

	return builder.createBuildInfo()
}

// Search, validate and create artifacts and dependencies of docker image.
func (builder *buildInfoBuilder) updateArtifactsAndDependencies() error {
	// Search for all the image layer to get the local path inside Artifactory (supporting virtual repos).
	searchResults, err := builder.getImageLayersFromArtifactory()
	if err != nil {
		return err
	}

	manifest, manifestArtifact, manifestDependency, err := getManifest(builder.imageId, searchResults, builder.serviceManager)
	if err != nil {
		return err
	}

	configLayer, configLayerArtifact, configLayerDependency, err := getConfigLayer(builder.imageId, searchResults, builder.serviceManager)
	if err != nil {
		return err
	}

	if builder.commandType == Push {
		return builder.handlePush(manifestArtifact, configLayerArtifact, manifest, configLayer, searchResults)
	}

	return builder.handlePull(manifestDependency, configLayerDependency, manifest, searchResults)
}

// First we will try to get assuming using a reverse proxy (sub domain or port methods).
// If fails, we will try the repository path (proxy-less).
func (builder *buildInfoBuilder) getImageLayersFromArtifactory() (map[string]utils.ResultItem, error) {
	var searchResults map[string]utils.ResultItem
	imagePath := builder.image.Path()

	// Search layers - assuming reverse proxy.
	searchResults, err := searchImageLayers(builder.imageId, path.Join(builder.repository, imagePath, "*"), builder.serviceManager)
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	// Search layers - assuming proxy-less (repository path).
	// Need to remove the "/" from the image path.
	searchResults, err = searchImageLayers(builder.imageId, path.Join(imagePath[1:], "*"), builder.serviceManager)
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	if builder.commandType == Push {
		return nil, errorutils.CheckError(errors.New(fmt.Sprintf(ImageNotFoundErrorMessage, builder.imageId)))
	}

	// If image path includes more than 3 slashes, Artifactory doesn't store this image under 'library',
	// thus we should not look further.
	if strings.Count(imagePath, "/") > 3 {
		return nil, errorutils.CheckError(errors.New(fmt.Sprintf(ImageNotFoundErrorMessage, builder.imageId)))
	}

	// Assume reverse proxy - this time with 'library' as part of the path.
	searchResults, err = searchImageLayers(builder.imageId, path.Join(builder.repository, "library", imagePath, "*"), builder.serviceManager)
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	// Assume proxy-less - this time with 'library' as part of the path.
	searchResults, err = searchImageLayers(builder.imageId, path.Join(builder.buildReverseProxyPathWithLibrary(), "*"), builder.serviceManager)
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	// Image layers not found.
	return nil, errorutils.CheckError(errors.New(fmt.Sprintf(ImageNotFoundErrorMessage, builder.imageId)))
}

func (builder *buildInfoBuilder) buildReverseProxyPathWithLibrary() string {
	endOfRepoNameIndex := strings.Index(builder.image.Path()[1:], "/")
	return path.Join(builder.repository, "library", builder.image.Path()[endOfRepoNameIndex+ 1:])
}

func (builder *buildInfoBuilder) handlePull(manifestDependency, configLayerDependency buildinfo.Dependency, imageManifest *manifest, searchResults map[string]utils.ResultItem) error {
	// Add dependencies
	builder.dependencies = append(builder.dependencies, manifestDependency)
	builder.dependencies = append(builder.dependencies, configLayerDependency)

	// Add image layers as dependencies
	for i := 0; i < len(imageManifest.Layers); i++ {
		layerFileName := digestToLayer(imageManifest.Layers[i].Digest)
		item, layerExists := searchResults[layerFileName]
		if !layerExists {
			// Check if layer marker exists in Artifactory
			item, layerExists = searchResults[layerFileName + ".marker"]
			if !layerExists {
				return errorutils.CheckError(errors.New("Could not find layer: " + layerFileName + " or its marker in Artifactory"))
			}
		}
		builder.dependencies = append(builder.dependencies, item.ToDependency())
	}
	return nil
}

func (builder *buildInfoBuilder) handlePush(manifestArtifact, configLayerArtifact buildinfo.Artifact, imageManifest *manifest, configurationLayer *configLayer, searchResults map[string]utils.ResultItem) error {
	// Add artifacts
	builder.artifacts = append(builder.artifacts, manifestArtifact)
	builder.artifacts = append(builder.artifacts, configLayerArtifact)
	// Add layers
	builder.layers = append(builder.layers, searchResults["manifest.json"])
	builder.layers = append(builder.layers, searchResults[digestToLayer(builder.imageId)])

	// Add image layers as artifacts and dependencies
	for i := 0; i < configurationLayer.getNumberLayers(); i++ {
		layerFileName := digestToLayer(imageManifest.Layers[i].Digest)
		item, layerExists := searchResults[layerFileName]
		if !layerExists {
			return errorutils.CheckError(errors.New("Could not find layer: " + layerFileName + " in Artifactory"))
		}
		// Decide if the layer is also a dependency
		if i < configurationLayer.getNumberOfDependentLayers() {
			builder.dependencies = append(builder.dependencies, item.ToDependency())
		}

		builder.artifacts = append(builder.artifacts, item.ToArtifact())
		builder.layers = append(builder.layers, item)
	}
	return nil
}

// Set build properties on docker image layers in Artifactory
func (builder *buildInfoBuilder) setBuildProperties() (int, error) {
	props, err := buildutils.CreateBuildProperties(builder.buildName, builder.buildNumber)
	if err != nil {
		return 0, err
	}
	return builder.serviceManager.SetProps(&services.PropsParamsImpl{Items: builder.layers, Props: props})
}

// Create docker build info
func (builder *buildInfoBuilder) createBuildInfo() (*buildinfo.BuildInfo, error) {
	imageProperties := map[string]string{}
	imageProperties["docker.image.id"] = builder.imageId
	imageProperties["docker.image.tag"] = builder.image.Tag()

	parentId, err := builder.image.ParentId()
	if err != nil {
		return nil, err
	}
	if parentId != "" {
		imageProperties["docker.image.parent"] = parentId
	}

	buildInfo := &buildinfo.BuildInfo{Modules: []buildinfo.Module{{
		Id:           builder.image.Name(),
		Properties:   imageProperties,
		Artifacts:    builder.artifacts,
		Dependencies: builder.dependencies,
	}}}
	return buildInfo, nil
}

// Download and read the manifest from Artifactory.
// Returned values:
// imageManifest - pointer to the manifest struct, retrieved from Artifactory.
// artifact - manifest as buildinfo.Artifact object.
// dependency - manifest as buildinfo.Dependency object.
func getManifest(imageId string, searchResults map[string]utils.ResultItem, serviceManager *artifactory.ArtifactoryServicesManager) (imageManifest *manifest, artifact buildinfo.Artifact, dependency buildinfo.Dependency, err error) {
	item := searchResults["manifest.json"]
	ioReaderCloser, err := serviceManager.ReadRemoteFile(item.GetItemRelativePath())
	if err != nil {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, err
	}
	defer ioReaderCloser.Close()
	content, err := ioutil.ReadAll(ioReaderCloser)
	if err != nil {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, err
	}

	imageManifest = new(manifest)
	err = json.Unmarshal(content, &imageManifest)
	if errorutils.CheckError(err) != nil {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, err
	}

	// Check that the manifest ID is the right one
	if imageManifest.Config.Digest != imageId {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, errorutils.CheckError(errors.New("Found incorrect manifest.json file, expecting image ID: " + imageId))
	}

	artifact = buildinfo.Artifact{Name: "manifest.json", Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}}
	dependency = buildinfo.Dependency{Id: "manifest.json", Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}}
	return
}

// Download and read the config layer from Artifactory
// Returned values:
// configurationLayer - pointer to the configuration layer struct, retrieved from Artifactory.
// artifact - configuration layer as buildinfo.Artifact object.
// dependency - configuration layer as buildinfo.Dependency object.
func getConfigLayer(imageId string, searchResults map[string]utils.ResultItem, serviceManager *artifactory.ArtifactoryServicesManager) (configurationLayer *configLayer, artifact buildinfo.Artifact, dependency buildinfo.Dependency, err error) {
	item := searchResults[digestToLayer(imageId)]
	ioReaderCloser, err := serviceManager.ReadRemoteFile(item.GetItemRelativePath())
	if err != nil {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, err
	}
	defer ioReaderCloser.Close()
	content, err := ioutil.ReadAll(ioReaderCloser)
	if err != nil {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, err
	}

	configurationLayer = new(configLayer)
	err = json.Unmarshal(content, &configurationLayer)
	if err != nil {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, err
	}

	artifact = buildinfo.Artifact{Name: digestToLayer(imageId), Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}}
	dependency = buildinfo.Dependency{Id: digestToLayer(imageId), Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}}
	return
}

// Search for image layers in Artifactory
func searchImageLayers(imageId, imagePathPattern string, serviceManager *artifactory.ArtifactoryServicesManager) (map[string]utils.ResultItem, error) {
	params := utils.SearchParams{}
	params.ArtifactoryCommonParams = &utils.ArtifactoryCommonParams{}
	params.Pattern = imagePathPattern
	results, err := serviceManager.Search(params)
	if err != nil {
		return nil, err
	}
	resultMap := map[string]utils.ResultItem{}
	for _, v := range results {
		resultMap[v.Name] = v
	}

	// Validate image ID layer exists.
	if _, ok := resultMap[digestToLayer(imageId)]; !ok {
		return nil, nil
	}
	return resultMap, nil
}

// Digest of type sha256:30daa5c11544632449b01f450bebfef6b89644e9e683258ed05797abe7c32a6e to
// sha256__30daa5c11544632449b01f450bebfef6b89644e9e683258ed05797abe7c32a6e
func digestToLayer(digest string) string {
	return strings.Replace(digest, ":", "__", 1)
}

// Get the total number of layers from the config
func (configLayer *configLayer) getNumberLayers() int {
	layersNum := len(configLayer.History)
	for i := len(configLayer.History) - 1; i >= 0; i-- {
		if configLayer.History[i].EmptyLayer {
			layersNum--
		}
	}
	return layersNum
}

// Get the number of dependencies layers from the config
func (configLayer *configLayer) getNumberOfDependentLayers() int {
	layersNum := len(configLayer.History)
	newImageLayers := true
	for i := len(configLayer.History) - 1; i >= 0; i-- {
		if newImageLayers {
			layersNum--
		}

		if !newImageLayers && configLayer.History[i].EmptyLayer {
			layersNum--
		}

		createdBy := configLayer.History[i].CreatedBy
		if strings.Contains(createdBy, "ENTRYPOINT") || strings.Contains(createdBy, "MAINTAINER") {
			newImageLayers = false
		}
	}
	return layersNum
}

// To unmarshal config layer file
type configLayer struct {
	History []history `json:"history,omitempty"`
}

type history struct {
	Created    string `json:"created,omitempty"`
	CreatedBy  string `json:"created_by,omitempty"`
	EmptyLayer bool   `json:"empty_layer,omitempty"`
}

// To unmarshal manifest.json file
type manifest struct {
	Config manifestConfig `json:"config,omitempty"`
	Layers []layer        `json:"layers,omitempty"`
}

type manifestConfig struct {
	Digest string `json:"digest,omitempty"`
}

type layer struct {
	Digest string `json:"digest,omitempty"`
}

type CommandType string
