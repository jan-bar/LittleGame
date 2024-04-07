# LittleGame

build: `.\build.bat [directory]`

generate `test.exe` to window platform

generate `test.wasm` web browser

generate `httpServer.exe` to http service

form `%GOROOT%\misc\wasm` copy `wasm_exec.html` and `wasm_exec.js` to http directory

example: 
```shell
# build game
.\build.bat minesweeper

# run game in windows
.\test.exe

# start http server
.\httpServer

# run the game in the browser, click run button in browser
start http://127.0.0.1:8080
```

* [GomokuGo](GomokuGo)
* [ChineseChess](ChineseChess)
* [minesweeper](minesweeper)

