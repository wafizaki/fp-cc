package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type Tenant struct {
	ID            uint      `gorm:"primaryKey"`
	Name          string    `gorm:"uniqueIndex;not null"`
	Port          int       `gorm:"uniqueIndex;not null"`
	ContainerName string    `gorm:"not null"`
	VolumeName    string    `gorm:"not null"`
	Status        string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

func InitDB(dbPath string) error {
	var err error
	
	gormLogger := logger.Default.LogMode(logger.Silent)
	
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := DB.AutoMigrate(&Tenant{}); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database initialized successfully with GORM")
	return nil
}

func CreateTenant(tenant *Tenant) error {
	result := DB.Create(tenant)
	if result.Error != nil {
		return fmt.Errorf("failed to insert tenant: %w", result.Error)
	}
	return nil
}

func GetTenantByName(name string) (*Tenant, error) {
	var tenant Tenant
	result := DB.Where("name = ?", name).First(&tenant)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("tenant not found")
		}
		return nil, fmt.Errorf("failed to query tenant: %w", result.Error)
	}

	return &tenant, nil
}

func GetAllTenants(page, perPage int) ([]Tenant, int, error) {
	var tenants []Tenant
	var total int64

	if err := DB.Model(&Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	offset := (page - 1) * perPage

	result := DB.Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&tenants)

	if result.Error != nil {
		return nil, 0, fmt.Errorf("failed to query tenants: %w", result.Error)
	}

	return tenants, int(total), nil
}

func UpdateTenantStatus(name, status string) error {
	result := DB.Model(&Tenant{}).
		Where("name = ?", name).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update tenant status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

func DeleteTenant(name string) error {
	result := DB.Where("name = ?", name).Delete(&Tenant{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete tenant: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

func GetNextAvailablePort() (int, error) {
	var tenant Tenant
	result := DB.Order("port DESC").Limit(1).Find(&tenant)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to get last port: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return 9000, nil
	}

	return tenant.Port + 1, nil
}

func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}
