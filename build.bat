set pwd=%cd%

cd %1%

set GOOS=js
set GOARCH=wasm
set CGO_ENABLED=0
go build -ldflags "-s -w" -trimpath -o ..\test.wasm

set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -trimpath -o ..\test.exe

cd %pwd%
if not exist httpServer.exe (
  go build -ldflags "-s -w" -trimpath -o httpServer.exe
)

fc /b "%GOROOT%\misc\wasm\wasm_exec.html" wasm_exec.html > nul 2>&1
if %errorlevel% neq 0 (
  xcopy /y "%GOROOT%\misc\wasm\wasm_exec.html" wasm_exec.html*
)
fc /b "%GOROOT%\misc\wasm\wasm_exec.js" wasm_exec.js > nul 2>&1
if %errorlevel% neq 0 (
  xcopy /y "%GOROOT%\misc\wasm\wasm_exec.js" wasm_exec.js*
)
