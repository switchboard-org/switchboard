package internal

import (
	"os"
	"path/filepath"
)

// CurrentWorkingDir returns the value of the working directory that the CLI is executed in.
func CurrentWorkingDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}
