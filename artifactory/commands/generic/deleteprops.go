package generic

import (
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type DeletePropsCommand struct {
	PropsCommand
}

func NewDeletePropsCommand() *DeletePropsCommand {
	return &DeletePropsCommand{}
}

func (deleteProps *DeletePropsCommand) DeletePropsCommand(command PropsCommand) *DeletePropsCommand {
	deleteProps.PropsCommand = command
	return deleteProps
}

func (deleteProps *DeletePropsCommand) CommandName() string {
	return "rt_delete_properties"
}

func (deleteProps *DeletePropsCommand) Run() error {
	rtDetails, err := deleteProps.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}
	servicesManager, err := createPropsServiceManager(deleteProps.threads, rtDetails)
	if err != nil {
		return err
	}

	resultCr, searchErr := searchItems(deleteProps.Spec(), servicesManager)
	propsParams := GetPropsParams(resultCr, deleteProps.props)
	success, fail, err := servicesManager.DeleteProps(propsParams)
	result := deleteProps.Result()
	result.SetSuccessCount(success)
	result.SetFailCount(fail)
	if err == nil {
		return searchErr
	}
	return err
}
