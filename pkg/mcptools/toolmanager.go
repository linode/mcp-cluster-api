package mcptools

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/mark3labs/mcp-go/server"
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
		m.NewListClustersTool(),
		m.NewGetClusterTool(),
		m.NewListMachinesTool(),
		m.NewGetMachineTool(),
		m.NewGetKubeResourceTool(),
		m.NewRolloutControlPlaneTool(),
	}

	for _, tool := range tools {
		if m.readOnly && !tool.ReadOnly {
			m.logger.Info("Skipping tool because it is not read-only", "tool", tool.Tool.Name)
			continue
		}
		mcpServer.AddTool(tool.Tool, tool.Handler)
	}
}
