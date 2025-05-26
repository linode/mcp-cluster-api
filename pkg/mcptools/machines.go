package mcptools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// NewListMachinesTool registers a tool to list CAPI machines in a namespace.
func (m *ToolManager) NewListMachinesTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("list_machines",
			mcp.WithDescription("Get CAPI machines in the cluster"),
			mcp.WithString("namespace", mcp.Description("The namespace to list machines in")),
		),
		Handler:  m.HandleListMachines,
		ReadOnly: true,
	}
}

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

// HandleListMachines is the handler function for the list_machines tool.
func (m *ToolManager) HandleListMachines(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	var machines capi.MachineList
	if err := m.kubeClient.List(ctx, &machines, &client.ListOptions{Namespace: request.GetString("namespace", "")}); err != nil {
		return nil, fmt.Errorf("list machines: %w", err)
	}

	var sb strings.Builder
	for _, machine := range machines.Items {
		machine.SetManagedFields(nil)
		machineBytes, err := yaml.Marshal(machine.DeepCopyObject())
		if err != nil {
			return nil, fmt.Errorf("marshal machine: %w", err)
		}
		sb.WriteString("---\n")
		sb.Write(machineBytes)
	}
	return mcp.NewToolResultText(sb.String()), nil
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
