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
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"os"
	"strings"
)

type ApiCommand string

const (
	cliAiAppApiUrl     = "https://cli-ai-app.jfrog.io/api/"
	askRateLimitHeader = "X-JFrog-CLI-AI"
	// The latest version of the terms and conditions for using the AI interface. (https://docs.jfrog-applications.jfrog.io/jfrog-applications/jfrog-cli/cli-ai/terms)
	aiTermsRevision = 1
)

type ApiType string

const (
	ask      ApiType = "ask"
	feedback ApiType = "feedback"
)

func HowCmd(c *cli.Context) error {
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	if c.NArg() > 0 {
		return cliutils.WrongNumberOfArgumentsHandler(c)
	}
	log.Output(coreutils.PrintLink("This AI-based interface converts natural language inputs into AI-generated JFrog CLI commands.\n" +
		"For more information about this interface, see https://docs.jfrog-applications.jfrog.io/jfrog-applications/jfrog-cli/cli-ai\n" +
		"NOTE: This is an experimental version and it supports mostly Artifactory and Xray commands.\n"))

	// Ask the user to agree to the terms and conditions. If the user does not agree, the command will not proceed.
	// Ask this only once per JFrog CLI installation, unless the terms are updated.
	if agreed, err := handleAiTermsAgreement(); err != nil || !agreed {
		return err
	}

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
		fmt.Print("\n🤖 Generated command:\n")
		llmAnswer, err := askQuestion(question)
		if err != nil {
			return err
		}
		// Print the generated command within a styled table frame.
		coreutils.PrintMessageInsideFrame(coreutils.PrintBoldTitle(llmAnswer), "   ")

		log.Output()
		if err = sendFeedback(); err != nil {
			return err
		}

		log.Output("\n" + coreutils.PrintComment("-------------------") + "\n")
	}
}

type questionBody struct {
	Question string `json:"question"`
}

func askQuestion(question string) (response string, err error) {
	return sendRestAPI(ask, questionBody{Question: question})
}

type feedbackBody struct {
	IsGoodResponse bool `json:"is_good_response"`
}

func sendFeedback() (err error) {
	isGoodResponse, err := getUserFeedback()
	if err != nil {
		return err
	}
	_, err = sendRestAPI(feedback, feedbackBody{IsGoodResponse: isGoodResponse})
	return err
}

func getUserFeedback() (bool, error) {
	// Customize the template to place the options on the same line as the question
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   " 👉 {{ . | cyan  }}",
		Inactive: "    {{ . }}",
		Selected: "🙏 Thanks for your feedback!",
	}

	prompt := promptui.Select{
		Label:     "⭐ Rate this response:",
		Items:     []string{"👍 Good response!", "👎 Could be better..."},
		Templates: templates,
		HideHelp:  true,
	}
	selected, _, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return selected == 0, nil
}

func sendRestAPI(apiType ApiType, content interface{}) (response string, err error) {
	contentBytes, err := json.Marshal(content)
	if errorutils.CheckError(err) != nil {
		return
	}
	client, err := httpclient.ClientBuilder().Build()
	if errorutils.CheckError(err) != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, cliAiAppApiUrl+string(apiType), bytes.NewBuffer(contentBytes))
	if errorutils.CheckError(err) != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if apiType == ask {
		req.Header.Set(askRateLimitHeader, "true")
	}
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

	if apiType == feedback {
		// If the API is feedback, no response is expected
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

func handleAiTermsAgreement() (bool, error) {
	latestTermsVer, err := cliutils.GetLatestAiTermsRevision()
	if err != nil {
		return false, err
	}
	if latestTermsVer == nil || *latestTermsVer < aiTermsRevision {
		if !coreutils.AskYesNo("By using this interface, you agree to the terms of JFrog's AI Addendum on behalf of your organization as an active JFrog customer.\n"+
			"Review these terms at "+coreutils.PrintLink("https://docs.jfrog-applications.jfrog.io/jfrog-applications/jfrog-cli/cli-ai/terms")+
			"\nDo you agree?", false) {
			return false, nil
		}
		if err = cliutils.SetLatestAiTermsRevision(aiTermsRevision); err != nil {
			return false, err
		}
		log.Output()
	}
	return true, nil
}
