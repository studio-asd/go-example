// The test of database helper script is aimed for several things:
//
// 1. To ensure we are generating a working go code.
// 2. To ensure we are generating the correct types from sqlc.yaml.

package main

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"
)

//go:embed docker-compose.yaml sqlc.yaml main.go
var embeddedTestFiles embed.FS

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, nil))
}

// TestRunner uses testscript to run tests from testdata/script/*.txtar.
func TestScript(t *testing.T) {
	t.Skip()

	// Delete all databases created in the testdata/*.
	t.Cleanup(func() {
		// databases := []string{
		// 	"orders",
		// }
		// for _, db := range databases {
		// 	execQuery(
		// 		"postgres://postgres:postgres@localhost:5432?sslmode=disable",
		// 		fmt.Sprintf("DROP DATABASE IF EXISTS %s;", db),
		// 	)
		// }
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
			args: []string{"--sqlc_config=something.yaml"},
			expect: Flags{
				SQLCConfig: "something.yaml",
			},
		},
		{
			args: []string{"--sqlc_config=something.yaml", "--db_name=test"},
			expect: Flags{
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
