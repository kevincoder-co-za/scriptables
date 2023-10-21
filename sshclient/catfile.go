package sshclient

func ReadFileWithSudo(client *Client, filePath string) (string, error) {

	cmd, err := client.Cmd("sudo cat " + filePath).Output()
	if err != nil {
		return "", err
	}

	return string(cmd), nil

}
