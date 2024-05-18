package utils

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

func HasBinaryName(binaryName string) bool {
	_, err := exec.LookPath(binaryName)
	return err == nil
}

func TryExtractUserHomeDirFromPath(path string) (string, error) {
	path = strings.TrimSuffix(path, "/")

	if path == "~" || strings.HasPrefix(path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return usr.HomeDir, nil
	}

	if !strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("path must be absolute")
	}

	if path == "/root" || strings.HasPrefix(path, "/root/") {
		return "/root", nil
	}

	if !strings.HasPrefix(path, "/home/") && !strings.HasPrefix(path, "/Users/") {
		if IsLinux() {
			return "", fmt.Errorf("path must be under /home")
		} else if IsDarwin() {
			return "", fmt.Errorf("path must be under /Users")
		} else {
			return "", fmt.Errorf("path must be under /home or /Users")
		}
	}

	spl := strings.Split(path, "/")
	var nonEmpty []string
	for _, s := range spl {
		if s != "" {
			nonEmpty = append(nonEmpty, s)
		}
	}

	if len(nonEmpty) < 2 {
		return "", fmt.Errorf("path is not user home")
	}

	return fmt.Sprintf("/%s/%s", nonEmpty[0], nonEmpty[1]), nil
}
