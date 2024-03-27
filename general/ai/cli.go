package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-cli-core/v2/utils/ioutils"
	"github.com/jfrog/jfrog-cli/utils/cliutils"
	"github.com/jfrog/jfrog-client-go/http/httpclient"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"os/user"
	"strings"
)

type ApiCommand string

const (
	cliAiApiPath            = "https://cli-ai.jfrog.info/"
	questionApi  ApiCommand = "ask"
	feedbackApi  ApiCommand = "feedback"
)

type questionBody struct {
	Question string `json:"question"`
}
type feedbackBody struct {
	questionBody
	LlmAnswer      string `json:"llm_answer"`
	IsAccurate     bool   `json:"is_accurate"`
	ExpectedAnswer string `json:"expected_answer"`
	JfrogUser      string `json:"jfrog_user"`
}

func HowCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() < 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}

	args := cliutils.ExtractCommand(c)
	question := questionBody{Question: fmt.Sprintf("How %s", strings.Join(args, " "))}
	llmAnswer, err := askQuestion(question)
	if err != nil {
		return err
	}
	log.Output("AI generated JFrog CLI command:")
	err = coreutils.PrintTable("", "", coreutils.PrintTitle(llmAnswer), false)
	if err != nil {
		return err
	}

	feedback := feedbackBody{questionBody: question, LlmAnswer: llmAnswer}
	feedback.getUserFeedback()
	err = sendFeedback(feedback)
	if err != nil {
		return err
	}
	log.Output("Thank you for your feedback!")
	return nil
}

func (fb *feedbackBody) getUserFeedback() {
	currentUser, err := user.Current()
	if err != nil {
		log.Warn("Failed to get current JFrog user name")
	}
	fb.JfrogUser = currentUser.Name
	fb.IsAccurate = coreutils.AskYesNo(coreutils.PrintLink("Is the provided command accurate?"), true)
	if !fb.IsAccurate {
		for {
			ioutils.ScanFromConsole("Please provide the exact command you expected", &fb.ExpectedAnswer, "")
			if strings.HasPrefix(fb.ExpectedAnswer, "jf ") {
				break
			} else {
				log.Output("Please provide a valid JFrog CLI command that start with jf. (Example: 'jf rt u ...')")
			}
		}
	}
}

func askQuestion(question questionBody) (response string, err error) {
	return sendRequestToCliAiServer(question, questionApi)
}

func sendFeedback(feedback feedbackBody) (err error) {
	_, err = sendRequestToCliAiServer(feedback, feedbackApi)
	return
}

func sendRequestToCliAiServer(content interface{}, apiCommand ApiCommand) (response string, err error) {
	contentBytes, err := json.Marshal(content)
	if err != nil {
		return
	}
	client, err := httpclient.ClientBuilder().Build()
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, cliAiApiPath+string(apiCommand), bytes.NewBuffer(contentBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	log.Debug(fmt.Sprintf("Sending HTTP %s request to: %s", req.Method, req.URL))
	resp, err := client.GetClient().Do(req)
	if err != nil {
		return
	}
	if resp == nil {
		err = errors.New("received empty response from server")
		return
	}
	if err = errorutils.CheckResponseStatus(resp, http.StatusOK); err != nil {
		if resp.StatusCode == http.StatusInternalServerError {
			err = errors.New("AI model Endpoint is not available.\n" + err.Error())
		} else if resp.StatusCode == http.StatusNotFound {
			err = errors.New("CLI-AI app server is no available. Please notice that the above command are support from inside JFrog internal network only.\n" + err.Error())
		}
		return
	}
	if apiCommand == questionApi {
		defer func() {
			if resp.Body != nil {
				err = errors.Join(err, resp.Body.Close())
			}
		}()
		body, _ := io.ReadAll(resp.Body)
		response = string(body)
	}
	return
}
