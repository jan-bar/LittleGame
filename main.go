package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

func main() {
	fa := flag.String("addr", ":8080", "listen address")
	fb := flag.Bool("b", false, "build all game")
	flag.Parse()

	if *fb {
		err := build()
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		fh := http.FileServer(http.Dir("."))

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/wasm.html" || filepath.Ext(r.URL.Path) == ".wasm" {
				fh.ServeHTTP(w, r)
			} else {
				http.Redirect(w, r, "/wasm.html", http.StatusFound)
			}
		})

		log.Printf("listening on %q...", *fa)
		err := http.ListenAndServe(*fa, nil)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func build() error {
	de, err := os.ReadDir(".")
	if err != nil {
		return err
	}

	var (
		info []byte
		game []string
		ldSW = "-s -w"
	)
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == `windows` {
		ldSW += " -H windowsgui"
	}

	for _, d := range de {
		name := d.Name()
		if d.IsDir() && !strings.HasPrefix(name, ".") {
			cmd := exec.Command("go", "build", "-C", name, "-trimpath", "-ldflags", ldSW, "-o", "..")
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
			info, err = cmd.CombinedOutput()
			if err != nil {
				log.Printf("%s", info)
				return err
			}

			cmd = exec.Command("go", "build", "-C", name, "-trimpath", "-ldflags", "-s -w", "-o", fmt.Sprintf("../%s.wasm", name))
			cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOOS=js", "GOARCH=wasm")
			info, err = cmd.CombinedOutput()
			if err != nil {
				log.Printf("%s", info)
				return err
			}

			game = append(game, name)
		}
	}

	if len(game) == 0 {
		return nil
	}

	t, err := template.New("game").Parse(`<html>
<head>
    <meta charset="utf-8">
    <title>Games</title>
</head>
<body>
<script src="https://cdn.jsdelivr.net/gh/golang/go/misc/wasm/wasm_exec.js"></script>
<script>
    if (!WebAssembly.instantiateStreaming) {
        WebAssembly.instantiateStreaming = async (resp, importObject) => {
            const source = await (await resp).arrayBuffer();
            return await WebAssembly.instantiate(source, importObject);
        };
    }
    function run(wasm) {
        const go = new Go();
        WebAssembly.instantiateStreaming(fetch(wasm), go.importObject).then((res) => {
            go.run(res.instance);
            WebAssembly.instantiate(res.module, go.importObject);
        }).catch((err) => {
            console.error(err);
        });
    }
</script>
<ul>
{{range $i,$v := .games -}}
<li><button onClick="run('{{$v}}.wasm');">Run {{$v}}</button></li>
{{end -}}
</ul>
</body>
</html>`)
	if err != nil {
		return err
	}

	fw, err := os.Create("wasm.html")
	if err != nil {
		return err
	}
	defer fw.Close()

	return t.Execute(fw, map[string]any{
		"games": game,
	})
}
