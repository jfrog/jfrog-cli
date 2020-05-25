package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	buildutils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-client-go/artifactory"
	"github.com/jfrog/jfrog-client-go/artifactory/buildinfo"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

const (
	Pull                      CommandType = "pull"
	Push                      CommandType = "push"
	foreignLayerMediaType     string      = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"
	imageNotFoundErrorMessage string      = "Could not find docker image in Artifactory, expecting image ID: %s"
	dockerMarkerLayerSuffix   string      = ".marker"
)

// Docker image build info builder.
type Builder interface {
	Build(module string) (*buildinfo.BuildInfo, error)
}

// Create instance of docker build info builder.
func NewBuildInfoBuilder(image Image, repository, buildName, buildNumber string, serviceManager *artifactory.ArtifactoryServicesManager, commandType CommandType) (Builder, error) {
	builder := &buildInfoBuilder{}
	builder.image = image
	builder.repository = repository
	builder.buildName = buildName
	builder.buildNumber = buildNumber
	builder.serviceManager = serviceManager
	builder.commandType = commandType
	builder.repository = repository
	repoDetails, err := serviceManager.GetRepository(repository)
	if err != nil {
		return nil, err
	}
	if repoDetails.Rclass == "remote" {
		builder.repository += "-cache"
	}
	return builder, nil
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

// Create build info for docker image.
func (builder *buildInfoBuilder) Build(module string) (*buildinfo.BuildInfo, error) {
	var err error
	builder.imageId, err = builder.image.Id()
	if err != nil {
		return nil, err
	}
	err = builder.updateArtifactsAndDependencies()
	if err != nil {
		return nil, err
	}

	// Set build properties only when pushing image.
	if builder.commandType == Push {
		_, err = builder.setBuildProperties()
		if err != nil {
			return nil, err
		}
	}

	return builder.createBuildInfo(module)
}

// Search, validate and create artifacts and dependencies of docker image.
func (builder *buildInfoBuilder) updateArtifactsAndDependencies() error {
	// Search for all image layers to get the local path inside Artifactory.
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
func (builder *buildInfoBuilder) getImageLayersFromArtifactory() (searchResults map[string]utils.ResultItem, err error) {
	imagePath := builder.image.Path()

	// Search layers - assuming reverse proxy.
	searchResults, err = searchImageLayers(builder, path.Join(builder.repository, imagePath, "*"))
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	// Search layers - assuming proxy-less (repository path).
	// Need to remove the "/" from the image path.
	searchResults, err = searchImageLayers(builder, path.Join(imagePath[1:], "*"))
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	if builder.commandType == Push {
		return nil, errorutils.CheckError(errors.New(fmt.Sprintf(imageNotFoundErrorMessage, builder.imageId)))
	}

	// If image path includes more than 3 slashes, Artifactory doesn't store this image under 'library',
	// thus we should not look further.
	if strings.Count(imagePath, "/") > 3 {
		return nil, errorutils.CheckError(errors.New(fmt.Sprintf(imageNotFoundErrorMessage, builder.imageId)))
	}

	// Assume reverse proxy - this time with 'library' as part of the path.
	searchResults, err = searchImageLayers(builder, path.Join(builder.repository, "library", imagePath, "*"))
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	// Assume proxy-less - this time with 'library' as part of the path.
	searchResults, err = searchImageLayers(builder, path.Join(builder.buildReverseProxyPathWithLibrary(), "*"))
	if err != nil || searchResults != nil {
		return searchResults, err
	}

	// Image layers not found.
	return nil, errorutils.CheckError(errors.New(fmt.Sprintf(imageNotFoundErrorMessage, builder.imageId)))
}

func (builder *buildInfoBuilder) buildReverseProxyPathWithLibrary() string {
	endOfRepoNameIndex := strings.Index(builder.image.Path()[1:], "/")
	return path.Join(builder.repository, "library", builder.image.Path()[endOfRepoNameIndex+1:])
}

func (builder *buildInfoBuilder) handlePull(manifestDependency, configLayerDependency buildinfo.Dependency, imageManifest *manifest, searchResults map[string]utils.ResultItem) error {
	// Add dependencies.
	builder.dependencies = append(builder.dependencies, manifestDependency)
	builder.dependencies = append(builder.dependencies, configLayerDependency)

	// Add image layers as dependencies.
	for i := 0; i < len(imageManifest.Layers); i++ {
		layerFileName := digestToLayer(imageManifest.Layers[i].Digest)
		item, layerExists := searchResults[layerFileName]
		if !layerExists {
			// Check if layer marker exists in Artifactory.
			item, layerExists = searchResults[layerFileName+".marker"]
			if !layerExists {
				err := builder.handleMissingLayer(imageManifest.Layers[i].MediaType, layerFileName)
				if err != nil {
					return err
				}
				continue
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
	totalLayers := len(imageManifest.Layers)
	totalDependencies := configurationLayer.getNumberOfDependentLayers()
	// Add image layers as artifacts and dependencies.
	for i := 0; i < totalLayers; i++ {
		layerFileName := digestToLayer(imageManifest.Layers[i].Digest)
		item, layerExists := searchResults[layerFileName]
		if !layerExists {
			err := builder.handleMissingLayer(imageManifest.Layers[i].MediaType, layerFileName)
			if err != nil {
				return err
			}
			continue
		}
		// Decide if the layer is also a dependency.
		if i < totalDependencies {
			builder.dependencies = append(builder.dependencies, item.ToDependency())
		}
		builder.artifacts = append(builder.artifacts, item.ToArtifact())
		builder.layers = append(builder.layers, item)
	}
	return nil
}

func (builder *buildInfoBuilder) handleMissingLayer(layerMediaType, layerFileName string) error {
	// Allow missing layer to be of a foreign type.
	if layerMediaType == foreignLayerMediaType {
		log.Info(fmt.Sprintf("Foreign layer: %s is missing in Artifactory and therefore will not be added to the build-info.", layerFileName))
		return nil
	}

	return errorutils.CheckError(errors.New("Could not find layer: " + layerFileName + " in Artifactory"))
}

// Set build properties on docker image layers in Artifactory.
func (builder *buildInfoBuilder) setBuildProperties() (int, error) {
	props, err := buildutils.CreateBuildProperties(builder.buildName, builder.buildNumber)
	if err != nil {
		return 0, err
	}
	return builder.serviceManager.SetProps(services.PropsParams{Items: builder.layers, Props: props})
}

// Create docker build info
func (builder *buildInfoBuilder) createBuildInfo(module string) (*buildinfo.BuildInfo, error) {
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

	if module == "" {
		module = builder.image.Name()
	}
	buildInfo := &buildinfo.BuildInfo{Modules: []buildinfo.Module{{
		Id:           module,
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
	// Remove duplicate layers.
	// Docker manifest may hold 'empty layers', as a result, docker promote will fail to promote the same layer more than once.
	imageManifest.Layers = removeDuplicateDockerLayers(imageManifest.Layers)
	// Check that the manifest ID is the right one.
	if imageManifest.Config.Digest != imageId {
		return nil, buildinfo.Artifact{}, buildinfo.Dependency{}, errorutils.CheckError(errors.New("Found incorrect manifest.json file, expecting image ID: " + imageId))
	}

	artifact = buildinfo.Artifact{Name: "manifest.json", Type: "json", Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}, Path: path.Join(item.Repo, item.Path, item.Name)}
	dependency = buildinfo.Dependency{Id: "manifest.json", Type: "json", Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}}
	return
}

// Download and read the config layer from Artifactory.
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

	artifact = buildinfo.Artifact{Name: digestToLayer(imageId), Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}, Path: path.Join(item.Repo, item.Path, item.Name)}
	dependency = buildinfo.Dependency{Id: digestToLayer(imageId), Checksum: &buildinfo.Checksum{Sha1: item.Actual_Sha1, Md5: item.Actual_Md5}}
	return
}

// Search for image layers in Artifactory.
func searchImageLayers(builder *buildInfoBuilder, imagePathPattern string) (map[string]utils.ResultItem, error) {
	resultMap, err := searchImageHandler(imagePathPattern, builder)
	if err != nil {
		return nil, err
	}

	// Validate image ID layer exists.
	if _, ok := resultMap[digestToLayer(builder.imageId)]; !ok {
		// In case of a fat-manifest, Artifactory will create two folders.
		// One folder named as the image tag, which contains the fat manifest.
		// The second folder, named as image's manifest digest, contains the image layers and the image's manifest.
		if _, ok := resultMap["list.manifest.json"]; ok {
			v, err := builder.image.Manifest()
			if err != nil {
				log.Error("Fail to run docker inspect " + builder.image.Tag() + " \nError: " + err.Error())
				return nil, err
			}
			var listManifest []Manifest
			err = json.Unmarshal([]byte(v), &listManifest)
			if err != nil {
				log.Error("json.Unmarshal failed with " + err.Error() + "\n" + v)
				return nil, err
			}
			result := ""
			for _, manifest := range listManifest {
				if *manifest.SchemaV2Manifest.Config.Digest == builder.imageId {
					result = *manifest.Descriptor.Digest
					break
				}
			}
			if result != "" {
				// Remove the tag from the pattern, and place the manifest digest instead.
				imagePathPattern = strings.Replace(imagePathPattern, "/*", "", 1)
				imagePathPattern = path.Join(imagePathPattern[:strings.LastIndex(imagePathPattern, "/")], strings.Replace(result, ":", "__", 1), "*")
				return searchImageHandler(imagePathPattern, builder)
			}
		}
		return nil, nil
	}
	return resultMap, nil
}

func searchImageHandler(imagePathPattern string, builder *buildInfoBuilder) (map[string]utils.ResultItem, error) {
	resultMap, err := performSearch(imagePathPattern, builder.serviceManager)
	if err != nil {
		return resultMap, err
	}
	if totalDownloaded, err := downloadMarkerLayersToRemoteCache(resultMap, builder); err != nil || totalDownloaded == 0 {
		return resultMap, err
	}
	return performSearch(imagePathPattern, builder.serviceManager)
}

func performSearch(imagePathPattern string, serviceManager *artifactory.ArtifactoryServicesManager) (map[string]utils.ResultItem, error) {
	searchParams := services.NewSearchParams()
	searchParams.ArtifactoryCommonParams = &utils.ArtifactoryCommonParams{}
	searchParams.Pattern = imagePathPattern
	results, err := serviceManager.SearchFiles(searchParams)
	if err != nil {
		return nil, err
	}
	resultMap := map[string]utils.ResultItem{}
	for _, v := range results {
		resultMap[v.Name] = v
	}
	return resultMap, nil
}

// Digest of type sha256:30daa5c11544632449b01f450bebfef6b89644e9e683258ed05797abe7c32a6e to
// sha256__30daa5c11544632449b01f450bebfef6b89644e9e683258ed05797abe7c32a6e
func digestToLayer(digest string) string {
	return strings.Replace(digest, ":", "__", 1)
}

// Get the number of dependencies layers from the config.
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

func removeDuplicateDockerLayers(imageMLayers []layer) []layer {
	res := imageMLayers[:0]
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	for _, v := range imageMLayers {
		if !encountered[v.Digest] {
			res = append(res, v)
			encountered[v.Digest] = true
		}
	}
	return res
}

// When a client tries to pull a docker image from a remote repository in Artifactory and the client has some the layers cached locally on the disk,
// then Artifactory will not download these layers into the remote repository cache. Instead, it will mark the layer artifacts with .marker suffix files in the remote cache.
// This function download all the marker layers into the remote cache repository.
func downloadMarkerLayersToRemoteCache(resultMap map[string]utils.ResultItem, builder *buildInfoBuilder) (int, error) {
	if !strings.HasSuffix(builder.repository, "-cache") || len(resultMap) == 0 {
		return 0, nil
	}
	totalDownloaded := 0
	remoteRepo := strings.TrimSuffix(builder.repository, "-cache")
	imageName := getDockerImageName(builder.image.Tag())
	clientDetails := builder.serviceManager.GetConfig().GetServiceDetails().CreateHttpClientDetails()
	// Search for marker layers
	for _, layerData := range resultMap {
		if strings.HasSuffix(layerData.Name, dockerMarkerLayerSuffix) {
			log.Debug(fmt.Sprintf("Downloading %s layer into remote repository cache...", layerData.Name))
			baseUrl := builder.serviceManager.GetConfig().GetServiceDetails().GetUrl()
			endpoint := "api/docker/" + remoteRepo + "/v2/" + imageName + "/blobs/" + toNoneMarkerLayer(layerData.Name)
			resp, body, err := builder.serviceManager.Client().SendHead(baseUrl+endpoint, &clientDetails)
			if err != nil {
				return totalDownloaded, err
			}
			if resp.StatusCode != http.StatusOK {
				return totalDownloaded, errorutils.CheckError(errors.New("Artifactory response: " + resp.Status + "for" + string(body)))
			}
			totalDownloaded++
		}
	}
	return totalDownloaded, nil
}

func getDockerImageName(image string) string {
	imageId, tag := strings.Index(image, "/"), strings.Index(image, ":")
	if imageId == -1 || tag == -1 {
		return ""
	}
	return image[imageId+1 : tag]
}

func toNoneMarkerLayer(layer string) string {
	imageId := strings.Replace(layer, "__", ":", 1)
	return strings.Replace(imageId, ".marker", "", 1)
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
	Digest    string `json:"digest,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
}

type CommandType string
