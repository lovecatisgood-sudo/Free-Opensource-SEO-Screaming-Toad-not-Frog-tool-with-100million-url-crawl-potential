package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/api"
	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/config"
	"github.com/seo-auditor/seo-auditor/internal/webui"
)

func main() {
	paths, err := config.ResolvePaths()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := paths.Ensure(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	serverConfig, err := config.ResolveServer()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	service, err := application.Open(ctx, paths.Data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer service.Close()
	address := net.JoinHostPort(serverConfig.Host, fmt.Sprint(serverConfig.Port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	origin := "http://" + address
	handler, err := api.New(service, origin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", handler)
	mux.Handle("/", webui.Handler())
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	fmt.Fprintf(os.Stderr, "SEO Auditor listening on %s\n", origin)
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
