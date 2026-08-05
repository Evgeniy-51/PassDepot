//go:build !windows

package gitremote

import "os/exec"

func applySysProcAttr(cmd *exec.Cmd) {}

