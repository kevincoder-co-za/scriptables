package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GlobDirectory(dirPath string) ([]string, error) {
	defer func() {

		if r := recover(); r != nil {
			fmt.Println("Serious error: cannot find any scriptables for server type:", dirPath, r)
		}

	}()

	files, err := filepath.Glob(filepath.Join("./scriptables/", dirPath, "*.sh"))
	if err != nil {
		return nil, err
	}

	return files, nil
}

func GetScriptables(scriptableList string) []string {
	var scriptables []string

	if strings.Contains(scriptableList, ",") {
		scriptables = strings.Split(scriptableList, ",")
	} else {
		scriptables = append(scriptables, scriptableList)
	}

	scripts := []string{}

	for _, scriptable := range scriptables {
		scripts_found, err := GlobDirectory(scriptable)
		if len(scripts_found) == 0 || err != nil {
			fmt.Println("No scriptable found.", scriptable, err)
			return []string{}
		}

		scripts = append(scripts, scripts_found...)
	}

	return scripts
}

func GetSharedScriptable(name string) (string, error) {
	data, err := os.ReadFile(filepath.Join("./scriptables/__shared/" + name + ".sh"))
	return string(data), err

}
