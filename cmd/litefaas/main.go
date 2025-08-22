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
	ContainerImagePrefix    string
	ContainerPrivileged     bool
	ContainerNetworkMode    string
	LogLevel                string
	LogFormat               string
	MaxMemoryPerFunction    string
	MaxCPUPerFunction       string
	DevMode                 bool
}

func loadConfig() *Config {
	config := &Config{}

	flag.IntVar(&config.Port, "port", 8080, "Port to listen on")
	flag.StringVar(&config.DBPath, "db-path", "/var/lib/litefaas/functions.db", "Database path")
	flag.StringVar(&config.ContainerdSocket, "containerd-socket", "/run/containerd/containerd.sock", "Containerd socket path")
	flag.IntVar(&config.FunctionPortStart, "function-port-start", 9000, "Start of function port range")
	flag.IntVar(&config.FunctionPortEnd, "function-port-end", 9999, "End of function port range")
	flag.StringVar(&config.ContainerImagePrefix, "container-image-prefix", "mirror.gcr.io/library/", "Container image prefix")
	flag.BoolVar(&config.ContainerPrivileged, "container-privileged", true, "Run containers in privileged mode")
	flag.StringVar(&config.ContainerNetworkMode, "container-network-mode", "host", "Container network mode")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level")
	flag.StringVar(&config.LogFormat, "log-format", "json", "Log format")
	flag.StringVar(&config.MaxMemoryPerFunction, "max-memory-per-function", "512MB", "Max memory per function")
	flag.StringVar(&config.MaxCPUPerFunction, "max-cpu-per-function", "0.5", "Max CPU per function")
	flag.BoolVar(&config.DevMode, "dev", false, "Development mode (disables containerd)")

	flag.Parse()

	if port := os.Getenv("LITEFAAS_PORT"); port != "" {
		if p, err := fmt.Sscanf(port, "%d", &config.Port); err == nil && p == 1 {
			// Port updated from environment
		}
	}
	if dbPath := os.Getenv("LITEFAAS_DB_PATH"); dbPath != "" {
		config.DBPath = dbPath
	}
	if socket := os.Getenv("LITEFAAS_CONTAINERD_SOCKET"); socket != "" {
		config.ContainerdSocket = socket
	}

	return config
}

func setupLogging(config *Config) {
	level, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	if config.LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
}

func main() {
	config := loadConfig()
	setupLogging(config)

	logrus.Info("Starting LiteFaaS...")

	db, err := database.New(config.DBPath)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	var containerManager *container.Manager
	if !config.DevMode {
		containerManager, err = container.NewManager(config.ContainerdSocket, config.ContainerImagePrefix, config.ContainerPrivileged, config.ContainerNetworkMode)
		if err != nil {
			logrus.Fatalf("Failed to initialize container manager: %v", err)
		}
	} else {
		logrus.Warn("Running in development mode - container management disabled")
	}

	proxy := proxy.New(db, containerManager, config.FunctionPortStart, config.FunctionPortEnd)

	apiHandler := api.NewHandler(db, containerManager, proxy)
	webHandler := web.NewHandler()

	router := http.NewServeMux()

	router.Handle("/api/", http.StripPrefix("/api", apiHandler))
	router.Handle("/", webHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: router,
	}

	go func() {
		logrus.Infof("Server starting on port %d", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logrus.Info("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("Server shutdown error: %v", err)
	}

	logrus.Info("LiteFaaS stopped")
}
