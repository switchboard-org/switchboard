package internal

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/json"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func findAllFiles(root, ext string) []string {
	var allFilePaths []string
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if filepath.Ext(d.Name()) == ext {
			allFilePaths = append(allFilePaths, s)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error finding files. %s", err)
	}
	return allFilePaths
}

// loadAllHclFilesInDir finds all '.hcl' files in the working
// directory and any child directories (including deeply nested dirs) and transforms it into a parsed hcl.Body.
func loadAllHclFilesInDir(path string) (hcl.Body, hcl.Diagnostics) {
	parser := hclparse.NewParser()
	allHclFiles := findAllFiles(path, ".hcl")
	var parsedFiles []*hcl.File
	for _, file := range allHclFiles {
		parsedFile, diag := parser.ParseHCLFile(file)
		if diag.HasErrors() {
			return nil, diag
		}
		parsedFiles = append(parsedFiles, parsedFile)
	}

	return hcl.MergeFiles(parsedFiles), nil
}

// CurrentWorkingDir returns the value of the working directory that the CLI is executed in.
func CurrentWorkingDir() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(ex)
}

// getVariableDataFromJSONFile loads a json object file and serializes it into a map of name/value pairs.
// It does not throw an error if the file doesn't exist, but will instead return an empty map. It throws
// an error if the file does exist and is not formatted correctly (i.e. not a basic JSON object), or cannot
// imply the value types.
func getVariableDataFromJSONFile(varFile string) map[string]cty.Value {
	jsonFile, err := os.Open(varFile)
	if err == nil {
		fileBytes, err := io.ReadAll(jsonFile)
		if err != nil {
			log.Fatalf("Failed to read '%s' file: %s", varFile, err)
		}
		varType, err := json.ImpliedType(fileBytes)
		if err != nil {
			log.Fatalf("Failed to decode variables in '%s' file: %s", varFile, err)
		}
		val, err := json.Unmarshal(fileBytes, varType)
		if err != nil {
			log.Fatalf("Failed to decode variables in '%s' file: %s", varFile, err)
		}
		if val.CanIterateElements() == false {
			log.Fatalf("'%s' JSON file must be in object format (key/val)", varFile)
		}
		return val.AsValueMap()
	}
	log.Printf("WARNING: could not open '%s' json file with variable overrides", varFile)
	return make(map[string]cty.Value)
}
