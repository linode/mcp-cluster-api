# capi-mcp

**capi-mcp** is a tool that integrates [CAPI](https://cluster-api.sigs.k8s.io/) (Cluster API) resources with the [Model Context Protocol (MCP)](https://mcp.so/), enabling programmatic and prompt-based management of Kubernetes clusters and machines.

### What is MCP?
**MCP (Model Context Protocol)** is an open protocol for describing, invoking, and managing tools and prompts in a standardized way. It enables interoperability between tools, agents, and UIs.

### What is CAPI?
**CAPI (Cluster API)** is a Kubernetes project to manage Kubernetes clusters declaratively using Kubernetes-style APIs. It provides a consistent way to create, update, and manage clusters and their infrastructure.

## Prerequisites

- [Go](https://golang.org/doc/install) (v1.20+ recommended)
- [Make](https://www.gnu.org/software/make/)
- [npm](https://nodejs.org/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- Access to a Kubernetes cluster and a valid `kubeconfig.yaml`

## Installation & Build

Clone the repository and build the binary:

```sh
make build
```

## Linting

Run linting (installs golangci-lint if not present):

```sh
make lint
```

## Running & Testing

**Transport support**: `capi-mcp` supports only stdio transport. 

To run the MCP inspector with your built binary:

```sh
npx @modelcontextprotocol/inspector -e KUBECONFIG=kubeconfig.yaml bin/capi-mcp
```

Replace `kubeconfig.yaml` with your actual kubeconfig file.

## Usage with Claude and VSCode

### Using with Claude Desktop

To use this MCP server with Claude Desktop, you need to configure it in Claude's settings:

1. **Build the binary** (if not already done):
   ```sh
   make build
   ```

2. **Locate Claude's configuration file**:
   - On macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - On Windows: `%APPDATA%\Claude\claude_desktop_config.json`
   - On Linux: `~/.config/Claude/claude_desktop_config.json`

3. **Add the MCP server configuration**:
   ```json
   {
     "mcpServers": {
       "capi-mcp": {
         "command": "/absolute/path/to/capi-mcp/bin/capi-mcp",
         "env": {
           "KUBECONFIG": "/absolute/path/to/your/kubeconfig.yaml"
         }
       }
     }
   }
   ```

4. **Restart Claude Desktop** to load the new configuration.

5. **Verify the connection**: In a new conversation with Claude, you should see that the CAPI MCP server is connected and running through developer settings.
<img width="2578" height="1226" alt="image" src="https://github.com/user-attachments/assets/953dc171-0d08-40aa-ade3-67b67fc5aad4" />


### Using with VSCode

#### Via Claude Extension

If you're using Claude through a VSCode extension that supports MCP:

1. **Install the Claude extension** from the VSCode marketplace.

2. **Configure the MCP server** in your VSCode settings or the extension's configuration file:
   ```json
   {
     "claude.mcpServers": {
       "capi-mcp": {
         "command": "/absolute/path/to/capi-mcp/bin/capi-mcp",
         "env": {
           "KUBECONFIG": "/absolute/path/to/your/kubeconfig.yaml"
         }
       }
     }
   }
   ```

3. **Restart VSCode** to apply the configuration.

#### Via VSCode Toolsets

You can also integrate this MCP server with VSCode toolsets for enhanced development workflows:

1. **Create a toolset configuration** in your workspace:
   - Open the Command Palette (`Cmd/Ctrl + Shift + P`)
   - Run "Toolsets: Configure Toolsets" 
   - Add a new toolset configuration

2. **Configure the CAPI MCP toolset**:
   ```json
   {
     "name": "CAPI MCP Server",
     "description": "Cluster API management via MCP",
     "tools": [
       {
         "name": "capi-mcp",
         "command": "/absolute/path/to/capi-mcp/bin/capi-mcp",
         "env": {
           "KUBECONFIG": "/absolute/path/to/your/kubeconfig.yaml"
         },
         "protocol": "mcp"
       }
     ]
   }
   ```

3. **Use the toolset**:
   - Access via the Toolsets panel in the sidebar
   - Invoke tools directly from the Command Palette
   - Integrate with other VSCode AI features and extensions

This allows you to use CAPI management tools directly within your VSCode development environment alongside other development toolsets.

### Available Tools and Prompts

Once connected, you can use Claude to:

**Tools:**
- `list_clusters` - List all CAPI clusters in your Kubernetes environment
- `get_cluster` - Get detailed information about a specific cluster
- `get_cluster_kubeconfig` - Generate and retrieve a kubeconfig for accessing a CAPI cluster
- `check_upgrade_eligibility` - Verify if a cluster can be safely upgraded to a target Kubernetes version
- `list_machines` - List all machines (nodes) across clusters
- `get_machine` - Get detailed information about a specific machine
- `get_kube_resource` - Get any Kubernetes resource by name, kind, and API version
- `rollout_controlplane` - Rollout a restart of control plane (triggers rolling update - **non-read-only**)
- `get_control_plane_status` - Get status of control plane components for a CAPI cluster

**Prompts:**
- `debug_capi_cluster` - Get step-by-step debugging guidance for cluster issues
- `debug_capi_machine` - Get step-by-step debugging guidance for machine issues
- `restart_control_plane` - Get step-by-step instructions for restarting a cluster's control plane

### Example Usage

Ask Claude / Copilot questions like:
- "Show me all my CAPI clusters"
- "What's the status of my cluster named 'production'?"
- "List all machines in the cluster 'staging'"
- "Get me the kubeconfig for my cluster 'dev-cluster' in namespace 'default'"
- "Check if my cluster can be upgraded to Kubernetes v1.32.0"
- "Help me debug my cluster using the debug_capi_cluster prompt"
- "Get the status of control plane components for my cluster"
- "Get the deployment named 'my-app' in the default namespace using kind 'Deployment' and apiVersion 'apps/v1'"

**Note**: Make sure your `kubeconfig.yaml` file has the necessary permissions to access CAPI resources in your Kubernetes cluster. Some tools like `rollout_controlplane` are write operations and require appropriate RBAC permissions.
