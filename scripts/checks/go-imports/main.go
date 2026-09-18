package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

type fileImports struct {
	Package string   `json:"package"`
	File    string   `json:"file"`
	Imports []string `json:"imports"`
}

func run() error {
	var files []string
	if err := json.NewDecoder(os.Stdin).Decode(&files); err != nil {
		return fmt.Errorf("Go parser input: %w", err)
	}
	results := make([]fileImports, 0, len(files))
	for _, name := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.AllErrors)
		if err != nil {
			return fmt.Errorf("Go parse/read error %s: %w", name, err)
		}
		result := fileImports{Package: parsed.Name.Name, File: name, Imports: []string{}}
		for _, imported := range parsed.Imports {
			value, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return fmt.Errorf("Go import %s: %w", name, err)
			}
			result.Imports = append(result.Imports, value)
		}
		results = append(results, result)
	}
	return json.NewEncoder(os.Stdout).Encode(results)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
