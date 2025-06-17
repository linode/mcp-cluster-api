how to test
## capi-mcp

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

To run the MCP inspector with your built binary:

```sh
npx @modelcontextprotocol/inspector -e KUBECONFIG=kubeconfig.yaml bin/capi-mcp
```

Replace `kubeconfig.yaml` with your actual kubeconfig file if different.


TODO:
## Cluster Management Tools
1. **upgrade_cluster** - Upgrade a cluster to a new Kubernetes version

## Machine Management Tools
1. **drain_machine** - Safely drain workloads from a machine before maintenance
2. **cordon_machine/uncordon_machine** - Mark machines as unschedulable/schedulable
3. **delete_machine** - Remove a specific machine from the cluster
4. **get_machine_logs** - Fetch logs from a specific machine for troubleshooting

## Cluster Health and Diagnostics
1. **get_cluster_health** - Check overall health status of a CAPI cluster
2. **get_cluster_events** - Retrieve recent events related to a cluster
3. **troubleshoot_cluster** - Run a diagnostic suite on a problematic cluster
4. **describe_node_pool** - Get detailed information about a cluster's node pools

## Infrastructure Provider Integration
1. **list_machine_templates** - Show available machine templates for the cluster
2. **get_machine_template** - Get details of a specific machine template
3. **list_provider_regions** - Show available regions for the infrastructure provider
4. **list_instance_types** - List available VM instance types for your provider

## Cluster Upgrade and Maintenance
1. **upgrade_cluster** - Upgrade a cluster

