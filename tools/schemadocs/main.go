package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Column struct {
	Name        string
	DataType    string
	IsNullable  bool
	IsPrimary   bool
	Description string
}

type Table struct {
	Name        string
	Columns     []Column
	Description string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <schema_file>")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	tables := parseSQLSchema(file)
	generateMarkdown(tables)
}

func parseSQLSchema(file *os.File) []Table {
	scanner := bufio.NewScanner(file)
	var tables []Table
	var currentTable *Table
	var tableComment string
	var columnComment string

	// Regex patterns
	createTablePattern := regexp.MustCompile(`(?i)create\s+table\s+if\s+not\s+exists\s+([^\s(]+)\s*\(`)
	columnPattern := regexp.MustCompile(`^\s*"?([^"\s]+)"?\s+([^,\s]+)(.*)$`)
	commentPattern := regexp.MustCompile(`--\s*(.*)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and pure comment lines that don't belong to columns
		if line == "" || (strings.HasPrefix(line, "--") && currentTable == nil) {
			continue
		}

		// Extract comments
		if strings.Contains(line, "--") {
			if matches := commentPattern.FindStringSubmatch(line); len(matches) > 1 {
				if currentTable == nil {
					tableComment = matches[1]
				} else {
					columnComment = matches[1]
				}
			}
			line = commentPattern.ReplaceAllString(line, "")
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
		}

		// Check for CREATE TABLE
		if matches := createTablePattern.FindStringSubmatch(line); len(matches) > 1 {
			if currentTable != nil {
				tables = append(tables, *currentTable)
			}
			tableName := strings.Trim(matches[1], `"`)
			currentTable = &Table{
				Name:        tableName,
				Description: tableComment,
			}
			tableComment = "" // Reset table comment
			continue
		}

		// Check for end of table
		if strings.Contains(line, ");") {
			if currentTable != nil {
				tables = append(tables, *currentTable)
				currentTable = nil
			}
			continue
		}

		// Parse column if we're inside a table
		if currentTable != nil && !strings.HasPrefix(line, "CREATE") && !strings.HasPrefix(line, "DROP") {
			line = strings.TrimSuffix(line, ",")
			if matches := columnPattern.FindStringSubmatch(line); len(matches) > 1 {
				columnName := matches[1]
				dataType := matches[2]
				constraints := strings.ToUpper(matches[3])

				column := Column{
					Name:        columnName,
					DataType:    dataType,
					IsNullable:  !strings.Contains(constraints, "NOT NULL"),
					IsPrimary:   strings.Contains(constraints, "PRIMARY KEY"),
					Description: columnComment,
				}
				currentTable.Columns = append(currentTable.Columns, column)
				columnComment = "" // Reset column comment
			}
		}
	}

	return tables
}

func generateMarkdown(tables []Table) {
	for _, table := range tables {
		// Print table name as header
		fmt.Printf("**%s Table:**\n", table.Name)

		// Print column headers
		fmt.Println("| Column Name | Data Type | Nullable | Primary Key | Description |")
		fmt.Println("|------------|-----------|----------|-------------|-------------|")

		// Print each column
		for _, col := range table.Columns {
			nullable := "Yes"
			if !col.IsNullable {
				nullable = "No"
			}
			primary := "No"
			if col.IsPrimary {
				primary = "Yes"
			}

			fmt.Printf("| %s | %s | %s | %s | %s |\n",
				col.Name,
				col.DataType,
				nullable,
				primary,
				col.Description)
		}

		// Print table description if exists
		if table.Description != "" {
			fmt.Printf("\n*Table Description: %s*\n", table.Description)
		}
		fmt.Println()
	}
}
