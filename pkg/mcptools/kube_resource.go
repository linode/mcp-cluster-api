package mcptools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func (m *ToolManager) NewGetKubeResourceTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("get_kube_resource",
			mcp.WithDescription("Get a specific Kubernetes resource"),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the resource")),
			mcp.WithString("name", mcp.Required(), mcp.Description("The name of the resource")),
			mcp.WithString("kind", mcp.Required(), mcp.Description("The kind of the resource")),
			mcp.WithString("apiVersion", mcp.Required(), mcp.Description("The API version of the resource")),
		),
		Handler:  m.HandleGetKubeResource,
		ReadOnly: true,
	}
}

func (m *ToolManager) HandleGetKubeResource(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters from the request
	namespace, err := request.RequireString("namespace")
	if err != nil {
		return nil, err
	}
	name, err := request.RequireString("name")
	if err != nil {
		return nil, err
	}
	kind, err := request.RequireString("kind")
	if err != nil {
		return nil, err
	}
	apiVersion, err := request.RequireString("apiVersion")
	if err != nil {
		return nil, err
	}

	var resource unstructured.Unstructured
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}

	resource.SetGroupVersionKind(gv.WithKind(kind))

	err = m.kubeClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &resource)
	if err != nil {
		return nil, err
	}
	resource.SetManagedFields(nil) // Remove managed fields to avoid cluttering the output

	resourceBytes, err := yaml.Marshal(resource.DeepCopyObject())
	if err != nil {
		return nil, err
	}
	resourceString := string(resourceBytes)
	return mcp.NewToolResultText(resourceString), nil
}
