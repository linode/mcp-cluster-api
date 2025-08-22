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
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

var caplListDescription = `
CAPI Resource Lister - Query CAPI resources in the cluster and return their details in YAML format.

Inputs:
- resource_type: The type of CAPI resource to list (e.g., clusters, machines, classes, machinedeployments, machinesets, machinehealthchecks)
- namespace: The namespace to list resources in. If not provided, resources from all namespaces will be listed.

Outputs:
- A YAML representation of the requested CAPI resources.

Example Usage:
- Prompt: List all clusters
  Response: (YAML output of all clusters in all the namespaces)

- Prompt: List all machines in the 'default' namespace
  Response: (YAML output of all machines in the 'default' namespace)

- Prompt: Retrieve all machinedeployments in the 'capi-system' namespace
  Response: (YAML output of all machinedeployments in the 'capi-system' namespace)

- Prompt: Show all machinehealthchecks across all namespaces
  Response: (YAML output of all machinehealthchecks in all namespaces)

- Prompt: Get all machineclasses in the 'production' namespace
  Response: (YAML output of all machineclasses in the 'production' namespace)

- Prompt: Can you list all machinesets
  Response: (YAML output of all machinesets in all the namespaces)
`

type CAPIResourceType struct {
	List *unstructured.UnstructuredList
	Get  *unstructured.Unstructured
}

func getCAPIResourceGVK(config *rest.Config, resource, operation string) (*schema.GroupVersionKind, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get discovery client: %w", err)
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))
	objGVR, err := mapper.ResourceFor(schema.GroupVersionResource{
		Resource: resource,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get GVR for %s: %w", resource, err)
	}

	objGVK, err := mapper.KindFor(objGVR)
	if err != nil {
		return nil, fmt.Errorf("failed to get GVK for %s: %w", resource, err)
	}
	if operation == "list" {
		objGVK.Kind += "List"
	}

	// ignore any object that is not part of the CAPI group
	if !strings.HasSuffix(objGVK.Group, "x-k8s.io") {
		return nil, fmt.Errorf("resource %s is not a CAPI resource", resource)
	}

	return &objGVK, nil
}

// removeManagedFields removes managedFields from all items in the ObjectList using reflection
func removeManagedFields(objList *unstructured.UnstructuredList) {
	for i := range objList.Items {
		objList.Items[i].SetManagedFields(nil)
	}
}

// getObjList retrieves the appropriate ObjectList based on the resource type in the request
func getObjList(ctx context.Context, cfg *rest.Config, kubeClient client.Client, request mcp.CallToolRequest) (*unstructured.UnstructuredList, error) {
	resourceType := request.GetString("resource_type", "")
	resourceObj := unstructured.UnstructuredList{}
	objGVK, err := getCAPIResourceGVK(cfg, resourceType, "list")
	if err != nil {
		return nil, err
	}
	resourceObj.SetGroupVersionKind(*objGVK)

	if err := kubeClient.List(ctx, &resourceObj, &client.ListOptions{Namespace: request.GetString("namespace", "")}); err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}

	removeManagedFields(&resourceObj)
	return &resourceObj, nil
}

func (m *ToolManager) NewListTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("list_resources",
			mcp.WithDescription("CAPI Resource Lister - Query CAPI resources in the cluster and return their details in YAML format"),
			mcp.WithString("resource_type", mcp.Description(caplListDescription)),
			mcp.WithString("namespace", mcp.Description("The namespace to list resources in")),
		),
		Handler:  m.HandleListResources,
		ReadOnly: true,
	}
}

// HandleListClusters is the handler function for the list_clusters tool.
func (m *ToolManager) HandleListResources(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	objList, err := getObjList(ctx, m.cfg, m.kubeClient, request)
	if err != nil {
		return nil, err
	}
	var sb strings.Builder
	removeManagedFields(objList)
	yamlData, err := yaml.Marshal(objList)
	if err != nil {
		return nil, fmt.Errorf("error marshalling %s list: %w", request.GetString("resource_type", ""), err)
	}
	sb.Write(yamlData)
	return mcp.NewToolResultText(sb.String()), nil
}
