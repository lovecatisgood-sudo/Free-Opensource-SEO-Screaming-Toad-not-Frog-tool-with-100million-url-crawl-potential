package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/seo-auditor/seo-auditor/internal/config"
	"github.com/seo-auditor/seo-auditor/internal/localclient"
	"github.com/seo-auditor/seo-auditor/internal/mcpserver"
)

func main() {
	serverConfig, err := config.ResolveServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client, err := localclient.New("http://" + net.JoinHostPort(serverConfig.Host, strconv.Itoa(serverConfig.Port)))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	server := mcpserver.New(client)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
