package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// skipDirs contains the directories that we should skip when generating everything.
var skipDirs = []string{
	"testdata",
	"tmp",
	".",
}

var requiredFiles = []string{
	"schema.sql",
	"sqlc.yaml",
}

// SchemaDris retrurns the list of schema directories detected inside the 'dir' parameter.
func SchemaDirs(dbName string, flags Flags, dir string) (schemaDirs []string, skippedDirs []string, err error) {
	if !flags.All {
		if dbName == "" {
			return nil, nil, errors.New("database name cannot be empty if --all flag is not used")
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
		if filepath.Base(path) != path {
			return filepath.SkipDir
		}
		if slices.Contains(skipDirs, d.Name()) {
			skippedDirs = append(skippedDirs, d.Name())
			return nil
		}
		// Check the required files list, and skip the directory if all required files are not inside the directory.
		for _, mf := range requiredFiles {
			_, err = os.Stat(filepath.Join(d.Name(), mf))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					skippedDirs = append(skippedDirs, d.Name())
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
