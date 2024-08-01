package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
)

// skipDirs contains the directories that we should skip when generating everything.
var skipDirs = []string{
	"testdata",
	"tmp",
	".",
}

// DatabaseList retrieves the list of databases for the generator. The list is also a directory as
// we construct the directories to be per-database basis. Please note that this function will perform a walkFunc
// from the current directory.
func DatabaseList(dbName string, flags Flags, dir string) (databases []string, err error) {
	if !flags.All {
		if dbName == "" {
			return nil, errors.New("database name cannot be empty if --all flag is not used")
		}
		databases = append(databases, dbName)
		return
	}
	// Give the user a warning when --all is being set but database name is also passed.
	if dbName != "" {
		fmt.Println("[WARNING] database name ignored when --all is being set")
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
		databases = append(databases, d.Name())
		return nil
	})
	return
}

func CreateDatabasesAndApplySchema(ctx context.Context) {
}
