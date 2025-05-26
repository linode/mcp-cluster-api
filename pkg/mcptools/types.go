package mcptools

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

type ToolHandler struct {
	Tool     mcp.Tool
	Handler  server.ToolHandlerFunc
	ReadOnly bool
}

type ToolManager struct {
	kubeClient k8s.Client
	timeout    time.Duration
	readOnly   bool
	logger     *logr.Logger
}
