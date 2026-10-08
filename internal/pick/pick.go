// Package pick opens the system's folder chooser.
package pick

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// ErrCancelled: the user closed the chooser without a folder.
var ErrCancelled = errors.New("cancelled")

// Folder shows the system folder chooser and returns the chosen path.
func Folder(prompt string) (string, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		script := `POSIX path of (choose folder with prompt "` + strings.ReplaceAll(prompt, `"`, `'`) + `")`
		cmd = exec.Command("osascript", "-e", script)
	case "windows":
		ps := `Add-Type -AssemblyName System.Windows.Forms; $d = New-Object System.Windows.Forms.FolderBrowserDialog; ` +
			`$d.Description = '` + strings.ReplaceAll(prompt, `'`, `''`) + `'; ` +
			`if ($d.ShowDialog() -eq 'OK') { $d.SelectedPath }`
		cmd = exec.Command("powershell", "-NoProfile", "-STA", "-Command", ps)
	default:
		return "", errors.New("no folder chooser on " + runtime.GOOS)
	}
	out, err := cmd.Output()
	p := strings.TrimRight(strings.TrimSpace(string(out)), `/\`)
	if err != nil || p == "" {
		return "", ErrCancelled
	}
	return p, nil
}
