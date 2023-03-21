package runAI

import (
	"bytes"
	"os/exec"
	"sync"
	"time"
)

// AI 引擎要运行可执行程序作为子进程
// 然后调用方通过和子进程的标准输入输出交互,使ai玩游戏
type AI struct {
	cmd  *exec.Cmd
	send chan string
	lock sync.Mutex
	buf  bytes.Buffer
}

var setSysProcAttr func(*exec.Cmd)

func NewAI(exe string) (*AI, error) {
	cmd := exec.Command(exe)
	if setSysProcAttr != nil {
		setSysProcAttr(cmd)
	}

	e := &AI{cmd: cmd, send: make(chan string)}
	cmd.Stdin = e
	cmd.Stdout = e
	return e, cmd.Start()
}

func (pg *AI) Read(p []byte) (int, error) {
	return copy(p, <-pg.send), nil // chan给ai标准输入发送数据
}

func (pg *AI) Write(p []byte) (n int, err error) {
	pg.lock.Lock()
	n, err = pg.buf.Write(p) // 缓存标准输出
	pg.lock.Unlock()
	return
}

type MatchFunc func(line []byte) bool

// MatchLineContains 大部分情况只需要当前行包含特定字符串
func MatchLineContains(wait *string) MatchFunc {
	cw := []byte(*wait)
	return func(line []byte) bool {
		if bytes.Contains(line, cw) {
			*wait = string(line) // 匹配成功,返回该行数据
			return true
		}
		return false
	}
}

var newLine = []byte{'\n'}

func (pg *AI) Send(send string, match MatchFunc) {
	pg.send <- send + "\n" // 发送指令到标准输入

	if match != nil {
		for {
			pg.lock.Lock()
			for _, line := range bytes.Split(pg.buf.Bytes(), newLine) {
				if match(line) {
					pg.buf.Reset()
					pg.lock.Unlock()
					return // 匹配成功,清理缓存,立即返回
				}
			}
			pg.lock.Unlock() // 匹配失败,延时重试
			time.Sleep(time.Millisecond * 200)
		}
	}
}

func (pg *AI) Close() error {
	return pg.cmd.Process.Kill() // 关闭ai子进程
}
