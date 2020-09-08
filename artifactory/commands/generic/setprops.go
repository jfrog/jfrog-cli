package generic

import (
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

type SetPropsCommand struct {
	PropsCommand
}

func NewSetPropsCommand() *SetPropsCommand {
	return &SetPropsCommand{}
}

func (setProps *SetPropsCommand) SetPropsCommand(command PropsCommand) *SetPropsCommand {
	setProps.PropsCommand = command
	return setProps
}

func (setProps *SetPropsCommand) CommandName() string {
	return "rt_set_properties"
}

func (setProps *SetPropsCommand) Run() error {
	rtDetails, err := setProps.RtDetails()
	if errorutils.CheckError(err) != nil {
		return err
	}
	servicesManager, err := createPropsServiceManager(setProps.threads, rtDetails)
	if err != nil {
		return err
	}

	reader, searchErr := searchItems(setProps.Spec(), servicesManager)
	if searchErr != nil {
		return searchErr
	}
	defer reader.Close()
	propsParams := GetPropsParams(reader, setProps.props)
	success, err := servicesManager.SetProps(propsParams)

	result := setProps.Result()
	result.SetSuccessCount(success)
	totalLength, totalLengthErr := reader.Length()
	result.SetFailCount(totalLength - success)
	if totalLengthErr != nil {
		return totalLengthErr
	}
	return err
}
