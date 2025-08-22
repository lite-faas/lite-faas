package container

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type Manager struct {
	client            *containerd.Client
	imagePrefix       string
	privileged        bool
	networkMode       string
	namespace         string
	containerTemplate *Template
}

type Template struct {
	Python *PythonTemplate
	NodeJS *NodeJSTemplate
}

type PythonTemplate struct {
	BaseImage string
	Port      int
}

type NodeJSTemplate struct {
	BaseImage string
	Port      int
}

func NewManager(socketPath, imagePrefix string, privileged bool, networkMode string) (*Manager, error) {
	client, err := containerd.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	namespace := "litefaas"

	manager := &Manager{
		client:      client,
		imagePrefix: imagePrefix,
		privileged:  privileged,
		networkMode: networkMode,
		namespace:   namespace,
		containerTemplate: &Template{
			Python: &PythonTemplate{
				BaseImage: imagePrefix + "python:3.11-alpine",
				Port:      8080,
			},
			NodeJS: &NodeJSTemplate{
				BaseImage: imagePrefix + "node:18-alpine",
				Port:      8080,
			},
		},
	}

	return manager, nil
}

func (m *Manager) CreateFunctionContainer(ctx context.Context, functionName, language, code string, port int) (string, error) {
	ctx = namespaces.WithNamespace(ctx, m.namespace)

	var imageName string

	switch language {
	case "python":
		imageName = m.containerTemplate.Python.BaseImage
	case "nodejs":
		imageName = m.containerTemplate.NodeJS.BaseImage
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	image, err := m.client.Pull(ctx, imageName, containerd.WithPullUnpack)
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}

	containerID := fmt.Sprintf("litefaas-%s-%d", functionName, time.Now().Unix())
	containerName := fmt.Sprintf("%s-%s", m.namespace, containerID)

	tempDir, err := os.MkdirTemp("", "litefaas-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := m.createFunctionFiles(tempDir, language, code); err != nil {
		return "", fmt.Errorf("failed to create function files: %w", err)
	}

	var opts []oci.SpecOpts
	if m.privileged {
		opts = append(opts, oci.WithPrivileged)
	}

	opts = append(opts,
		oci.WithImageConfig(image),
		oci.WithMounts([]specs.Mount{
			{
				Source:      tempDir,
				Destination: "/app",
				Type:        "bind",
				Options:     []string{"rbind", "ro"},
			},
		}),
		oci.WithEnv([]string{
			fmt.Sprintf("PORT=%d", port),
			"HOST=127.0.0.1",
		}),
	)

	container, err := m.client.NewContainer(
		ctx,
		containerName,
		containerd.WithImage(image),
		containerd.WithNewSpec(opts...),
		containerd.WithContainerLabels(map[string]string{
			"litefaas.function": functionName,
			"litefaas.language": language,
			"litefaas.port":     strconv.Itoa(port),
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	if err != nil {
		container.Delete(ctx)
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		container.Delete(ctx)
		return "", fmt.Errorf("failed to start task: %w", err)
	}

	return containerID, nil
}

func (m *Manager) createFunctionFiles(tempDir, language, code string) error {
	switch language {
	case "python":
		return m.createPythonFiles(tempDir, code)
	case "nodejs":
		return m.createNodeJSFiles(tempDir, code)
	default:
		return fmt.Errorf("unsupported language: %s", language)
	}
}

func (m *Manager) createPythonFiles(tempDir, code string) error {
	functionCode := fmt.Sprintf(`import os
import json
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse, parse_qs

class FunctionHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()

        query = urlparse(self.path).query
        params = parse_qs(query)

        response = {
            'method': 'GET',
            'path': self.path,
            'params': params,
            'result': 'Hello from Python function!'
        }

        self.wfile.write(json.dumps(response).encode())

    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)

        try:
            body = json.loads(post_data.decode('utf-8'))
        except:
            body = post_data.decode('utf-8')

        self.send_response(200)
        self.send_header('Content-type', 'application/json')
        self.end_headers()

        response = {
            'method': 'POST',
            'path': self.path,
            'body': body,
            'result': 'Hello from Python function!'
        }

        self.wfile.write(json.dumps(response).encode())

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    server = HTTPServer(('127.0.0.1', port), FunctionHandler)
    print(f'Python function server starting on port {port}')
    server.serve_forever()
`)

	if err := os.WriteFile(filepath.Join(tempDir, "function.py"), []byte(functionCode), 0644); err != nil {
		return err
	}

	requirements := `requests==2.31.0`
	return os.WriteFile(filepath.Join(tempDir, "requirements.txt"), []byte(requirements), 0644)
}

func (m *Manager) createNodeJSFiles(tempDir, code string) error {
	functionCode := fmt.Sprintf(`const http = require('http');
const url = require('url');

const server = http.createServer((req, res) => {
    const parsedUrl = url.parse(req.url, true);

    let body = '';
    req.on('data', chunk => {
        body += chunk.toString();
    });

    req.on('end', () => {
        let parsedBody = body;
        try {
            parsedBody = JSON.parse(body);
        } catch (e) {
            // Keep as string if not JSON
        }

        const response = {
            method: req.method,
            path: req.url,
            params: parsedUrl.query,
            body: parsedBody,
            result: 'Hello from Node.js function!'
        };

        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify(response));
    });
});

const port = process.env.PORT || 8080;
const host = process.env.HOST || '127.0.0.1';

server.listen(port, host, () => {
    console.log('Node.js function server starting on ' + host + ':' + port);
});
`)

	if err := os.WriteFile(filepath.Join(tempDir, "function.js"), []byte(functionCode), 0644); err != nil {
		return err
	}

	packageJSON := `{
  "name": "litefaas-function",
  "version": "1.0.0",
  "main": "function.js",
  "dependencies": {}
}`
	return os.WriteFile(filepath.Join(tempDir, "package.json"), []byte(packageJSON), 0644)
}

func (m *Manager) StopContainer(ctx context.Context, containerID string) error {
	ctx = namespaces.WithNamespace(ctx, m.namespace)
	containerName := fmt.Sprintf("%s-%s", m.namespace, containerID)

	container, err := m.client.LoadContainer(ctx, containerName)
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

	if err := container.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	return nil
}

func (m *Manager) ListContainers(ctx context.Context) ([]containerd.Container, error) {
	ctx = namespaces.WithNamespace(ctx, m.namespace)
	return m.client.Containers(ctx, "labels.litefaas.function")
}

func (m *Manager) GetContainerStatus(ctx context.Context, containerID string) (string, error) {
	ctx = namespaces.WithNamespace(ctx, m.namespace)
	containerName := fmt.Sprintf("%s-%s", m.namespace, containerID)

	container, err := m.client.LoadContainer(ctx, containerName)
	if err != nil {
		return "not_found", nil
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return "stopped", nil
	}

	status, err := task.Status(ctx)
	if err != nil {
		return "unknown", err
	}

	return string(status.Status), nil
}

func (m *Manager) Close() error {
	return m.client.Close()
}
