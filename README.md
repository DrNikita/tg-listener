$env:CGO_ENABLED=1; 

$env:CGO_CFLAGS="-IC:/td/tdlib/include";

$env:CGO_LDFLAGS="-LC:/td/tdlib/bin -ltdjson";

go build -trimpath -ldflags="-s -w" -o demo.exe main.go