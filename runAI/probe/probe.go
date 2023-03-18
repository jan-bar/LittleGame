package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	err := probe()
	if err != nil {
		panic(err)
	}
}

// 用于探测ai引擎的标准输入输出语法
func probe() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dir := filepath.Dir(exe)

	b, err := os.ReadFile(filepath.Join(dir, "run.json"))
	if err != nil {
		return err
	}

	var cnf struct {
		Exe string `json:"exe"`
		In  string `json:"in"`
		Out string `json:"out"`
	}
	err = json.Unmarshal(b, &cnf)
	if err != nil {
		return err
	}

	if cnf.In == "" {
		cnf.In = "in.txt"
	}
	if cnf.Out == "" {
		cnf.Out = "out.txt"
	}

	rw, err := NewRW(os.Stdin, os.Stdout,
		filepath.Join(dir, cnf.In), filepath.Join(dir, cnf.Out))
	if err != nil {
		return err
	}

	cmd := exec.Command(cnf.Exe)
	cmd.Stdin = rw
	cmd.Stdout = rw
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	return cmd.Run()
}

type RW struct {
	r io.Reader
	w io.Writer

	in, out *os.File
}

func NewRW(r io.Reader, w io.Writer, in, out string) (*RW, error) {
	var (
		rw  = RW{r: r, w: w}
		err error
	)
	rw.in, err = os.Create(in)
	if err != nil {
		return nil, err
	}

	rw.out, err = os.Create(out)
	if err != nil {
		return nil, err
	}
	return &rw, nil
}

func (rw *RW) Read(p []byte) (n int, err error) {
	n, err = rw.r.Read(p)
	if err == nil {
		rw.in.Write(p[:n])
		rw.in.Sync()
	}
	return
}

func (rw *RW) Write(p []byte) (n int, err error) {
	n, err = rw.w.Write(p)
	if err == nil {
		rw.out.Write(p[:n])
		rw.out.Sync()

		if bytes.Contains(p[:n], []byte("bye")) {
			err = io.ErrShortWrite
		}
	}
	return
}

func (rw *RW) Close() error {
	ei := rw.in.Close()
	eo := rw.out.Close()
	if ei != nil || eo != nil {
		return fmt.Errorf("in:%v,out:%v", ei, eo)
	}
	return nil
}
