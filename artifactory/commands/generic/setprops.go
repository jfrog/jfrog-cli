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

	resultItems, searchErr := searchItems(setProps.Spec(), servicesManager)

	propsParams := GetPropsParams(resultItems, setProps.props)
	success, fails, err := servicesManager.SetProps(propsParams)

	result := setProps.Result()
	result.SetSuccessCount(success)
	result.SetFailCount(fails)
	if err == nil {
		return searchErr
	}
	return err
}
