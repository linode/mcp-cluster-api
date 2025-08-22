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
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/linode/capi-mcp/pkg/mcptools"
	"github.com/linode/capi-mcp/pkg/prompts"
)

var (
	defaultPort    = 8080
	defaultTimeout = 10
	version        = "v0.0.0-alpha0"
	scheme         *runtime.Scheme
)

type CliOptions struct {
	ReadOnly  bool
	Transport string
	Port      int
	Timeout   int
}

var cliOptions CliOptions

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "capi-mcp",
	Short: "A MCP server for Cluster API",
	Long: `Cluster API Model Context Protocol (MCP) server. This server can be used in conjunction with a LLM client to 
provide a conversational interface to the Cluster API. 
The server can be used to manage clusters, machines, and other resources in a Kubernetes cluster.`,
	RunE: cmdRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&cliOptions.ReadOnly, "read-only", false,
		"Run the server in read-only mode, disabling write operations.")
	rootCmd.PersistentFlags().StringVar(&cliOptions.Transport, "transport", "stdio",
		"The transport protocol to use for the MCP server. Options: [stdio, sse].")
	rootCmd.PersistentFlags().IntVar(&cliOptions.Port, "port", defaultPort,
		"The port to use for the MCP server. Used only when the transport is set to 'sse'.")
	rootCmd.PersistentFlags().IntVar(&cliOptions.Timeout, "timeout", defaultTimeout, "The timeout for the MCP server.")
	rootCmd.SetOut(os.Stdout)

	scheme = runtime.NewScheme()
	if err := capi.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add CAPI scheme: %v\n", err)
		os.Exit(1)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add corev1 scheme: %v\n", err)
		os.Exit(1)
	}
}

func cmdRun(cmd *cobra.Command, args []string) error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.WriteTo(os.Stdout)))
	log := ctrl.Log.WithName("capi-mcp")

	if os.Getenv("KUBECONFIG") == "" {
		return errors.New("KUBECONFIG environment variable is not set")
	}

	kubeConfig := config.GetConfigOrDie()

	// initialize the client using kubeconfig
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	// create our manager
	toolManager := mcptools.NewToolManager(
		mcptools.WithConfig(kubeConfig),
		mcptools.WithKubeClient(kubeClient),
		mcptools.WithTimeout(cliOptions.Timeout),
		mcptools.WithReadOnly(cliOptions.ReadOnly),
		mcptools.WithLogger(&log),
	)

	promptManager := prompts.NewPromptManager()
	mcpServer := server.NewMCPServer(
		"capi-mcp",
		version,
		server.WithToolCapabilities(true),
		server.WithPromptCapabilities(true),
	)

	toolManager.RegisterTools(mcpServer)
	promptManager.RegisterPrompts(mcpServer)

	if cliOptions.Transport == "sse" {
		sseServer := server.NewSSEServer(mcpServer, server.WithKeepAlive(true))
		if err := sseServer.Start(fmt.Sprintf(":%d", cliOptions.Port)); err != nil {
			return err
		}
	} else {
		if err := server.ServeStdio(mcpServer); err != nil {
			return err
		}
	}
	return nil
}
