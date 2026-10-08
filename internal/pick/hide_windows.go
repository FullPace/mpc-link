package pick

import (
	"os/exec"
	"syscall"
)

// hide keeps the PowerShell window that runs the folder chooser out of sight.
func hide(cmd *exec.Cmd) { cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} }
