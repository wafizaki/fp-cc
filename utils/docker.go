package utils

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerClient struct {
	cli *client.Client
}

func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &DockerClient{cli: cli}, nil
}

func (dc *DockerClient) CreateAndStartContainer(ctx context.Context, tenantName string, port int, filesPath string, configPath string) (string, error) {
	containerName := fmt.Sprintf("files_%s", tenantName)
	volumeName := fmt.Sprintf("%s_settings_vol", tenantName)

	if err := dc.createVolume(ctx, volumeName); err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}

	if err := dc.ensureImage(ctx, "filebrowser/filebrowser"); err != nil {
		return "", fmt.Errorf("failed to ensure image: %w", err)
	}

	containerPort := nat.Port("80/tcp")
	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: fmt.Sprintf("%d", port),
	}

	config := &container.Config{
		Image: "filebrowser/filebrowser",
		ExposedPorts: nat.PortSet{
			containerPort: struct{}{},
		},
		User: "0:0",
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{hostBinding},
		},
		Binds: []string{
			fmt.Sprintf("%s:/srv:rw", filesPath),
			fmt.Sprintf("%s:/database:rw", volumeName),
			fmt.Sprintf("%s:/config:rw", configPath),
		},
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
	}

	resp, err := dc.cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := dc.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return containerName, nil
}

func (dc *DockerClient) createVolume(ctx context.Context, volumeName string) error {
	volumes, err := dc.cli.VolumeList(ctx, volume.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list volumes: %w", err)
	}

	for _, vol := range volumes.Volumes {
		if vol.Name == volumeName {
			return nil
		}
	}

	_, err = dc.cli.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	return nil
}

func (dc *DockerClient) ensureImage(ctx context.Context, imageName string) error {
	_, err := dc.cli.ImageInspect(ctx, imageName)
	if err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	reader, err := dc.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to read image pull output: %w", err)
	}

	return nil
}

func (dc *DockerClient) StartContainer(ctx context.Context, containerName string) error {
	if err := dc.cli.ContainerStart(ctx, containerName, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	return nil
}

func (dc *DockerClient) StopContainer(ctx context.Context, containerName string) error {
	timeout := 10
	stopOptions := container.StopOptions{
		Timeout: &timeout,
	}
	if err := dc.cli.ContainerStop(ctx, containerName, stopOptions); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	return nil
}

func (dc *DockerClient) RemoveContainer(ctx context.Context, containerName string) error {
	_ = dc.StopContainer(ctx, containerName)

	removeOptions := container.RemoveOptions{
		Force: true,
	}
	if err := dc.cli.ContainerRemove(ctx, containerName, removeOptions); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}

func (dc *DockerClient) RemoveVolume(ctx context.Context, volumeName string) error {
	if err := dc.cli.VolumeRemove(ctx, volumeName, true); err != nil {
		return fmt.Errorf("failed to remove volume: %w", err)
	}
	return nil
}

func (dc *DockerClient) GetContainerLogs(ctx context.Context, containerName string) (string, error) {
	time.Sleep(2 * time.Second)

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	logs, err := dc.cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer logs.Close()

	logBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("failed to read logs: %w", err)
	}

	logContent := string(logBytes)

	re := regexp.MustCompile(`generated password:\s+(\S+)`)
	matches := re.FindStringSubmatch(logContent)

	if len(matches) < 2 {
		return "", fmt.Errorf("password not found in logs")
	}

	password := strings.TrimSpace(matches[1])
	
	return password, nil
}

func (dc *DockerClient) InspectContainer(ctx context.Context, containerName string) (string, error) {
	containerJSON, err := dc.cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %w", err)
	}

	if containerJSON.State.Running {
		return "running", nil
	} else if containerJSON.State.Status == "exited" {
		return "stopped", nil
	}

	return containerJSON.State.Status, nil
}

func (dc *DockerClient) ContainerExists(ctx context.Context, containerName string) bool {
	_, err := dc.cli.ContainerInspect(ctx, containerName)
	return err == nil
}

func (dc *DockerClient) Close() error {
	if dc.cli != nil {
		return dc.cli.Close()
	}
	return nil
}