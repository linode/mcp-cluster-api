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

package mcptools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// NewGetMachineTool registers a tool to get a single CAPI machine by name and namespace.
func (m *ToolManager) NewGetMachineTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("get_machine",
			mcp.WithDescription("Get a single CAPI machine by name and namespace"),
			mcp.WithString("name", mcp.Required(), mcp.Description("The name of the machine to get")),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the machine")),
		),
		Handler:  m.HandleGetMachine,
		ReadOnly: true,
	}
}

// HandleGetMachine is the handler function for the get_machine tool.
func (m *ToolManager) HandleGetMachine(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	name, err := request.RequireString("name")
	if err != nil {
		return nil, err
	}
	namespace, err := request.RequireString("namespace")
	if err != nil {
		return nil, err
	}

	var machine capi.Machine
	if err := m.kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &machine); err != nil {
		return nil, err
	}

	machine.SetManagedFields(nil)
	machineBytes, err := yaml.Marshal(machine.DeepCopyObject())
	if err != nil {
		return nil, fmt.Errorf("marshal machine: %w", err)
	}

	return mcp.NewToolResultText(string(machineBytes)), nil
}
