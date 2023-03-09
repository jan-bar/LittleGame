package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

func main() {
	port := flag.Uint("addr", 8080, "listen address")
	dir := flag.String("dir", ".", "files directory to serve")
	flag.Parse()

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("listening on %q...", addr)
	go func() {
		time.Sleep(time.Second * 2) // 2秒以后浏览器打开一个网页
		//goland:noinspection HttpUrlsUsage
		_, _ = exec.Command("explorer", "http://"+addr+"/wasm_exec.html").Output()
	}()
	err := http.ListenAndServe(addr, http.FileServer(http.Dir(*dir)))
	if err != nil {
		log.Fatalln(err)
	}
}
