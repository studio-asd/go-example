package main

import (
	"context"
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

func SchemaDirs(dbName string, flags Flags, dir string) (schemaDirs []string, err error) {
	if !flags.All {
		if dbName == "" {
			return nil, errors.New("database name cannot be empty if --all flag is not used")
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
			return nil
		}
		// Check whether the 'schema.sql' is exists within the directory.
		_, err = os.Stat(filepath.Join(d.Name(), "schema.sql"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return filepath.SkipDir
			}
			return err
		}
		schemaDirs = append(schemaDirs, d.Name())
		return nil
	})
	return
}

func CreateschemaDirsAndApplySchema(ctx context.Context) {
}
