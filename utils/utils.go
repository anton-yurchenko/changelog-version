package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func Wrap(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}

func Fatal(format string, a ...any) {
	fmt.Println(Wrap(format, a...))
	os.Exit(1)
}

func Output(k, v string) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("echo \"%s=%s\" >> $GITHUB_OUTPUT", k, v))
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
