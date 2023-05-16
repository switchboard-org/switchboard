package internal

import "os"

type OsManager interface {
	GetCurrentWorkingDirectory() string
	CreateDirectoryIfNotExists(path string) error
}

type DefaultOsManager struct{}

func NewDefaultOsManager() OsManager {
	return &DefaultOsManager{}
}
func (d *DefaultOsManager) GetCurrentWorkingDirectory() string {
	return CurrentWorkingDir()
}

func (d *DefaultOsManager) CreateDirectoryIfNotExists(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}
