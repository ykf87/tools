//go:build !windows

package hideconsole

import (
	"os"
	"os/exec"
	"syscall"
)

func HideConsole() {
	if os.Getenv("DAEMON") == "1" {
		return
	}

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "DAEMON=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	cmd.Start()
	os.Exit(0)
}
