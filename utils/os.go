package utils

import (
	"github.com/pkg/errors"
	"os/user"
)

func MustGetCurrentUsername() string {
	usr, err := user.Current()
	if err != nil {
		panic(errors.Wrap(err, "failed to get current user"))
	}
	return usr.Username
}

func MustGetCurrentUserHomeDirectory() string {
	usr, err := user.Current()
	if err != nil {
		panic(errors.Wrap(err, "failed to get current user"))
	}
	return usr.HomeDir
}

func MustNotUserRoot() {
	if MustGetCurrentUsername() == "root" {
		panic("this action should not be run as root")
	}
}
