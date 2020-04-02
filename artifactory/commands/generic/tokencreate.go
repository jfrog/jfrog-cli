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
	AdminPrivilegesPrefix = "jfrt@"
	AdminPrivilegesSuffix = ":admin"
	UserScopedNotation    = "*"
)

type TokenCreateCommand struct {
	rtDetails                 *config.ArtifactoryDetails
	refreshable               bool
	expiry                    int
	userName                  string
	audience                  string
	groups                    string
	adminPrivilegesInstanceId string
	response                  *services.CreateTokenResponseData
}

func NewTokenCreateCommand() *TokenCreateCommand {
	return &TokenCreateCommand{response: new(services.CreateTokenResponseData)}
}

func (tcc *TokenCreateCommand) SetRtDetails(rtDetails *config.ArtifactoryDetails) *TokenCreateCommand {
	tcc.rtDetails = rtDetails
	return tcc
}

func (tcc *TokenCreateCommand) SetRefreshable(refreshable bool) *TokenCreateCommand {
	tcc.refreshable = refreshable
	return tcc
}

func (tcc *TokenCreateCommand) SetExpiry(expiry int) *TokenCreateCommand {
	tcc.expiry = expiry
	return tcc
}

func (tcc *TokenCreateCommand) SetUserName(userName string) *TokenCreateCommand {
	tcc.userName = userName
	return tcc
}

func (tcc *TokenCreateCommand) SetAudience(audience string) *TokenCreateCommand {
	tcc.audience = audience
	return tcc
}

func (tcc *TokenCreateCommand) SetAdminPrivilegesInstanceId(instanceId string) *TokenCreateCommand {
	tcc.adminPrivilegesInstanceId = instanceId
	return tcc
}

func (tcc *TokenCreateCommand) SetGroups(groups string) *TokenCreateCommand {
	tcc.groups = groups
	return tcc
}

func (tcc *TokenCreateCommand) Response() ([]byte, error) {
	content, err := json.Marshal(*tcc.response)
	if errorutils.CheckError(err) != nil {
		return nil, err
	}
	return content, nil
}

func (tcc *TokenCreateCommand) RtDetails() (*config.ArtifactoryDetails, error) {
	return tcc.rtDetails, nil
}

func (tcc *TokenCreateCommand) CommandName() string {
	return "rt_token_create"
}

func (tcc *TokenCreateCommand) Run() error {
	servicesManager, err := rtUtils.CreateServiceManager(tcc.rtDetails, false)
	if err != nil {
		return err
	}
	tokenParams := tcc.getTokenParams()
	*tcc.response, err = servicesManager.CreateToken(tokenParams)

	return err
}

func (tcc *TokenCreateCommand) getTokenParams() (tokenParams services.CreateTokenParams) {
	tokenParams = services.NewCreateTokenParams()
	tokenParams.Username = tcc.userName
	tokenParams.ExpiresIn = tcc.expiry
	tokenParams.Refreshable = tcc.refreshable
	tokenParams.Audience = tcc.audience
	// By default we will create "user-scoped token", unless specific groups or admin-privilege-instance were specified
	if len(tcc.groups) == 0 && len(tcc.adminPrivilegesInstanceId) == 0 {
		tcc.groups = UserScopedNotation
	}
	if len(tcc.groups) > 0 {
		tokenParams.Scope = GroupsPrefix + tcc.groups
	}
	if len(tcc.adminPrivilegesInstanceId) > 0 {
		if len(tokenParams.Scope) > 0 {
			tokenParams.Scope += " "
		}
		tokenParams.Scope += AdminPrivilegesPrefix + tcc.adminPrivilegesInstanceId + AdminPrivilegesSuffix
	}

	return
}
