package utils

import (
	"os"
	"os/exec"
)

func LaunchApp(appName string, args []string) int {
	return LaunchAppWithSetup(appName, args, func(launchCmd *exec.Cmd) {
		launchCmd.Stdin = os.Stdin
		launchCmd.Stdout = os.Stdout
		launchCmd.Stderr = os.Stderr
	})
}

func LaunchAppWithSetup(appName string, args []string, setup func(launchCmd *exec.Cmd)) int {
	launchCmd := exec.Command(appName, args...)
	setup(launchCmd)
	err := launchCmd.Run()
	if err != nil {
		PrintlnStdErr("ERR: problem when running process", appName)
		PrintlnStdErr(err)
		return 1
	}
	return 0
}
