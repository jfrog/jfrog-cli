package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/artifactory/services/utils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"os"
	"strings"
)

type ApiCommand string

const (
	cliAiAskApiPath = "https://cli-ai-app.jfrog.info/api/ask"
	apiHeader       = "X-JFrog-CLI-AI"
)

type QuestionBody struct {
	Question string `json:"question"`
}

func HowCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() > 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	log.Output(coreutils.PrintTitle("This AI-based interface converts your natural language inputs into fully functional JFrog CLI commands.\n" +
		"NOTE: This is an experimental version and it supports mostly Artifactory and Xray commands.\n"))

	for {
		var question string
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("🐸 Your request:\n   ")
		for {
			// Ask the user for a question
			scanner.Scan()
			question = strings.TrimSpace(scanner.Text())
			if question != "" {
				// If the user entered a question, break the loop
				break
			}
		}
		fmt.Print("\n🤖 Generated command:\n   ")
		llmAnswer, err := askQuestion(question)
		if err != nil {
			return err
		}
		log.Output(coreutils.PrintLink(llmAnswer))
		log.Output("\n" + coreutils.PrintComment("-------------------") + "\n")
	}
}

func askQuestion(question string) (response string, err error) {
	contentBytes, err := json.Marshal(QuestionBody{Question: question})
	if errorutils.CheckError(err) != nil {
		return
	}
	client, err := httpclient.ClientBuilder().Build()
	if errorutils.CheckError(err) != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, cliAiAskApiPath, bytes.NewBuffer(contentBytes))
	if errorutils.CheckError(err) != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiHeader, "true")
	log.Debug(fmt.Sprintf("Sending HTTP %s request to: %s", req.Method, req.URL))
	resp, err := client.GetClient().Do(req)
	if err != nil {
		err = errorutils.CheckErrorf("CLI-AI server is not available. Please check your network or try again later.")
		return
	}
	if resp == nil {
		err = errorutils.CheckErrorf("received empty response from server")
		return
	}
	if err = errorutils.CheckResponseStatus(resp, http.StatusOK); err != nil {
		switch resp.StatusCode {
		case http.StatusInternalServerError:
			err = errorutils.CheckErrorf("CLI-AI model endpoint is not available. Please try again later.")
		case http.StatusNotAcceptable:
			err = errorutils.CheckErrorf("The system is currently handling multiple requests from other users\n" +
				"Please try submitting your question again in a few minutes. Thank you for your patience!")
		default:
			err = errorutils.CheckErrorf("CLI-AI server is not available. Please check your network or try again later. Note that the this command is supported while inside JFrog's internal network only.\n" + err.Error())
		}
		return
	}
	defer func() {
		if resp.Body != nil {
			err = errors.Join(err, errorutils.CheckError(resp.Body.Close()))
		}
	}()
	var body []byte
	// Limit size of response body to 10MB
	body, err = io.ReadAll(io.LimitReader(resp.Body, 10*utils.SizeMiB))
	if errorutils.CheckError(err) != nil {
		return
	}
	response = strings.TrimSpace(string(body))
	return
}
