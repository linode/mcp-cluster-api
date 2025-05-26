package mcptools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/yaml"
)

func (m *ToolManager) NewListClustersTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("list_clusters",
			mcp.WithDescription("Get CAPI clusters in the cluster"),
			mcp.WithString("namespace", mcp.Description("The namespace to list clusters in")),
		),
		Handler:  m.HandleListClusters,
		ReadOnly: true,
	}
}

// NewGetClusterTool registers a tool to get a single CAPI cluster by name and (optionally) namespace.
func (m *ToolManager) NewGetClusterTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("get_cluster",
			mcp.WithDescription("Get a single CAPI cluster by name and namespace"),
			mcp.WithString("name", mcp.Required(), mcp.Description("The name of the cluster to get")),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the cluster (optional)")),
		),
		Handler:  m.HandleGetCluster,
		ReadOnly: true,
	}
}

// HandleListClusters is the handler function for the list_clusters tool.
func (m *ToolManager) HandleListClusters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	var clusters capi.ClusterList
	if err := m.kubeClient.List(ctx, &clusters, &client.ListOptions{Namespace: request.GetString("namespace", "")}); err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}

	var sb strings.Builder
	for _, cluster := range clusters.Items {
		// Remove managed fields to avoid cluttering the output
		cluster.SetManagedFields(nil)
		clusterBytes, err := yaml.Marshal(cluster.DeepCopyObject())
		if err != nil {
			return nil, fmt.Errorf("marshal cluster: %w", err)
		}
		sb.WriteString("---\n")
		sb.Write(clusterBytes)
	}
	return mcp.NewToolResultText(sb.String()), nil
}

// HandleGetCluster is the handler function for the get_cluster tool.
func (m *ToolManager) HandleGetCluster(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	var cluster capi.Cluster
	if err := m.kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cluster); err != nil {
		return nil, err
	}

	// Remove managed fields to avoid cluttering the output
	cluster.SetManagedFields(nil)
	clusterBytes, err := yaml.Marshal(cluster.DeepCopyObject())
	if err != nil {
		return nil, fmt.Errorf("marshal cluster: %w", err)
	}

	return mcp.NewToolResultText(string(clusterBytes)), nil
}

func (m *ToolManager) NewRolloutControlPlaneTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("rollout_controlplane",
			mcp.WithDescription("Rollout a restart of control plane of a CAPI cluster. This triggers a rolling update of the control plane machines, new machines will be created and old machines will be deleted."),
			mcp.WithString("name", mcp.Required(), mcp.Description("The name of the cluster to rollout")),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the cluster")),
		),
		Handler:  m.HandleRolloutControlPlane,
		ReadOnly: false,
	}
}

func (m *ToolManager) HandleRolloutControlPlane(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Get the cluster
	var cluster capi.Cluster
	if err := m.kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cluster); err != nil {
		return nil, err
	}

	if cluster.Spec.ControlPlaneRef == nil {
		return nil, fmt.Errorf("cluster %s/%s does not have a control plane reference", name, namespace)
	}

	if cluster.Spec.ControlPlaneRef.Kind != "KubeadmControlPlane" {
		return nil, fmt.Errorf("cluster %s/%s control plane reference is not a KubeadmControlPlane. Rollout is unsupported", name, namespace)
	}

	var controlPlane unstructured.Unstructured
	controlPlane.SetGroupVersionKind(cluster.Spec.ControlPlaneRef.GroupVersionKind())

	// Get the control plane object
	if err := m.kubeClient.Get(ctx, client.ObjectKey{Name: cluster.Spec.ControlPlaneRef.Name, Namespace: namespace}, &controlPlane); err != nil {
		return nil, err
	}

	if annotations.HasPaused(&controlPlane) {
		return nil, fmt.Errorf("control plane %s/%s is paused. Unpause it before rolling out", controlPlane.GetName(), controlPlane.GetNamespace())
	}

	_, exists, err := unstructured.NestedFieldNoCopy(controlPlane.Object, "spec", "rolloutAfter")
	if err != nil {
		return nil, fmt.Errorf("failed to check rolloutAfter field: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("control plane %s/%s has rolloutAfter set. Remove it before rolling out", controlPlane.GetName(), controlPlane.GetNamespace())
	}
	patch := client.RawPatch(types.MergePatchType, []byte(fmt.Sprintf(`{"spec":{"rolloutAfter":"%v"}}`, time.Now().Format(time.RFC3339))))
	if err := m.kubeClient.Patch(ctx, &controlPlane, patch); err != nil {
		return nil, fmt.Errorf("failed to patch control plane: %w", err)
	}

	return mcp.NewToolResultText("Requested rollout of control plane successfully"), nil
}
