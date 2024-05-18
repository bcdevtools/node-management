package keys

import (
	"fmt"
	"github.com/bcdevtools/node-management/utils"
	"os"
	"path"
)

func addSshKey(keyName string) {
	sshDir := prepareSshDir()

	keyPath := path.Join(sshDir, keyName)
	_, exists, _, err := utils.FileInfo(keyPath)
	if err != nil {
		utils.ExitWithErrorMsgf(`ERR: failed to check key file at %s
%v
`, keyPath, err)
		return
	}
	if exists {
		utils.ExitWithErrorMsgf("ERR: key file %s already exists\n", keyPath)
		return
	}

	fmt.Println("Creating key file", keyName)
	fmt.Println("at", keyPath)
	fmt.Println("\nConfirm? (y/n)")
	if !utils.ReadYesNo() {
		utils.ExitWithErrorMsg("ERR: operation canceled")
		return
	}

	ec := utils.LaunchApp("ssh-keygen", []string{"-t", "ed25519", "-C", keyName, "-f", keyPath, "-N", ""})
	if ec != 0 {
		utils.ExitWithErrorMsgf("ERR: ssh-keygen exited with code %d\n", ec)
		return
	}
	fmt.Println("Key file created successfully at", keyPath)
}

func prepareSshDir() string {
	homeDir := utils.MustGetCurrentUserHomeDirectory()

	sshDir := path.Join(homeDir, ".ssh")
	_, exists, isDir, err := utils.FileInfo(sshDir)
	if err != nil {
		utils.ExitWithErrorMsgf(`ERR: failed to check .ssh directory at %s
%v
`, sshDir, err)
		return ""
	}
	if !exists {
		fmt.Println("Creating .ssh directory at", sshDir)
		err = os.Mkdir(sshDir, 0750)
		if err != nil {
			utils.ExitWithErrorMsgf(`ERR: failed to create .ssh directory at %s
%v
`, sshDir, err)
			return ""
		}
	}
	if !isDir {
		utils.ExitWithErrorMsgf("ERR: %s is not a directory\n", sshDir)
		return ""
	}

	return sshDir
}
