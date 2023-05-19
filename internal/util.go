package internal

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// CurrentWorkingDir returns the value of the working directory that the CLI is executed in.
func CurrentWorkingDir() string {
	ex, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Clean(ex)
}

// Ptr is a Helper function to return a pointer value to any type
func Ptr[T any](t T) *T {
	return &t
}

func CapitalizeFirstLetterOfWord(word string) string {
	if len(word) == 0 {
		return word
	}
	return string(unicode.ToUpper(rune(word[0]))) + word[1:]
}

func PackageName(source string) string {
	packageNameParts := strings.Split(source, "/")
	return packageNameParts[len(packageNameParts)-1]
}

func MergeMaps[T any](mapOne map[string]T, mapTwo map[string]T) map[string]T {
	result := map[string]T{}
	for key, value := range mapOne {
		result[key] = value
	}
	for key, value := range mapTwo {
		result[key] = value
	}
	return result
}
