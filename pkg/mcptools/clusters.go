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
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/annotations"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

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

// NewGetClusterKubeconfigTool registers a tool to get a kubeconfig for accessing a CAPI cluster.
func (m *ToolManager) NewGetClusterKubeconfigTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("get_cluster_kubeconfig",
			mcp.WithDescription("Generate and retrieve a kubeconfig for accessing a CAPI cluster"),
			mcp.WithString("name", mcp.Required(), mcp.Description("The name of the cluster to get kubeconfig for")),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the cluster")),
		),
		Handler:  m.HandleGetClusterKubeconfig,
		ReadOnly: true,
	}
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

func (m *ToolManager) GetClusterKubeconfig(ctx context.Context, clusterName, namespace string) ([]byte, error) {
	// Get the Secret containing the kubeconfig
	secret := &corev1.Secret{}
	secretName := fmt.Sprintf("%s-kubeconfig", clusterName)
	if err := m.kubeClient.Get(ctx, client.ObjectKey{Name: secretName, Namespace: namespace}, secret); err != nil {
		return nil, fmt.Errorf("get kubeconfig secret: %w", err)
	}

	// Extract the kubeconfig from the secret
	kubeconfig, ok := secret.Data["value"]
	if !ok {
		return nil, fmt.Errorf("kubeconfig not found in secret %s/%s", namespace, secretName)
	}
	return kubeconfig, nil
}

func (m *ToolManager) BuildClientForWorkload(ctx context.Context, clusterName, namespace string) (*kubernetes.Clientset, error) {
	// Get the workload cluster's kubeconfig
	kubeconfig, err := m.GetClusterKubeconfig(ctx, clusterName, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster kubeconfig: %w", err)
	}

	// Create a Kubernetes client for the workload cluster
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST config from kubeconfig: %w", err)
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

// HandleGetClusterKubeconfig is the handler function for the get_cluster_kubeconfig tool.
func (m *ToolManager) HandleGetClusterKubeconfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		return nil, fmt.Errorf("get cluster: %w", err)
	}

	// Check if the cluster is provisioned
	if !cluster.Status.ControlPlaneReady {
		return nil, fmt.Errorf("cluster control plane is not ready yet, cannot retrieve kubeconfig")
	}

	kubeconfig, err := m.GetClusterKubeconfig(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("get cluster kubeconfig: %w", err)
	}

	return mcp.NewToolResultText(string(kubeconfig)), nil
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

	if cluster.Spec.ControlPlaneRef.Kind != "KubeadmControlPlane" && cluster.Spec.ControlPlaneRef.Kind != "KThreesControlPlane" {
		return nil, fmt.Errorf("cluster %s/%s control plane reference is not a KubeadmControlPlane or KThreesControlPlane. Rollout is unsupported", name, namespace)
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

// NewCheckUpgradeEligibilityTool registers a tool to check if a CAPI cluster can be safely upgraded.
func (m *ToolManager) NewCheckUpgradeEligibilityTool() ToolHandler {
	return ToolHandler{
		Tool: mcp.NewTool("check_upgrade_eligibility",
			mcp.WithDescription("Verify if a cluster can be safely upgraded"),
			mcp.WithString("name", mcp.Required(), mcp.Description("The name of the cluster to check for upgrade eligibility")),
			mcp.WithString("namespace", mcp.Required(), mcp.Description("The namespace of the cluster")),
			mcp.WithString("targetVersion", mcp.Required(), mcp.Description("The Kubernetes version to upgrade to (e.g., v1.27.1)")),
		),
		Handler:  m.HandleCheckUpgradeEligibility,
		ReadOnly: true,
	}
}

// HandleCheckUpgradeEligibility is the handler function for the check_upgrade_eligibility tool.
//
//nolint:cyclop // TODO: simplify later.
func (m *ToolManager) HandleCheckUpgradeEligibility(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	targetVersion, err := request.RequireString("targetVersion")
	if err != nil {
		return nil, err
	}

	// Get the cluster
	var cluster capi.Cluster
	if err := m.kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cluster); err != nil {
		return nil, fmt.Errorf("get cluster: %w", err)
	}

	// Check if the cluster is provisioned
	if !cluster.Status.ControlPlaneReady {
		return nil, fmt.Errorf("cluster control plane is not ready, cannot check upgrade eligibility")
	}

	// Get the control plane details to check current version
	var controlPlane unstructured.Unstructured
	if cluster.Spec.ControlPlaneRef == nil {
		return nil, fmt.Errorf("cluster %s/%s does not have a control plane reference", name, namespace)
	}

	controlPlane.SetGroupVersionKind(cluster.Spec.ControlPlaneRef.GroupVersionKind())

	// Get the control plane object
	if err := m.kubeClient.Get(ctx, client.ObjectKey{Name: cluster.Spec.ControlPlaneRef.Name, Namespace: namespace}, &controlPlane); err != nil {
		return nil, fmt.Errorf("failed to get control plane: %w", err)
	}

	// Check if control plane is paused
	if annotations.HasPaused(&controlPlane) {
		return nil, fmt.Errorf("control plane is paused, unpause it before checking upgrade eligibility")
	}

	// Get the current Kubernetes version from the control plane
	currentVersion, exists, err := unstructured.NestedString(controlPlane.Object, "spec", "version")
	if err != nil || !exists {
		return nil, fmt.Errorf("failed to get current Kubernetes version: %w", err)
	}

	// Format response with upgrade eligibility details
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Upgrade Eligibility Check for Cluster: %s/%s\n", namespace, name))
	sb.WriteString(fmt.Sprintf("Current Kubernetes Version: %s\n", currentVersion))
	sb.WriteString(fmt.Sprintf("Target Kubernetes Version: %s\n\n", targetVersion))

	// Check if current and target versions are different
	if currentVersion == targetVersion {
		sb.WriteString("The cluster is already running the target version.\n\n")
		return mcp.NewToolResultText(sb.String()), nil
	}

	// Check if cluster has ongoing operations
	_, exists, err = unstructured.NestedFieldNoCopy(controlPlane.Object, "status", "conditions")
	if err != nil {
		return nil, fmt.Errorf("failed to check control plane conditions: %w", err)
	}

	if exists {
		var hasOngoingOperations bool
		// Here we would normally check specific conditions, for this example we'll do a simple check
		// assuming the control plane has a 'Upgrading' or similar condition
		for _, condition := range cluster.Status.Conditions {
			if condition.Type == capi.ReadyCondition && condition.Status != corev1.ConditionTrue {
				hasOngoingOperations = true
				break
			}
		}

		if hasOngoingOperations {
			sb.WriteString("The cluster has ongoing operations. Wait for them to complete before upgrading.\n\n")
			return mcp.NewToolResultText(sb.String()), nil
		}
	}

	clientset, err := m.BuildClientForWorkload(ctx, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to build client for workload cluster: %w", err)
	}

	// Check node health status
	totalNodes, readyNodes, nodeDetails, err := checkNodeHealth(ctx, clientset)
	if err != nil {
		return nil, err
	}

	// Check if all nodes are ready
	sb.WriteString(fmt.Sprintf("Node Health: %d/%d Ready\n", readyNodes, totalNodes))
	for _, detail := range nodeDetails {
		sb.WriteString(detail + "\n")
	}
	sb.WriteString("\n")

	if readyNodes < totalNodes {
		sb.WriteString("Not all nodes are ready. Fix node issues before upgrading.\n\n")
		return mcp.NewToolResultText(sb.String()), nil
	}

	// Check for critical pod health in kube-system namespace
	criticalPods, unhealthyPods, err := checkCriticalPodHealth(ctx, clientset)
	if err != nil {
		return nil, err
	}

	sb.WriteString(fmt.Sprintf("Critical Control Plane Pods: %d\n", len(criticalPods)))
	if len(unhealthyPods) > 0 {
		sb.WriteString("Unhealthy critical pods:\n")
		for _, pod := range unhealthyPods {
			sb.WriteString(pod + "\n")
		}
		sb.WriteString("\n")
		sb.WriteString("Some critical pods are unhealthy. Resolve pod issues before upgrading.\n\n")
		return mcp.NewToolResultText(sb.String()), nil
	} else {
		sb.WriteString("All critical pods are healthy.\n\n")
	}

	// Final eligibility status
	sb.WriteString("Upgrade Eligibility Result:\n")
	sb.WriteString("The cluster is eligible for upgrade.\n")
	sb.WriteString("\n")
	sb.WriteString("Recommended steps:\n")
	sb.WriteString("1. Create a backup of the cluster's etcd data\n")
	sb.WriteString("2. Update the Kubernetes version in the control plane resource\n")
	sb.WriteString("3. Monitor the upgrade process with the 'get_control_plane_status' tool\n")

	return mcp.NewToolResultText(sb.String()), nil
}

