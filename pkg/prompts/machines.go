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
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// NewDebugCapiMachinePrompt returns a prompt for debugging a CAPI Machine
func (pm *PromptManager) NewDebugCapiMachinePrompt() PromptHandler {
	return PromptHandler{
		Prompt: mcp.NewPrompt("debug_capi_machine",
			mcp.WithArgument("name", mcp.ArgumentDescription("The name of the machine"), mcp.RequiredArgument()),
			mcp.WithArgument("namespace", mcp.ArgumentDescription("The namespace of the machine"), mcp.RequiredArgument())),
		Handler: pm.NewDebugCapiMachinePromptHandler,
	}
}

func (pm *PromptManager) NewDebugCapiMachinePromptHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	name, ok := request.Params.Arguments["name"]
	if !ok {
		return nil, errors.New("missing required argument: name")
	}
	namespace, ok := request.Params.Arguments["namespace"]
	if !ok {
		return nil, errors.New("missing required argument: namespace")
	}
	return mcp.NewGetPromptResult(
		"Debug Instructions for a CAPI Machine",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(
				fmt.Sprintf(`
To debug the CAPI machine named "%s" in the namespace "%s", follow these steps:
0. use the get_machine tool to get the machine object - Verify that the machine exists, and has no errors.
1. Identify the machine's bootstrap configuration using the spec.bootstrap.configRef field.
2. Use the get_kube_resource tool to get the bootstrap config object - Verify that the bootstrap config object exists, has no errors.
3. Identify the machine's infrastructure provider using the spec.infrastructureRef field.
4. Use the get_kube_resource tool to get the infrastructure provider object - Verify that the infrastructure provider object exists, has no errors.
5. Write a detailed report on the machine's status, including any errors or issues found during the investigation.
`, name, namespace)))}), nil
}
