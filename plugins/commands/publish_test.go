package commands

import (
	"github.com/jfrog/jfrog-cli/plugins/commands/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOrderedArchitectures(t *testing.T) {
	localArc, err := utils.GetLocalArchitecture()
	if err != nil {
		assert.NoError(t, err)
		return
	}

	// Assert ordering successful.
	ordered, err := getOrderedArchitectures(localArc)
	if err != nil {
		assert.NoError(t, err)
		return
	}
	assert.Len(t, ordered, len(utils.ArchitecturesMap))
	assert.Equal(t, localArc, ordered[0])

	// Assert ordering fails for unsupported architecture.
	_, err = getOrderedArchitectures("made-up-arc")
	assert.Error(t, err)

}
