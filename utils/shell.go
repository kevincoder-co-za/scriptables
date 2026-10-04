package utils

import (
	"os"
	"os/exec"
	"strings"
)

func RunScript(script string) (string, error) {
	cmd := exec.Command("bash", "-s")
	cmd.Stdin = strings.NewReader(script)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func RunCommandAsRoot(name string, args ...string) (string, error) {
	sudoArgs := append([]string{"-n", name}, args...)
	output, err := exec.Command("sudo", sudoArgs...).CombinedOutput()
	return string(output), err
}
