//go:build windows

package runAI

import (
	"os/exec"
	"syscall"
)

func init() {
	setSysProcAttr = func(cmd *exec.Cmd) {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true, // windows 隐藏黑窗
		}
	}
}
