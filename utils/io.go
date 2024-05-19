package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

func FileInfo(path string) (fileMode os.FileMode, exists, isDir bool, error error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		error = err
		return
	}

	fileMode = fi.Mode().Perm()
	exists = true
	isDir = fi.IsDir()
	return
}

func IsEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f.Close()
	}()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err // Either not empty or error, suits both cases
}

var mutexReader sync.Mutex
var singletonReader *bufio.Reader

func GetReader() *bufio.Reader {
	mutexReader.Lock()
	defer mutexReader.Unlock()

	if singletonReader == nil {
		singletonReader = bufio.NewReader(os.Stdin)
	}

	return singletonReader
}

func ReadYesNo() bool {
	for {
		text := readNormalizedText()
		switch text {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			fmt.Println("Please enter Yes or No")
		}
	}
}

func ReadNumber(min, max int64) int64 {
	for {
		text := readNormalizedText()
		number, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			if number >= min && number <= max {
				return number
			}
		}

		fmt.Println("Please enter a number between", min, "and", max)
	}
}

func ReadOptionalNumber(min, max, _default int64) int64 {
	for {
		text := readNormalizedText()
		if text == "" {
			return _default
		}
		number, err := strconv.ParseInt(text, 10, 64)
		if err == nil {
			if number >= min && number <= max {
				return number
			}
		}

		fmt.Println("Please enter a number between", min, "and", max)
	}
}

func ReadText(allowEmpty bool) string {
	for {
		text := readText()

		if allowEmpty {
			return text
		}

		if text != "" {
			return text
		}

		fmt.Println("Please enter some text")
	}
}

func readNormalizedText() string {
	text, _ := GetReader().ReadString('\n')
	return strings.ToLower(strings.TrimSpace(text))
}

func readText() string {
	text, _ := GetReader().ReadString('\n')
	return strings.TrimSpace(text)
}
