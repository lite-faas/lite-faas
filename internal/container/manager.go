package container

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	client containerd.Client
	logger *logrus.Logger
}

type ContainerConfig struct {
	Name     string
	Image    string
	Port     int
	Memory   string
	CPU      string
	Code     string
	Language string
}

func NewManager(socketPath string, logger *logrus.Logger) (*Manager, error) {
	client, err := containerd.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	return &Manager{
		client: *client,
		logger: logger,
	}, nil
}

func (m *Manager) CreateFunctionContainer(ctx context.Context, config *ContainerConfig) error {
	image, err := m.client.Pull(ctx, config.Image, containerd.WithPullUnpack)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	_, err = m.client.NewContainer(
		ctx,
		config.Name,
		containerd.WithImage(image),
		containerd.WithNewSpec(
			oci.WithImageConfig(image),
			oci.WithProcessArgs("python", "function.py"),
			oci.WithMounts([]specs.Mount{
				{
					Source:      "/tmp",
					Destination: "/app",
					Type:        "bind",
					Options:     []string{"rbind", "rw"},
				},
			}),
			oci.WithHostNamespace(specs.NetworkNamespace),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if err := m.writeFunctionCode(config); err != nil {
		return fmt.Errorf("failed to write function code: %w", err)
	}

	m.logger.Infof("Created function container: %s", config.Name)
	return nil
}

func (m *Manager) StartFunction(ctx context.Context, name string) error {
	container, err := m.client.LoadContainer(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	m.logger.Infof("Started function: %s", name)
	return nil
}

func (m *Manager) StopFunction(ctx context.Context, name string) error {
	container, err := m.client.LoadContainer(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if err := task.Kill(ctx, 9); err != nil {
		return fmt.Errorf("failed to kill task: %w", err)
	}

	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	m.logger.Infof("Stopped function: %s", name)
	return nil
}

func (m *Manager) DeleteFunction(ctx context.Context, name string) error {
	container, err := m.client.LoadContainer(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	m.logger.Infof("Deleted function: %s", name)
	return nil
}

func (m *Manager) GetFunctionStatus(ctx context.Context, name string) (string, error) {
	container, err := m.client.LoadContainer(ctx, name)
	if err != nil {
		return "", fmt.Errorf("failed to load container: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return "stopped", nil
	}

	status, err := task.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get task status: %w", err)
	}

	return string(status.Status), nil
}

func (m *Manager) writeFunctionCode(config *ContainerConfig) error {
	codeDir := filepath.Join("/tmp", config.Name)
	if err := os.MkdirAll(codeDir, 0755); err != nil {
		return fmt.Errorf("failed to create code directory: %w", err)
	}

	var filename string
	switch config.Language {
	case "python":
		filename = "function.py"
	case "nodejs":
		filename = "function.js"
	default:
		return fmt.Errorf("unsupported language: %s", config.Language)
	}

	filepath := filepath.Join(codeDir, filename)
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create function file: %w", err)
	}
	defer file.Close()

	if _, err := io.WriteString(file, config.Code); err != nil {
		return fmt.Errorf("failed to write function code: %w", err)
	}

	return nil
}
