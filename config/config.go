package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	ServerPort string
	DBPath     string
	BaseDir    string
}

func LoadConfig() *Config {
	baseDir, err := os.Getwd()
	if err != nil {
		baseDir = "."
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8081"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(baseDir, "filebrowser_db", "filebrowser.db")
	}

	return &Config{
		ServerPort: serverPort,
		DBPath:     dbPath,
		BaseDir:    baseDir,
	}
}

