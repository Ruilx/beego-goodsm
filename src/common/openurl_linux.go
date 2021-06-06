// +build linux

package common

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

const COMMAND = "xdg-open"

func getCommandStr() string{
	return COMMAND
}

func parsingUri(uri string) string{
	newUri := uri
	if !strings.HasPrefix(newUri, "http://") && !strings.HasPrefix(newUri, "https://"){
		newUri = "http://" + uri
	}
	return newUri
}

func openUriUsingCommand(uri string) error {
	if runtime.GOOS != "linux" {
		return errors.New("not a correct operating system, only Linux with Xdg can use this script")
	}
	cmd := exec.Command(getCommandStr(), parsingUri(uri))
	return cmd.Start()
}