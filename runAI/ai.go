//go:build windows

package runAI

import (
	"bytes"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// AI 引擎要运行可执行程序作为子进程
// 然后调用方通过和子进程的标准输入输出交互,使ai玩游戏
type AI struct {
	cmd  *exec.Cmd
	send chan string
	lock sync.RWMutex
	buf  bytes.Buffer
}

func NewAI(exe string) (*AI, error) {
	cmd := exec.Command(exe)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true, // 隐藏黑窗
	}

	e := &AI{cmd: cmd, send: make(chan string)}
	cmd.Stdin = e
	cmd.Stdout = e
	return e, cmd.Start()
}

func (e *AI) Read(p []byte) (int, error) {
	return copy(p, <-e.send), nil // chan给ai标准输入发送数据
}

func (e *AI) Write(p []byte) (int, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.buf.Write(p) // 缓存ai标准输出
}

// MatchLineContains 大部分情况只需要当前行包含特定字符串,因此提供通用方法
func MatchLineContains(wait *string) func(cmp []byte) bool {
	cw := []byte(*wait)
	return func(cmp []byte) bool {
		if bytes.Contains(cmp, cw) {
			// 这行数据匹配成功,返回该行数据
			*wait = string(cmp)
			return true
		}
		return false
	}
}

func (e *AI) Send(send string, match func(cmp []byte) bool) {
	e.send <- send + "\n" // 给ai引擎发送指令

	if match != nil {
		for {
			e.lock.RLock()

			// 定时检查标准输出是否匹配某行记录
			for _, val := range bytes.Split(e.buf.Bytes(), []byte{'\n'}) {
				if match(val) {
					e.lock.RUnlock()

					e.lock.Lock()
					e.buf.Reset()
					e.lock.Unlock()
					return
				}
			}

			e.lock.RUnlock()
			time.Sleep(time.Second)
		}
	}
}

func (e *AI) Close() error {
	return e.cmd.Process.Kill() // 关闭ai子进程
}
