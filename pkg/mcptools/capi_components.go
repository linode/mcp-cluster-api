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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewGetControlPlaneStatusTool registers a tool to get the status of control plane components
func (m *ToolManager) NewGetControlPlaneStatusTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("get_control_plane_status",
			mcp.WithDescription("Get status of control plane components for a CAPI cluster"),
			mcp.WithString("cluster", mcp.Required(), mcp.Description("The name of the cluster")),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the cluster")),
		),
		Handler:  m.HandleGetControlPlaneStatus,
		ReadOnly: true,
	}
}

// HandleGetControlPlaneStatus is the handler for the get_control_plane_status tool
func (m *ToolManager) HandleGetControlPlaneStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	clusterName, err := request.RequireString("cluster")
	if err != nil {
		return nil, err
	}

	namespace, err := request.RequireString("namespace")
	if err != nil {
		return nil, err
	}

	clientset, err := m.BuildClientForWorkload(ctx, clusterName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to build client for workload cluster: %w", err)
	}

	// Query control plane components directly from the workload cluster
	// List all pods in the kube-system namespace
	pods, err := clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component in (kube-apiserver,kube-controller-manager,kube-scheduler,etcd)",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list control plane pods: %w", err)
	}

	// Format the response
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Control Plane Status for Cluster: %s\n", clusterName))
	sb.WriteString("===================================\n\n")

	if len(pods.Items) == 0 {
		sb.WriteString("No control plane pods found for this cluster\n")
		return mcp.NewToolResultText(sb.String()), nil
	}

	for _, pod := range pods.Items {
		sb.WriteString(fmt.Sprintf("Pod: %s\n", pod.Name))
		sb.WriteString(fmt.Sprintf("  Status: %s\n", string(pod.Status.Phase)))
		sb.WriteString(fmt.Sprintf("  Node: %s\n", pod.Spec.NodeName))

		// Container statuses
		sb.WriteString("  Containers:\n")
		for _, containerStatus := range pod.Status.ContainerStatuses {
			sb.WriteString(fmt.Sprintf("    - %s: ", containerStatus.Name))
			if containerStatus.Ready {
				sb.WriteString("Ready")
			} else {
				sb.WriteString("Not Ready")
			}

			state := ""
			switch {
			case containerStatus.State.Waiting != nil:
				state = fmt.Sprintf(" (Waiting: %s)", containerStatus.State.Waiting.Reason)
			case containerStatus.State.Running != nil:
				state = fmt.Sprintf(" (Running since: %s)", containerStatus.State.Running.StartedAt)
			case containerStatus.State.Terminated != nil:
				state = fmt.Sprintf(" (Terminated: %s)", containerStatus.State.Terminated.Reason)
			}
			sb.WriteString(state)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}
