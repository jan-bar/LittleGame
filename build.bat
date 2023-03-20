set pwd=%cd%

cd %1%

set GOOS=js
set GOARCH=wasm
set CGO_ENABLED=0
go build -ldflags "-s -w" -trimpath -o ..\test.wasm

set GOOS=windows
set GOARCH=amd64
go build -ldflags "-H windowsgui -s -w" -trimpath -o ..\test.exe

cd %pwd%
if not exist httpServer.exe (
  go build -ldflags "-s -w" -trimpath -o httpServer.exe
)
