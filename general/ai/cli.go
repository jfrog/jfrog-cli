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
	cliAiApiPath            = "https://cli-ai-app.jfrog.info/"
	apiPrefix               = "api/"
	questionApi  ApiCommand = apiPrefix + "ask"
	feedbackApi  ApiCommand = apiPrefix + "feedback"
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
	if c.NArg() > 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	log.Output(coreutils.PrintTitle("This AI-based interface converts your natural language inputs into fully functional JFrog CLI commands.\n" +
		"NOTE: This is a beta version and it supports mostly Artifactory and Xray commands.\n"))

	for {
		var question string
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("🐸 Your request: ")
		for {
			// Ask the user for a question
			scanner.Scan()
			question = strings.TrimSpace(scanner.Text())
			if question != "" {
				// If the user entered a question, break the loop
				break
			}
		}
		questionBody := QuestionBody{Question: question}
		llmAnswer, err := askQuestion(questionBody)
		if err != nil {
			return err
		}

		log.Output("🤖 Generated command: " + coreutils.PrintLink(llmAnswer) + "\n")
		feedback := FeedbackBody{QuestionBody: questionBody, LlmAnswer: llmAnswer}
		feedback.getUserFeedback()
		if err = sendFeedback(feedback); err != nil {
			return err
		}
		log.Output("\n" + coreutils.PrintComment("-------------------") + "\n")
	}
}

func (fb *FeedbackBody) getUserFeedback() {
	fb.IsAccurate = coreutils.AskYesNo("Is the provided command accurate?", true)
	if !fb.IsAccurate {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("Please provide the exact command you expected (Example: 'jf rt u ...'): ")
		for {
			scanner.Scan()
			expectedAnswer := strings.TrimSpace(scanner.Text())
			if expectedAnswer != "" {
				// If the user entered an expected answer, break and return
				fb.ExpectedAnswer = expectedAnswer
				return
			}
		}
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
		// Limit size of response body to 10MB
		body, err = io.ReadAll(io.LimitReader(resp.Body, 10*utils.SizeMiB))
		if errorutils.CheckError(err) != nil {
			return
		}
		response = strings.TrimSpace(string(body))
	}
	return
}
