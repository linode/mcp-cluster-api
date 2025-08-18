/*
Copyright 2025 Akamai Technologies, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
