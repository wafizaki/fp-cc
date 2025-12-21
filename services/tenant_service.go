package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"tenant-manager/database"
	"tenant-manager/models"
	"tenant-manager/utils"
)

type TenantService struct {
	dockerClient *utils.DockerClient
	baseDir      string
}

func NewTenantService(dockerClient *utils.DockerClient, baseDir string) *TenantService {
	return &TenantService{
		dockerClient: dockerClient,
		baseDir:      baseDir,
	}
}

func (s *TenantService) CreateTenant(ctx context.Context, name string) (*models.Tenant, error) {
	_, err := database.GetTenantByName(name)
	if err == nil {
		return nil, fmt.Errorf("tenant already exists")
	}

	port, err := database.GetNextAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get next available port: %w", err)
	}

	tenantDir := filepath.Join(s.baseDir, "tenants", name)
	filesPath := filepath.Join(tenantDir, "files")
	configPath := filepath.Join(tenantDir, "config")

	if err := os.MkdirAll(filesPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create files directory: %w", err)
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		os.RemoveAll(tenantDir)
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	containerName := fmt.Sprintf("files_%s", name)
	volumeName := fmt.Sprintf("%s_settings_vol", name)

	_, err = s.dockerClient.CreateAndStartContainer(ctx, name, port, filesPath, configPath)
	if err != nil {
		os.RemoveAll(tenantDir)
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	password, err := s.dockerClient.GetContainerLogs(ctx, containerName)
	if err != nil {
		password = ""
	}

	dbTenant := &database.Tenant{
		Name:          name,
		Port:          port,
		ContainerName: containerName,
		VolumeName:    volumeName,
		Status:        models.StatusRunning,
	}

	if err := database.CreateTenant(dbTenant); err != nil {
		s.dockerClient.RemoveContainer(ctx, containerName)
		s.dockerClient.RemoveVolume(ctx, volumeName)
		os.RemoveAll(tenantDir)
		return nil, fmt.Errorf("failed to save tenant to database: %w", err)
	}

	tenant := &models.Tenant{
		ID:            int(dbTenant.ID),
		Name:          name,
		Port:          port,
		ContainerName: containerName,
		VolumeName:    volumeName,
		Status:        models.StatusRunning,
		URL:           fmt.Sprintf("http://localhost:%d", port),
		Username:      "admin",
		Password:      password,
		CreatedAt:     dbTenant.CreatedAt,
		UpdatedAt:     dbTenant.UpdatedAt,
	}

	return tenant, nil
}

func (s *TenantService) ListTenants(ctx context.Context, page, perPage int) ([]models.Tenant, models.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	dbTenants, total, err := database.GetAllTenants(page, perPage)
	if err != nil {
		return nil, models.PaginationMeta{}, fmt.Errorf("failed to get tenants: %w", err)
	}

	tenants := make([]models.Tenant, 0, len(dbTenants))
	for _, dbTenant := range dbTenants {
		status := dbTenant.Status
		if s.dockerClient.ContainerExists(ctx, dbTenant.ContainerName) {
			containerStatus, err := s.dockerClient.InspectContainer(ctx, dbTenant.ContainerName)
			if err == nil {
				status = containerStatus
				if status != dbTenant.Status {
					database.UpdateTenantStatus(dbTenant.Name, status)
				}
			}
		}

		tenant := models.Tenant{
			ID:            int(dbTenant.ID),
			Name:          dbTenant.Name,
			Port:          dbTenant.Port,
			ContainerName: dbTenant.ContainerName,
			VolumeName:    dbTenant.VolumeName,
			Status:        status,
			URL:           fmt.Sprintf("http://localhost:%d", dbTenant.Port),
			Username:      "admin",
			CreatedAt:     dbTenant.CreatedAt,
			UpdatedAt:     dbTenant.UpdatedAt,
		}
		tenants = append(tenants, tenant)
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	meta := models.PaginationMeta{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		MaxPage: totalPages,
	}

	return tenants, meta, nil
}

func (s *TenantService) GetTenant(ctx context.Context, name string) (*models.Tenant, error) {
	dbTenant, err := database.GetTenantByName(name)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	status := dbTenant.Status
	if s.dockerClient.ContainerExists(ctx, dbTenant.ContainerName) {
		containerStatus, err := s.dockerClient.InspectContainer(ctx, dbTenant.ContainerName)
		if err == nil {
			status = containerStatus
			if status != dbTenant.Status {
				database.UpdateTenantStatus(dbTenant.Name, status)
			}
		}
	}

	password := ""
	if status == models.StatusRunning {
		password, _ = s.dockerClient.GetContainerLogs(ctx, dbTenant.ContainerName)
	}

	tenant := &models.Tenant{
		ID:            int(dbTenant.ID),
		Name:          dbTenant.Name,
		Port:          dbTenant.Port,
		ContainerName: dbTenant.ContainerName,
		VolumeName:    dbTenant.VolumeName,
		Status:        status,
		URL:           fmt.Sprintf("http://localhost:%d", dbTenant.Port),
		Username:      "admin",
		Password:      password,
		CreatedAt:     dbTenant.CreatedAt,
		UpdatedAt:     dbTenant.UpdatedAt,
	}

	return tenant, nil
}

func (s *TenantService) StopTenantContainer(ctx context.Context, name string) error {
	dbTenant, err := database.GetTenantByName(name)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if dbTenant.Status != models.StatusRunning {
		return fmt.Errorf("container is not running")
	}

	if err := s.dockerClient.StopContainer(ctx, dbTenant.ContainerName); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	if err := database.UpdateTenantStatus(name, models.StatusStopped); err != nil {
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	return nil
}

func (s *TenantService) StartTenantContainer(ctx context.Context, name string) error {
	dbTenant, err := database.GetTenantByName(name)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if dbTenant.Status == models.StatusRunning {
		return fmt.Errorf("container is already running")
	}

	if err := s.dockerClient.StartContainer(ctx, dbTenant.ContainerName); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	if err := database.UpdateTenantStatus(name, models.StatusRunning); err != nil {
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	return nil
}

func (s *TenantService) DeleteTenant(ctx context.Context, name string) error {
	dbTenant, err := database.GetTenantByName(name)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if err := s.dockerClient.RemoveContainer(ctx, dbTenant.ContainerName); err != nil {
		fmt.Printf("Warning: failed to remove container: %v\n", err)
	}

	if err := s.dockerClient.RemoveVolume(ctx, dbTenant.VolumeName); err != nil {
		fmt.Printf("Warning: failed to remove volume: %v\n", err)
	}

	tenantDir := filepath.Join(s.baseDir, "tenants", name)
	if err := os.RemoveAll(tenantDir); err != nil {
		fmt.Printf("Warning: failed to remove tenant directory: %v\n", err)
	}

	if err := database.DeleteTenant(name); err != nil {
		return fmt.Errorf("failed to delete tenant from database: %w", err)
	}

	return nil
}

