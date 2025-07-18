// Copyright (c) 2025 Minand Nellipunath Manomohanan
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package config

// ConfigManager defines the interface for configuration management operations
type ConfigManager interface {
	Load() (*Config, error)
	Save(cfg *Config) error
	Validate() error
}

// FileSystemOperations defines the interface for file system operations used by config
type FileSystemOperations interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm uint32) error
	MkdirAll(path string, perm uint32) error
	Stat(name string) (FileInfo, error)
	UserHomeDir() (string, error)
}

// FileInfo defines the interface for file information
type FileInfo interface {
	IsDir() bool
	ModTime() interface{}
	Name() string
	Size() int64
}