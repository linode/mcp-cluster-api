# capi-mcp Copilot Instructions

## Project Architecture

This is a **Model Context Protocol (MCP) server** that provides an AI-conversational interface to **Cluster API (CAPI)** resources. The architecture follows a clear separation:

- `cmd/root.go`: Cobra CLI entry point, handles server initialization and transport modes (stdio/sse)
- `pkg/mcptools/`: Tool implementations for CAPI resource operations 
- `pkg/prompts/`: Structured troubleshooting workflows as MCP prompts
- `main.go`: Simple entry point delegating to cmd package

## Key Dependencies & Integration Points

- **mcp-go library**: `github.com/mark3labs/mcp-go` for MCP protocol implementation
- **CAPI**: `sigs.k8s.io/cluster-api` for Cluster API resource definitions
- **controller-runtime**: `sigs.k8s.io/controller-runtime/pkg/client` for Kubernetes client operations
- **KUBECONFIG environment variable**: Required for all operations, no default fallback

## Core Patterns

### Tool Definition Pattern
```go
func (m *ToolManager) NewExampleTool() ToolHandler {
    return ToolHandler{
        Tool: mcp.NewTool("tool_name",
            mcp.WithDescription("Clear description of what tool does"),
            mcp.WithString("param", mcp.Required(), mcp.Description("Parameter description")),
        ),
        Handler:  m.HandleExample,
        ReadOnly: true, // Set false for write operations
    }
}
```

### Handler Implementation Pattern
```go
func (m *ToolManager) HandleExample(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    ctx, cancel := context.WithTimeout(ctx, m.timeout)
    defer cancel()
    
    // Always validate required parameters first
    name, err := request.RequireString("name")
    if err != nil {
        return nil, err
    }
    
    // Clean up managed fields before returning YAML
    resource.SetManagedFields(nil)
    resourceBytes, err := yaml.Marshal(resource.DeepCopyObject())
    
    return mcp.NewToolResultText(string(resourceBytes)), nil
}
```

### Prompt Definition Pattern
```go
func (pm *PromptManager) NewExamplePrompt() PromptHandler {
    return PromptHandler{
        Prompt: mcp.NewPrompt("prompt_name",
            mcp.WithArgument("name", mcp.ArgumentDescription("Description"), mcp.RequiredArgument()),
        ),
        Handler: pm.NewExamplePromptHandler,
    }
}
```

## Critical Development Workflows

### Build & Development
```bash
make build          # Builds to ./bin/capi-mcp
make lint           # Auto-installs golangci-lint if missing, runs with --fix
```

### Testing with MCP Inspector
```bash
# Essential for testing tool/prompt functionality
npx @modelcontextprotocol/inspector -e KUBECONFIG=kubeconfig.yaml bin/capi-mcp
```

### Runtime Modes
- **stdio mode** (default): For integration with AI clients like Claude Desktop
- **sse mode**: `--transport=sse --port=8080` for web-based debugging

## Project-Specific Conventions

### Resource Handling
- **Always call** `resource.SetManagedFields(nil)` before marshaling to YAML - prevents output clutter
- **Always use** `resource.DeepCopyObject()` when marshaling to avoid modifying original objects
- **Always implement** timeout contexts: `ctx, cancel := context.WithTimeout(ctx, m.timeout)`

### Tool Categories
- **Read-only tools**: List/get operations, status checks (ReadOnly: true)
- **Write operations**: Only `rollout_controlplane` currently (ReadOnly: false)
- **ReadOnly mode**: CLI flag `--read-only` skips non-read-only tools during registration

### Error Handling Patterns
- Use `request.RequireString()` for required parameters - returns proper error format
- Use `request.GetString("param", "default")` for optional parameters  
- Timeout all Kubernetes operations with manager's configured timeout
- Return `fmt.Errorf("operation: %w", err)` for context in error chains

### Naming Conventions
- Tool names: `snake_case` (e.g., `get_cluster`, `list_machines`)
- Prompt names: `snake_case` (e.g., `debug_capi_cluster`)
- Package methods: `NewXxxTool()` and `HandleXxx()` pattern
- File organization: Group related tools in same file (`clusters.go`, `machines.go`)

## Integration Points

### Kubernetes Resource Access
- Uses controller-runtime client with custom scheme including CAPI and core v1 resources
- Requires RBAC permissions for CAPI resources (Cluster, Machine, etc.)
- Supports cross-namespace operations via namespace parameters

### MCP Protocol Implementation
- Tools return `mcp.NewToolResultText()` with YAML-formatted Kubernetes resources
- Prompts return structured debugging workflows via `mcp.NewGetPromptResult()`
- Server supports both tool and prompt capabilities simultaneously

This codebase prioritizes **clean YAML output** and **structured troubleshooting workflows** for AI-driven CAPI management.
