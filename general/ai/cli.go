package ai

import (
	"bufio"
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
	"os"
	"strings"
)

type ApiCommand string

const (
	cliAiApiPath            = "https://cli-ai-app.jfrog.info/"
	questionApi  ApiCommand = "ask"
	feedbackApi  ApiCommand = "feedback"
)

type QuestionBody struct {
	Question string `json:"question"`
}

type FeedbackBody struct {
	QuestionBody
	LlmAnswer      string `json:"llm_answer"`
	IsAccurate     bool   `json:"is_accurate"`
	ExpectedAnswer string `json:"expected_answer"`
}

func HowCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() > 1 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	log.Output(coreutils.PrintTitle("This AI-based interface converts your natural language inputs into fully functional JFrog CLI commands.\n" +
		"NOTE: This is a beta version and it supports mostly Artifactory and Xray commands.\n"))

	for {
		var question string
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("🐸 Your Request: ")
		for {
			// Ask the user for a question
			scanner.Scan()
			question = scanner.Text()
			if question != "" {
				// If the user entered a question, break the loop
				break
			}
		}
		questionBody := QuestionBody{Question: strings.TrimSpace(question)}
		llmAnswer, err := askQuestion(QuestionBody{Question: question})
		if err != nil {
			return err
		}
		if strings.ToLower(llmAnswer) == "i dont know" {
			log.Output("The current version of the AI model does not support this type of command yet.\n")
			break
		}
		log.Output("🤖 Generated Command: " + coreutils.PrintLink(llmAnswer))
		log.Output()
		feedback := FeedbackBody{QuestionBody: questionBody, LlmAnswer: llmAnswer}
		feedback.getUserFeedback()
		if err = sendFeedback(feedback); err != nil {
			return err
		}
		log.Output()
	}
	return nil
}

func (fb *FeedbackBody) getUserFeedback() {
	fb.IsAccurate = coreutils.AskYesNo("Is the provided command accurate?", true)
	if !fb.IsAccurate {
		ioutils.ScanFromConsole("Please provide the exact command you expected (Example: 'jf rt u ...')", &fb.ExpectedAnswer, "")
	}
}

func askQuestion(question QuestionBody) (response string, err error) {
	return sendRequestToCliAiServer(question, questionApi)
}

func sendFeedback(feedback FeedbackBody) (err error) {
	_, err = sendRequestToCliAiServer(feedback, feedbackApi)
	return
}

func sendRequestToCliAiServer(content interface{}, apiCommand ApiCommand) (response string, err error) {
	contentBytes, err := json.Marshal(content)
	if errorutils.CheckError(err) != nil {
		return
	}
	client, err := httpclient.ClientBuilder().Build()
	if errorutils.CheckError(err) != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, cliAiApiPath+string(apiCommand), bytes.NewBuffer(contentBytes))
	if errorutils.CheckError(err) != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	log.Debug(fmt.Sprintf("Sending HTTP %s request to: %s", req.Method, req.URL))
	resp, err := client.GetClient().Do(req)
	if errorutils.CheckError(err) != nil {
		return
	}
	if resp == nil {
		err = errorutils.CheckErrorf("received empty response from server")
		return
	}
	if err = errorutils.CheckResponseStatus(resp, http.StatusOK); err != nil {
		if resp.StatusCode == http.StatusInternalServerError {
			err = errorutils.CheckErrorf("AI model Endpoint is not available.\n" + err.Error())
		} else if resp.StatusCode == http.StatusNotFound {
			err = errorutils.CheckErrorf("CLI-AI app server is no available. Note that the this command is supported while inside JFrog's internal network only.\n" + err.Error())
		}
		return
	}
	if apiCommand == questionApi {
		defer func() {
			if resp.Body != nil {
				err = errors.Join(err, errorutils.CheckError(resp.Body.Close()))
			}
		}()
		var body []byte
		body, err = io.ReadAll(resp.Body)
		if errorutils.CheckError(err) != nil {
			return
		}
		response = strings.TrimSpace(string(body))
	}
	return
}
