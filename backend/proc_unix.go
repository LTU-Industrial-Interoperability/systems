//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

func setProcAttr(cmd *exec.Cmd) {
	// Own process group so SIGTERM reaches the binary spawned by go run.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}
