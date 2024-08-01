// The test of database helper script is aimed for several things:
//
// 1. To ensure we are generating a working go code.
// 2. To ensure we are generating the correct types from sqlc.yaml.

package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"
)

//go:embed docker-compose.yaml sqlc.yaml main.go helper.go
var embeddedTestFiles embed.FS

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, nil))
}

// TestRunner uses testscript to run tests from testdata/script/*.txtar.
func TestScript(t *testing.T) {
	// Delete all databases created in the testdata/*.
	t.Cleanup(func() {
		databases := []string{
			"orders",
		}
		for _, db := range databases {
			execQuery(
				"postgres://postgres:postgres@localhost:5432?sslmode=disable",
				fmt.Sprintf("DROP DATABASE IF EXISTS %s;", db),
			)
		}
	})

	testscript.Run(t, testscript.Params{
		Dir: "testdata",
		Setup: func(e *testscript.Env) error {
			entries, err := embeddedTestFiles.ReadDir(".")
			if err != nil {
				return err
			}
			for _, entry := range entries {
				out, err := embeddedTestFiles.ReadFile(entry.Name())
				if err != nil {
					return err
				}
				// Create everything inside the ./database directory to mimic the current repository condition.
				if err := os.WriteFile(filepath.Join(e.Cd, "database", entry.Name()), out, 0o666); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

func TestParseFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args   []string
		expect Flags
	}{
		{
			args: []string{"--all"},
			expect: Flags{
				All:        true,
				SQLCConfig: "sqlc.yaml",
			},
		},
		{
			args: []string{"--all", "--sqlc_config=something.yaml"},
			expect: Flags{
				All:        true,
				SQLCConfig: "something.yaml",
			},
		},
		{
			args: []string{"--all", "--sqlc_config=something.yaml", "--db_name=test"},
			expect: Flags{
				All:        true,
				SQLCConfig: "something.yaml",
				DBName:     "test",
			},
		},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.args, "_"), func(t *testing.T) {
			t.Parallel()

			got, err := parseFlags(test.args)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(test.expect, got); diff != "" {
				t.Fatalf("(-want/+got)\n%s", diff)
			}
		})
	}
}

func TestDatabaseList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		dirs   []string
		dbName string
		flags  Flags
		expect []string
	}{
		{
			name: "one_databse",
			dirs: []string{
				"one",
				"two",
			},
			dbName: "one",
			flags:  Flags{},
			expect: []string{
				"one",
			},
		},
		{
			name: "one_with_all_flags",
			dirs: []string{
				"one",
				"two",
				"three",
			},
			dbName: "one",
			flags: Flags{
				All: true,
			},
			expect: []string{
				"one",
				"two",
				"three",
			},
		},
		{
			name: "none_with_all_flags",
			dirs: []string{
				"one",
				"two",
				"three",
			},
			flags: Flags{
				All: true,
			},
			expect: []string{
				"one",
				"two",
				"three",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			if len(test.dirs) > 0 {
				for _, dir := range test.dirs {
					if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o766); err != nil {
						t.Fatal(err)
					}
				}
			}
			databases, err := DatabaseList(test.dbName, test.flags, tmpDir)
			if err != nil {
				t.Fatal(err)
			}

			// Loop through to check whether the expect database is there or not. We cannot do compare because
			// we will always have extra dirs from the go test.
			for _, ex := range test.expect {
				var found bool
				for _, db := range databases {
					if ex == db {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("database with name %s not found, %v", ex, databases)
				}
			}
		})
	}
}
