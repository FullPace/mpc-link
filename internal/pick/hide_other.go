//go:build !windows

package pick

import "os/exec"

func hide(*exec.Cmd) {}
