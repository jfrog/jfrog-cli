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
	clientutils "github.com/jfrog/jfrog-client-go/utils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"os"
	"os/exec"
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
	log.Output(coreutils.PrintLink("This AI-powered interface converts natural language inputs into AI-generated JFrog CLI commands.\n" +
		"For more information about this interface, see https://docs.jfrog-applications.jfrog.io/jfrog-applications/jfrog-cli/cli-ai\n" +
		"Try it out by typing a question, such as: 'How can I upload all .zip files from user/mylibs/ to the libs-local repository in Artifactory?'\n" +
		"Note: JFrog AI Assistant is in beta and currently supports primarily Artifactory and Xray commands.\n"))

	// Ask the user to agree to the terms and conditions. If the user does not agree, the command will not proceed.
	// Ask this only once per JFrog CLI installation, unless the terms are updated.
	if agreed, err := handleAiTermsAgreement(); err != nil {
		return err
	} else if !agreed {
		// If the user does not agree to the terms, the command will not proceed.
		return reportTermsDisagreement()
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
		validResponse := strings.HasPrefix(llmAnswer, "jf")
		// Print the generated command within a styled table frame.
		if validResponse {
			coreutils.PrintMessageInsideFrame(coreutils.PrintBoldTitle(llmAnswer), "   ")
		} else {
			log.Output("   " + coreutils.PrintYellow(llmAnswer))
		}

		// If the response is a valid JFrog CLI command, ask the user for feedback.
		if validResponse {
			log.Output()
			if err = handleResponseFeedback(); err != nil {
				return err
			}
		}

		log.Output("\n" + coreutils.PrintComment("-------------------") + "\n")
	}
}

func McpCmd(c *cli.Context) error {
	// Show help if needed
	if show, err := cliutils.ShowCmdHelpIfNeeded(c, c.Args()); show || err != nil {
		return err
	}
	//log.Output(coreutils.PrintBoldTitle("Welcome to the MCP-Start Command! 🚀"))
	// TODO this should downloaded from releases and version should be a variable
	//log.Output("Download MCP server binary version : v0.0.1 ... ")
	// TODO need to decide where the executable is being downloaded..maybe the current dir is okay.
	exePath, err := downloadServerExecutable()
	if err != nil {
		return err
	}
	//log.Output(fmt.Sprintf("✅ Successfully downloaded the MCP server binary to: %s", exePath))

	//binaryPath, err := resolveOrDownloadMcpBinary()
	//if err != nil {
	//	return fmt.Errorf("failed to get MCP binary: %w", err)
	//}

	cmd := exec.Command(exePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
	//
	//client, err := promptClientSelection()
	//if err != nil {
	//	return err
	//}
	//log.Output(fmt.Sprintf("You selected: %s", client))
	//
	//// Output the corresponding template
	//if err = outputClientTemplate(client, exePath); err != nil {
	//	return err
	//}
	//
	//log.Output("✅ Successfully completed the `mcp-start` process!\n")
	//log.Output("ℹ️ For further assistance, questions, or issues, please visit the repository: https://github.com/jfrog/mcp-jfrog-go")
	//return nil
}

func outputClientTemplate(client, path string) error {
	templates := map[string]string{
		"Cursor IDE": fmt.Sprintf(`Add the following to your ~/.cursor/mcp.json file:

{
  "mcpServers": {
    "jfrog-cli-mcp-server": {
      "command": "%s",
      "capabilities": {
        "tools": true
      }
    }
  }
}`, path),
		"VsCode": fmt.Sprintf(`Add the following to your VS Code settings (usually in settings.json):

{
  "mcp": {
    "servers": {
      "JFrog-Cli": {
        "type": "stdio",
        "command": "%s",
        }
      }
    }
  }
}`, path),
		"Calude": fmt.Sprintf(`Add the following to your VS Code settings (usually in settings.json):

{
  "mcpServers": {
    "JFrog-MCP-Server": {
      "command": "%s",
      "capabilities": {
        "tools": true
      }
    }
  }
}`, path),
	}

	template, exists := templates[client]
	if !exists {
		return nil
	}

	log.Output(coreutils.PrintBoldTitle("Configuration Template:"))
	log.Output(template)
	return nil
}

func promptClientSelection() (string, error) {
	clients := []string{"Calude", "VsCode", "Cursor IDE", "others"}
	prompt := promptui.Select{
		Label: "Select your client",
		Items: clients,
	}
	_, client, err := prompt.Run()
	return client, err
}

func downloadServerExecutable() (string, error) {
	//binaryName := "mcp-jfrog-go"
	// TODO this has to point to latest
	repoPath := "v0/0.0.1"

	targetDir, err := coreutils.GetJfrogHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get JFrog home directory: %w", err)
	}

	// Create the target directory if it doesn't exist
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.Mkdir(targetDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create directory '%s': %w", targetDir, err)
		}
	}

	// Change into the target directory
	if err := os.Chdir(targetDir); err != nil {
		return "", fmt.Errorf("failed to cd into directory '%s': %w", targetDir, err)
	}

	// Construct the full path for the binary
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Run the JFrog CLI download command
	cmd := exec.Command("jf", "rt", "dl", targetDir, repoPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}
	fullPath := fmt.Sprintf("%s/%s", wd, repoPath)
	// Make the binary executable
	if err := os.Chmod(fullPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Return the full path of the binary
	return fullPath, nil
}

type questionBody struct {
	Question string `json:"question"`
}

func askQuestion(question string) (response string, err error) {
	return sendRestAPI(ask, questionBody{Question: question})
}

type feedbackBody struct {
	IsGoodResponse *bool `json:"is_good_response,omitempty"`
	IsAgreedTerms  *bool `json:"is_agreed_terms,omitempty"`
}

func handleResponseFeedback() (err error) {
	isGoodResponse, err := getUserFeedback()
	if err != nil {
		return
	}
	_, err = sendRestAPI(feedback, feedbackBody{IsGoodResponse: &isGoodResponse})
	return
}

func reportTermsDisagreement() (err error) {
	_, err = sendRestAPI(feedback, feedbackBody{IsAgreedTerms: clientutils.Pointer(false)})
	return
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
			err = errorutils.CheckErrorf("JFrog CLI-AI model endpoint is not available. Please try again later.")
		case http.StatusNotAcceptable:
			err = errorutils.CheckErrorf("The system is currently handling multiple requests from other users\n" +
				"Please try submitting your question again in a few minutes. Thank you for your patience!")
		default:
			err = errorutils.CheckErrorf("JFrog CLI-AI server is not available. Please check your network or try again later:\n" + err.Error())
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
