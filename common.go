package apt

import (
	"io"
	"os/exec"
)

// AptManager creates an apt manager
type AptManager struct {
	executer *exec.Cmd
}

// SetStdout sets stdout writer
func (am *AptManager) SetStdout(w io.Writer) {
	am.executer.Stdout = w
}

// SetStderr sets stderr writer
func (am *AptManager) SetStderr(w io.Writer) {
	am.executer.Stderr = w
}