// checkNodeHealth verifies the health status of all nodes in the cluster.
// It returns the total number of nodes, number of ready nodes, and details about each node.
func checkNodeHealth(ctx context.Context, clientset *kubernetes.Clientset) (totalNodes, readyNodes int, nodeDetails []string, err error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	totalNodes = len(nodes.Items)
	readyNodes = 0
	nodeDetails = make([]string, 0, totalNodes)

	for _, node := range nodes.Items {
		isReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				isReady = true
				readyNodes++
				break
			}
		}

		status := "Ready"
		if !isReady {
			status = "Not Ready"
		}

		nodeDetails = append(nodeDetails, fmt.Sprintf("  - Node: %s (%s, Version: %s)",
			node.Name,
			status,
			node.Status.NodeInfo.KubeletVersion))
	}

	return totalNodes, readyNodes, nodeDetails, nil
}

// checkCriticalPodHealth checks the health of critical pods in the kube-system namespace.
// It returns the list of critical pods and any unhealthy pods found.
func checkCriticalPodHealth(ctx context.Context, clientset *kubernetes.Clientset) (criticalPods, unhealthyPods []string, err error) {
	pods, err := clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list kube-system pods: %w", err)
	}

	criticalPods = make([]string, 0)
	unhealthyPods = make([]string, 0)

	for _, pod := range pods.Items {
		// Consider pods with these labels as critical
		if strings.Contains(pod.Name, "kube-apiserver") ||
			strings.Contains(pod.Name, "kube-controller") ||
			strings.Contains(pod.Name, "kube-scheduler") ||
			strings.Contains(pod.Name, "etcd") {
			criticalPods = append(criticalPods, pod.Name)

			if pod.Status.Phase != corev1.PodRunning {
				unhealthyPods = append(unhealthyPods, fmt.Sprintf("  - %s (%s)", pod.Name, pod.Status.Phase))
			}
		}
	}

	return criticalPods, unhealthyPods, nil
}
