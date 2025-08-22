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
	"time"

	"github.com/go-logr/logr"
	"github.com/mark3labs/mcp-go/server"
	"k8s.io/client-go/rest"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

func NewToolManager(options ...func(*ToolManager)) *ToolManager {
	mgr := &ToolManager{}
	for _, o := range options {
		o(mgr)
	}
	return mgr
}

func WithKubeClient(kubeClient k8s.Client) func(*ToolManager) {
	return func(m *ToolManager) {
		m.kubeClient = kubeClient
	}
}

func WithConfig(cfg *rest.Config) func(*ToolManager) {
	return func(m *ToolManager) {
		m.cfg = cfg
	}
}

func WithTimeout(timeout int) func(*ToolManager) {
	return func(m *ToolManager) {
		m.timeout = time.Duration(timeout) * time.Second
	}
}

func WithReadOnly(readOnly bool) func(*ToolManager) {
	return func(m *ToolManager) {
		m.readOnly = readOnly
	}
}

func WithLogger(logger *logr.Logger) func(*ToolManager) {
	return func(m *ToolManager) {
		m.logger = logger
	}
}

func (m *ToolManager) RegisterTools(mcpServer *server.MCPServer) {
	tools := []ToolHandler{
		m.NewListTool(),
		m.NewGetClusterTool(),
		m.NewGetClusterKubeconfigTool(),
		m.NewCheckUpgradeEligibilityTool(),
		m.NewGetMachineTool(),
		m.NewGetKubeResourceTool(),
		m.NewRolloutControlPlaneTool(),
		m.NewGetControlPlaneStatusTool(),
	}

	for _, tool := range tools {
		if m.readOnly && !tool.ReadOnly {
			m.logger.Info("Skipping tool because it is not read-only", "tool", tool.Tool.Name)
			continue
		}
		mcpServer.AddTool(tool.Tool, tool.Handler)
	}
}
