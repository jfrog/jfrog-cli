package fish

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
)

func WriteFishCompletionScript(c *cli.Context, install bool) {
	jfApp := c.Parent().Parent().App
	fishAutocomplete, err := jfApp.ToFishCompletion()
	if err != nil {
		log.Error(err)
		return
	}
	if !install {
		fmt.Print(fishAutocomplete)
		return
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Error(err)
		return
	}
	completionPath := filepath.Join(homeDir, ".config", "fish", "completions", jfApp.Name+".fish")
	err = ioutil.WriteFile(completionPath, []byte(fishAutocomplete), 0600)
	if err != nil {
		log.Error(err)
	}
	fmt.Printf(`Generated fish completion script at %s`, completionPath)
}
