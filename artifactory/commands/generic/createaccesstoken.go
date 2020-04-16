package generic

import (
	"encoding/json"
	rtUtils "github.com/jfrog/jfrog-cli/artifactory/utils"
	"github.com/jfrog/jfrog-cli/utils/config"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

const (
	GroupsPrefix          = "member-of-groups:"
	UserScopedNotation    = "*"
	AdminPrivilegesSuffix = ":admin"
)

type AccessTokenCreateCommand struct {
	rtDetails   *config.ArtifactoryDetails
	refreshable bool
	expiry      int
	userName    string
	audience    string
	groups      string
	grantAdmin  bool
	response    *services.CreateTokenResponseData
}

func NewAccessTokenCreateCommand() *AccessTokenCreateCommand {
	return &AccessTokenCreateCommand{response: new(services.CreateTokenResponseData)}
}

func (atcc *AccessTokenCreateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *AccessTokenCreateCommand {
	atcc.rtDetails = rtDetails
	return atcc
}

func (atcc *AccessTokenCreateCommand) SetRefreshable(refreshable bool) *AccessTokenCreateCommand {
	atcc.refreshable = refreshable
	return atcc
}

func (atcc *AccessTokenCreateCommand) SetExpiry(expiry int) *AccessTokenCreateCommand {
	atcc.expiry = expiry
	return atcc
}

func (atcc *AccessTokenCreateCommand) SetUserName(userName string) *AccessTokenCreateCommand {
	atcc.userName = userName
	return atcc
}

func (atcc *AccessTokenCreateCommand) SetAudience(audience string) *AccessTokenCreateCommand {
	atcc.audience = audience
	return atcc
}

func (atcc *AccessTokenCreateCommand) SetGrantAdmin(grantAdmin bool) *AccessTokenCreateCommand {
	atcc.grantAdmin = grantAdmin
	return atcc
}

func (atcc *AccessTokenCreateCommand) SetGroups(groups string) *AccessTokenCreateCommand {
	atcc.groups = groups
	return atcc
}

func (atcc *AccessTokenCreateCommand) Response() ([]byte, error) {
	content, err := json.Marshal(*atcc.response)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	return content, nil
}

func (atcc *AccessTokenCreateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return atcc.rtDetails, nil
}

func (atcc *AccessTokenCreateCommand) CommandName() string {
	return "rt_create_access_token"
}

func (atcc *AccessTokenCreateCommand) Run() error {
	servicesManager, err := rtUtils.CreateServiceManager(atcc.rtDetails, false)
	if err != nil {
		return err
	}
	tokenParams, err := atcc.getTokenParams()
	if err != nil {
		return err
	}

	*atcc.response, err = servicesManager.CreateToken(tokenParams)
	return err
}

func (atcc *AccessTokenCreateCommand) getTokenParams() (tokenParams services.CreateTokenParams, err error) {
	tokenParams = services.NewCreateTokenParams()
	tokenParams.Username = atcc.userName
	tokenParams.ExpiresIn = atcc.expiry
	tokenParams.Refreshable = atcc.refreshable
	tokenParams.Audience = atcc.audience
	// By default we will create "user-scoped token", unless specific groups or admin-privilege-instance were specified
	if len(atcc.groups) == 0 && !atcc.grantAdmin {
		atcc.groups = UserScopedNotation
	}
	if len(atcc.groups) > 0 {
		tokenParams.Scope = GroupsPrefix + atcc.groups
	}
	if atcc.grantAdmin {
		instanceId, err := getInstanceId(atcc.rtDetails)
		if err != nil {
			return tokenParams, err
		}
		if len(tokenParams.Scope) > 0 {
			tokenParams.Scope += " "
		}
		tokenParams.Scope += instanceId + AdminPrivilegesSuffix
	}

	return
}

func getInstanceId(rtDetails *config.ArtifactoryDetails) (string, error) {
	servicesManager, err := rtUtils.CreateServiceManager(rtDetails, false)
	if err != nil {
		return "", err
	}
	return servicesManager.CreateSystemService().GetArtifactoryServiceId()
}
