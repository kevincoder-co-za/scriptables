package utils

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

const defaultPanelPort = "1323"

func PanelPort() string {
	if port := os.Getenv("SCRIPTABLES_SERVER_DSN_PORT"); port != "" {
		return port
	}

	return defaultPanelPort
}

func HashSystemPassword(password string) (string, error) {
	cmd := exec.Command("openssl", "passwd", "-6", "-stdin")
	cmd.Stdin = strings.NewReader(password + "\n")

	output, err := cmd.Output()
	hash := strings.TrimSpace(string(output))
	if err != nil || !strings.HasPrefix(hash, "$6$") {
		return "", errors.New("openssl could not create a password hash")
	}

	return hash, nil
}
