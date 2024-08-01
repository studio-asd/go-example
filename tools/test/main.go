package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/albertwidi/pkg/postgres"
)

var pg *postgres.Postgres

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var err error
	pg, err = postgres.Connect(context.Background(), postgres.ConnectConfig{
		Driver:   "pgx",
		Username: "postgres",
		Password: "postgress",
		Host:     "localhost",
		Port:     "5432",
	})
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	out = bytes.TrimSuffix(out, []byte("\n"))

	// Change the directory to the repository root directory.
	if err := os.Chdir(string(out)); err != nil {
		return err
	}
	// Run docker compose.
	cmd = exec.Command("docker", "compose", "up", "-d")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	defer func() {
		cmd = exec.Command("docker", "compose", "down", "--remove-orphans")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Run()
	}()

	cmd = exec.Command("go", "test", "-v", "-race", "-flaky", "-fullpath", "./...")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// createDatabaseAndApplySchema reads the ./database directory and loop through it. It expects every directory
// is a separate database and has 'schema.sql' inside it.
func createDatabaseAndApplySchema(ctx context.Context, dir string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}
		fmt.Printf("=== %s ===\n", d.Name())

		return nil
	})
}
