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
	reader, err := searchItems(deleteProps.Spec(), servicesManager)
	if err != nil {
		return err
	}
	defer reader.Close()
	propsParams := GetPropsParams(reader, deleteProps.props)
	success, err := servicesManager.DeleteProps(propsParams)
	result := deleteProps.Result()
	result.SetSuccessCount(success)
	totalLength, totalLengthErr := reader.Length()
	result.SetFailCount(totalLength - success)
	if totalLengthErr != nil {
		return totalLengthErr
	}
	return err
}
