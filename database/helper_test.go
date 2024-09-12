package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSchemasInDirectory(t *testing.T) {
	t.Parallel()

	testDir := t.TempDir()

	tests := []struct {
		name    string
		files   map[string]string
		expects []SQLSchema
	}{
		{
			name: "list schema dirs",
			files: map[string]string{
				"test-schema.sql": `
CREATE TABLE something(
	id int PRIMARY KEY
);
`,
				"another.schema.sql": `
CREATE TABLE another_something(
	id int PRIMARY KEY
);
`,
				"testing.sql": "a testing",
				"schema.txt":  "a schema with txt",
			},
			expects: []SQLSchema{
				{
					FileName: "test-schema.sql",
					Content: []byte(`
CREATE TABLE something(
	id int PRIMARY KEY
);
`,
					),
				},
				{
					FileName: "another.schema.sql",
					Content: []byte(`
CREATE TABLE another_something(
	id int PRIMARY KEY
);
`,
					),
				},
				{
					FileName: "testing.sql",
					Content:  []byte("a testing"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for fname, content := range test.files {
				fname := filepath.Join(testDir, fname)
				f, err := os.Create(filepath.Join(fname))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					f.Close()
				})
				if _, err := f.Write([]byte(content)); err != nil {
					t.Fatal(err)
				}
				f.Close()
				if err != nil {
					t.Fatal(err)
				}
			}
			schemas, err := SchemasInDirectory(testDir)
			if err != nil {
				t.Fatal(err)
			}
			// Find the schema file and content.
			for _, e := range test.expects {
				var found bool
				for _, sc := range schemas {
					if e.FileName == sc.FileName {
						found = true
						if diff := cmp.Diff(e, sc); diff != "" {
							t.Fatalf("(-want/+got)\n%s", diff)
						}
						break
					}
				}
				if !found {
					t.Fatalf("schema file %s not found\nschemas: %v", e.FileName, schemas)
				}
			}
		})
	}

}

func TestSchemaDirs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dirs          []string
		requiredFiles [][]string
		dbName        string
		flags         Flags
		expect        []string
	}{
		{
			name: "one_database",
			dirs: []string{
				"one",
				"two",
			},
			requiredFiles: [][]string{
				{
					"schema.sql",
					"sqlc.yaml",
				},
			},
			dbName: "one",
			flags:  Flags{},
			expect: []string{
				"one",
			},
		},
		{
			name: "one_database_no_required_files",
			dirs: []string{
				"one",
				"two",
			},
			requiredFiles: [][]string{
				{
					"schema.sql",
					"another-schema.sql",
				},
			},
			dbName: "one",
			flags:  Flags{},
			expect: []string{},
		},
		{
			name: "one_with_all_flags",
			dirs: []string{
				"one",
				"two",
				"three",
			},
			requiredFiles: [][]string{
				{
					"schema.sql",
					"sqlc.yaml",
				},
				{
					"schema.sql",
					"sqlc.yaml",
				},
				{
					"schema.sql",
					"sqlc.yaml",
				},
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
			requiredFiles: [][]string{
				{
					"schema.sql",
					"sqlc.yaml",
				},
				{
					"schema.sql",
					"sqlc.yaml",
				},
				{
					"schema.sql",
					"sqlc.yaml",
				},
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
				for idx, dir := range test.dirs {
					if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o766); err != nil {
						t.Fatal(err)
					}
					// Create the required files.
					if len(test.requiredFiles) > idx {
						for _, fname := range test.requiredFiles[idx] {
							fullPath := filepath.Join(tmpDir, dir, fname)
							f, err := os.Create(fullPath)
							if err != nil {
								t.Fatal(err)
							}
							if err := f.Close(); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
			}
			dirs, skippedDirs, err := SchemaDirs(test.dbName, test.flags, tmpDir)
			if err != nil {
				t.Fatal(err)
			}

			// Check whether the length of the directories match.
			if len(test.expect) != len(dirs) {
				t.Fatalf("expecting %d dirs but got %d", len(test.expect), len(dirs))
			}
			// Loop through to check whether the expect database is there or not. We cannot do compare because
			// we will always have extra dirs from the go test.
			for _, ex := range test.expect {
				var found bool
				for _, db := range dirs {
					if ex == db {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("directory with name %s not found, %v\nSkippedDirs: %v", ex, dirs, skippedDirs)
				}
			}
		})
	}
}
