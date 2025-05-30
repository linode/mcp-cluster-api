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
