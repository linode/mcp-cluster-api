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

func (pm *PromptManager) NewDebugCAPIClusterPrompt() PromptHandler {
	return PromptHandler{
		Prompt: mcp.NewPrompt("debug_capi_cluster",
			mcp.WithArgument("name", mcp.ArgumentDescription("The name of the cluster"), mcp.RequiredArgument()),
			mcp.WithArgument("namespace", mcp.ArgumentDescription("The namespace of the cluster"), mcp.RequiredArgument())),
		Handler: pm.NewDebugCAPIClusterPromptHandler,
	}
}

func (pm *PromptManager) NewDebugCAPIClusterPromptHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Extract the name and namespace from the request
	name, ok := request.Params.Arguments["name"]
	if !ok {
		return nil, errors.New("missing required argument: name")
	}
	namespace, ok := request.Params.Arguments["namespace"]
	if !ok {
		return nil, errors.New("missing required argument: namespace")
	}

	return mcp.NewGetPromptResult(
		"Debug Instructions for a CAPI Cluster",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(
				fmt.Sprintf(`
To debug the CAPI cluster named "%s" in the namespace "%s", follow these steps:
0. use the get_cluster tool to get the cluster object - Verify that the cluster exists, and has no errors.
1. Identify the cluster's controlplane using the spec.controlPlaneRef field.
2. Use the get_kube_resource tool to get the controlplane object - Verify that the controlplane object exists, has no errors. 
3. Identify the cluster's infrastructure provider using the spec.infrastructureRef field.
4. Use the get_kube_resource tool to get the infrastructure provider object - Verify that the infrastructure provider object exists, has no errors.
5. Use the get_control_plane_status tool to check the status of the control plane pods - This will provide information on the health of the control plane pods.
6. Write a detailed report on the cluster's status, including any errors or issues found during the investigation.
7. If the cluster controlplane is not healthy, follow instructions to restart the controlplane after getting user's consent.
`, name, namespace)))}), nil
}

func (pm *PromptManager) NewRestartControlPlanePrompt() PromptHandler {
	return PromptHandler{
		Prompt: mcp.NewPrompt("restart_control_plane",
			mcp.WithArgument("name", mcp.ArgumentDescription("The name of the cluster"), mcp.RequiredArgument()),
			mcp.WithArgument("namespace", mcp.ArgumentDescription("The namespace of the cluster"), mcp.RequiredArgument())),
		Handler: pm.NewRestartControlPlanePromptHandler,
	}
}

func (pm *PromptManager) NewRestartControlPlanePromptHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Extract the name and namespace from the request
	name, ok := request.Params.Arguments["name"]
	if !ok {
		return nil, errors.New("missing required argument: name")
	}
	namespace, ok := request.Params.Arguments["namespace"]
	if !ok {
		return nil, errors.New("missing required argument: namespace")
	}

	return mcp.NewGetPromptResult(
		"Restart Instructions for a CAPI Cluster's control plane",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(mcp.RoleAssistant, mcp.NewTextContent(
				fmt.Sprintf(`
To restart the control plane of the CAPI cluster named "%s" in the namespace "%s", follow these steps:
0. use the get_cluster tool to get the cluster object - Verify that the cluster exists, and has no errors.
1. Identify the cluster's controlplane using the spec.controlPlaneRef field.
2. Use the get_kube_resource tool to get the controlplane object - Verify that the controlplane object exists, has no errors.
3. Use the rollout_controlplane tool to restart the control plane - This will trigger a rolling update of the control plane machines. ALWAYS get user's consent before proceeding.
4. Monitor the status of the control plane machines to ensure they are healthy and ready - use the get_machine tool to get the machine objects as needed.
5. Once the machines are healthy, verify the control plane pods are running correctly - use the get_control_plane_status tool to check the status of the control plane pods.
6. Write a detailed report on the control plane's status, including any errors or issues found during the investigation.
`, name, namespace)))}), nil
}
