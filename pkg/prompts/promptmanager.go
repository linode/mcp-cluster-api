package prompts

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type PromptManager struct{}

func NewPromptManager() *PromptManager {
	return &PromptManager{}
}

type PromptHandler struct {
	Prompt  mcp.Prompt
	Handler server.PromptHandlerFunc
}

func (pm *PromptManager) RegisterPrompts(mcpServer *server.MCPServer) {
	prompts := []PromptHandler{
		pm.NewDebugCAPIClusterPrompt(),
		pm.NewDebugCapiMachinePrompt(),
		pm.NewRestartControlPlanePrompt(),
	}

	for _, prompt := range prompts {
		mcpServer.AddPrompt(prompt.Prompt, prompt.Handler)
	}
}
