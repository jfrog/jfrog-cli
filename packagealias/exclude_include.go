package packagealias

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

// ExcludeCommand excludes a tool from Ghost Frog interception
type ExcludeCommand struct {
	tool string
}

func NewExcludeCommand(tool string) *ExcludeCommand {
	return &ExcludeCommand{tool: tool}
}

func (ec *ExcludeCommand) CommandName() string {
	return "package_alias_exclude"
}

func (ec *ExcludeCommand) Run() error {
	return setToolMode(ec.tool, ModePass)
}

func (ec *ExcludeCommand) SetRepo(repo string) *ExcludeCommand {
	return ec
}

func (ec *ExcludeCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}

// IncludeCommand includes a tool in Ghost Frog interception
type IncludeCommand struct {
	tool string
}

func NewIncludeCommand(tool string) *IncludeCommand {
	return &IncludeCommand{tool: tool}
}

func (ic *IncludeCommand) CommandName() string {
	return "package_alias_include"
}

func (ic *IncludeCommand) Run() error {
	return setToolMode(ic.tool, ModeJF)
}

func (ic *IncludeCommand) SetRepo(repo string) *IncludeCommand {
	return ic
}

func (ic *IncludeCommand) ServerDetails() (*config.ServerDetails, error) {
	return nil, nil
}

// setToolMode sets the mode for a specific tool
func setToolMode(tool string, mode AliasMode) error {
	// Validate tool name
	tool = strings.ToLower(tool)
	isValid := false
	for _, supportedTool := range SupportedTools {
		if tool == supportedTool {
			isValid = true
			break
		}
	}
	if !isValid {
		return errorutils.CheckError(fmt.Errorf("unsupported tool: %s. Supported tools: %s", tool, strings.Join(SupportedTools, ", ")))
	}

	// Validate mode
	if mode != ModeJF && mode != ModeEnv && mode != ModePass {
		return errorutils.CheckError(fmt.Errorf("invalid mode: %s. Valid modes: jf, env, pass", mode))
	}

	aliasDir, err := GetAliasHomeDir()
	if err != nil {
		return err
	}

	// Check if aliases are installed
	binDir := filepath.Join(aliasDir, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return errorutils.CheckError(fmt.Errorf("package aliases are not installed. Run 'jf package-alias install' first"))
	}

	// Load existing config or create new one
	configPath := filepath.Join(aliasDir, configFile)
	cfg := &Config{
		ToolModes: make(map[string]AliasMode),
		Enabled:   true,
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			log.Warn(fmt.Sprintf("Failed to parse existing config, creating new one: %v", err))
		}
	}

	// Update tool mode
	cfg.ToolModes[tool] = mode

	// Save config
	jsonData, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errorutils.CheckError(err)
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return errorutils.CheckError(err)
	}

	// Show result
	modeDescription := map[AliasMode]string{
		ModeJF:  "intercepted by JFrog CLI",
		ModeEnv: "run natively with environment injection",
		ModePass: "run natively (excluded from interception)",
	}

	log.Info(fmt.Sprintf("Tool '%s' is now configured to: %s", tool, modeDescription[mode]))
	log.Info(fmt.Sprintf("Mode: %s", mode))
	
	if mode == ModePass {
		log.Info(fmt.Sprintf("When you run '%s', it will execute the native tool directly without JFrog CLI interception.", tool))
	} else if mode == ModeJF {
		log.Info(fmt.Sprintf("When you run '%s', it will be intercepted and run as 'jf %s'.", tool, tool))
	}

	return nil
}

