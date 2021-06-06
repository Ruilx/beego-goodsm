// +build windows

package common

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const COMMAND = "cmd"

func getCommandStr() string{
	//return strings.Join([]string{COMMAND, "/c", "start", "\"\""}, " ")
	return COMMAND
}

func parsingUri(uri string) string{
	newUri := uri
	if !strings.HasPrefix(newUri, "http://") && !strings.HasPrefix(newUri, "https://"){
		newUri = "http://" + uri
	}
	return newUri
}

func getCommandArguments(uri string) []string{
	return []string{"/c", "start", parsingUri(uri)}
}

func openUriUsingCommand(uri string) error {
	if runtime.GOOS != "windows"{
		return errors.New("not a correct operating system, only Windows can use this script")
	}
	fmt.Println(getCommandStr(), parsingUri(uri))
	cmd := exec.Command(getCommandStr(), getCommandArguments(uri)...)
	return cmd.Start()
}
