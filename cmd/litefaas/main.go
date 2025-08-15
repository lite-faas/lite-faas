package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/litefaas/litefaas/internal/api"
	"github.com/litefaas/litefaas/internal/container"
	"github.com/litefaas/litefaas/internal/database"
	"github.com/litefaas/litefaas/internal/proxy"
	"github.com/litefaas/litefaas/internal/web"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Port                    int
	DBPath                  string
	ContainerdSocket        string
	FunctionPortStart       int
	FunctionPortEnd         int
	LogLevel                string
	MaxMemoryPerFunction    string
	MaxCPUPerFunction       string
}

func main() {
	config := parseFlags()

	logger := setupLogger(config.LogLevel)

	db, err := database.New(config.DBPath)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	containerMgr, err := container.NewManager(config.ContainerdSocket, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize container manager: %v", err)
	}

	proxyHandler := proxy.NewHandler(containerMgr, db, logger)
	apiHandler := api.NewHandler(db, containerMgr, logger)
	webHandler := web.NewHandler(logger)

	router := setupRouter(proxyHandler, apiHandler, webHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}

	go func() {
		logger.Infof("Starting LiteFaaS server on port %d", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()

	waitForShutdown(server, logger)
}

func parseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.Port, "port", 8080, "Server port")
	flag.StringVar(&config.DBPath, "db-path", "/var/lib/litefaas/functions.db", "Database path")
	flag.StringVar(&config.ContainerdSocket, "containerd-socket", "/run/containerd/containerd.sock", "Containerd socket path")
	flag.IntVar(&config.FunctionPortStart, "function-port-start", 9000, "Start of function port range")
	flag.IntVar(&config.FunctionPortEnd, "function-port-end", 9999, "End of function port range")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level")
	flag.StringVar(&config.MaxMemoryPerFunction, "max-memory", "512MB", "Max memory per function")
	flag.StringVar(&config.MaxCPUPerFunction, "max-cpu", "0.5", "Max CPU per function")

	flag.Parse()

	return config
}

func setupLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logger.SetLevel(lvl)

	return logger
}

func setupRouter(proxyHandler *proxy.Handler, apiHandler *api.Handler, webHandler *web.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", proxyHandler.HandleRequest)
	mux.HandleFunc("/api/", apiHandler.HandleRequest)
	mux.HandleFunc("/web/", webHandler.HandleRequest)

	return mux
}

func waitForShutdown(server *http.Server, logger *logrus.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
