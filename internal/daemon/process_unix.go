//go:build !windows

package daemon

import (
	"os/exec"
	"runtime"
	"syscall"
)

func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func gracefulKill(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// killStaleServer kills any leftover whisper-server process from a previous run.
func killStaleServer() {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("killall", "-9", "whisper-server").Run()
	default:
		exec.Command("pkill", "-9", "-x", "whisper-server").Run()
	}
}
