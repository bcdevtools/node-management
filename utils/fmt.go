package utils

import (
	"fmt"
	"os"
)

func PrintlnStdErr(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}

func PrintfStdErr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

func ExitWithErrorMsg(error ...any) {
	PrintlnStdErr()
	PrintlnStdErr(error)
	os.Exit(1)
}

func ExitWithErrorMsgf(format string, a ...any) {
	PrintlnStdErr()
	PrintfStdErr(format, a...)
	os.Exit(1)
}
