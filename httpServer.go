package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Uint("addr", 8080, "listen address")
	dir := flag.String("dir", ".", "files directory to serve")
	flag.Parse()

	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	log.Printf("listening on %q...", addr)

	fh := http.FileServer(http.Dir(*dir))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/wasm_exec.html", "/wasm_exec.js", "/test.wasm":
			fh.ServeHTTP(w, r)
		default:
			http.Redirect(w, r, "/wasm_exec.html", http.StatusFound)
		}
	})
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
