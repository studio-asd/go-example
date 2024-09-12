package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// skipDirs contains the directories that we should skip when generating everything.
var skipDirs = []string{
	"testdata",
	"tmp",
	".",
}

var requiredFiles = []string{
	"sqlc.yaml",
}

// SchemaDris retrurns the list of schema directories detected inside the 'dir' parameter.
func SchemaDirs(dbName string, flags Flags, dir string) (schemaDirs []string, skippedDirs []string, err error) {
	if !flags.All {
		if dbName == "" {
			return nil, nil, errors.New("database name cannot be empty if --all flag is not used")
		}
		if slices.Contains(skippedDirs, dbName) {
			skippedDirs = append(skippedDirs, dbName)
			return
		}

		fullPath := filepath.Join(dir, dbName)
		// Check whether there are schemas inside the directory.
		schemas, errSc := SchemasInDirectory(fullPath)
		if errSc != nil {
			return nil, nil, errSc
		}
		if len(schemas) == 0 {
			skippedDirs = append(skippedDirs, dbName)
			return
		}
		// Check the required files list, and skip the directory if all required files are not
		// inside the directory.
		for _, mf := range requiredFiles {
			_, err = os.Stat(filepath.Join(dir, dbName, mf))
			if err != nil {
				skippedDirs = append(skippedDirs, dbName)
				if os.IsNotExist(err) {
					// Append the skipped dirs while return a nil error because it is sometimes expected for
					// the file to not exist.
					err = nil
					return
				}
				return
			}
		}
		schemaDirs = append(schemaDirs, dbName)
		return
	}

	// Give the user a warning when --all is being set but directory name is also passed.
	if dbName != "" {
		fmt.Println("[WARNING] schema directory name ignored when --all is being set")
	}

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if dir == path {
			return nil
		}
		if slices.Contains(skipDirs, d.Name()) {
			skippedDirs = append(skippedDirs, d.Name())
			return nil
		}

		// Check whether there are schemas inside the directory.
		schemas, errSc := SchemasInDirectory(path)
		if errSc != nil {
			return errSc
		}
		if len(schemas) == 0 {
			skippedDirs = append(skippedDirs, dbName)
			return nil
		}
		// Check the required files list, and skip the directory if all required files are not
		// inside the directory.
		for _, mf := range requiredFiles {
			_, err = os.Stat(filepath.Join(path, mf))
			if err != nil {
				skippedDirs = append(skippedDirs, d.Name())
				if errors.Is(err, os.ErrNotExist) {
					return filepath.SkipDir
				}
				return err
			}
		}
		schemaDirs = append(schemaDirs, d.Name())
		return nil
	})
	return
}

// ReadSQLCConfiguration returns SQLCConfiguration within a directory with a specific file name.
func ReadSQLCConfiguration(dir, fileName string) (SQLCConfig, error) {
	var config SQLCConfig

	out, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		return config, nil
	}
	if err := yaml.Unmarshal(out, &config); err != nil {
		return config, err
	}
	return config, nil
}

type SQLSchema struct {
	FileName string
	Content  []byte
}

// SchemasInDirectory returns all schemas within a directory.
func SchemasInDirectory(dir string) ([]SQLSchema, error) {
	var schemas []SQLSchema
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if dir == path {
			return nil
		}
		// Skip the file if the name is not contains schema and not a sql file.
		if !strings.Contains(d.Name(), "schema") && filepath.Ext(d.Name()) != ".sql" {
			return nil
		}
		out, err := os.ReadFile(filepath.Join(dir, d.Name()))
		if err != nil {
			return fmt.Errorf("failed to read schema %s. Error: %w", d.Name(), err)
		}
		schemas = append(schemas, SQLSchema{
			FileName: d.Name(),
			Content:  out,
		})
		return nil
	})
	return schemas, err
}
