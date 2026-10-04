package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const scriptablesDirectory = "./scriptables"
const importDirective = "SCRIPTABLE::IMPORT"
const maxImportDepth = 5

func FindScriptables(scriptableList string) []string {
	scripts := []string{}

	for _, scriptable := range strings.Split(scriptableList, ",") {
		found, err := filepath.Glob(filepath.Join(scriptablesDirectory, strings.TrimSpace(scriptable), "*.sh"))
		if err != nil || len(found) == 0 {
			fmt.Println("No scriptable found.", scriptable, err)
			return []string{}
		}

		scripts = append(scripts, found...)
	}

	return scripts
}

func ReadSharedScriptable(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join(scriptablesDirectory, "__shared", name+".sh"))
	return string(data), err
}

func ExpandScriptableImports(script string) (string, error) {
	for depth := 0; depth < maxImportDepth && strings.Contains(script, importDirective); depth++ {
		for _, line := range strings.Split(script, "\n") {
			if !strings.Contains(line, importDirective) {
				continue
			}

			name := strings.TrimSpace(strings.ReplaceAll(line, importDirective, ""))
			imported, err := ReadSharedScriptable(name)
			if err != nil {
				return script, fmt.Errorf("failed to import shared scriptable %q: %w", name, err)
			}

			script = strings.ReplaceAll(script, line, imported)
		}
	}

	return script, nil
}

func MustStopOnFailure(script string) bool {
	return strings.Contains(script, "# exit-on-failure=yes")
}
