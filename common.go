package apt

import (
	"io"
	"os/exec"
)

var executer *exec.Cmd

func SetStdout(w io.Writer) {
	executer.Stdout = w
}

func SetStderr(w io.Writer) {
	executer.Stderr = w
}
