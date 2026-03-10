//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
)

func setProcAttr(cmd *exec.Cmd) {
	// New process group on Windows so we can target it with taskkill.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func killProcess(cmd *exec.Cmd) error {
	// taskkill /F /T terminates the process and all its children (covers go run's child binary).
	kill := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid))
	if err := kill.Run(); err != nil {
		return fmt.Errorf("taskkill: %w", err)
	}
	return nil
}
