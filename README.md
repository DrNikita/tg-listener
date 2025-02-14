    win:

$env:CGO_ENABLED=1; 
$env:CGO_CFLAGS="-IC:/td/tdlib/include";
$env:CGO_LDFLAGS="-LC:/td/tdlib/bin -ltdjson";


    macos:

export CGO_ENABLED=1
export CGO_CFLAGS="-I/Users/nikita/td/tdlib/include -I/usr/local/opt/openssl/include"
export CGO_LDFLAGS="-L/Users/nikita/td/tdlib/lib -L/opt/homebrew/lib/ -lssl -lcrypto"


    linux:

export CGO_ENABLED=1
export CGO_CFLAGS="-I/home/nikita/td/tdlib/include"
export CGO_LDFLAGS="-L/home/nikita/td/tdlib/lib -ltdjson -ltdutils -ltdclient -lstdc++ -lm"


CGO_ENABLED=1
CGO_CFLAGS=-I/home/nikita/td/tdlib/include
CGO_LDFLAGS=-L/home/nikita/td/tdlib/lib -ltdjson


export CGO_ENABLED=1
export CGO_CFLAGS="-I$HOME/td/tdlib/include"
export CGO_LDFLAGS="-L$HOME/td/tdlib/lib -ltdjson"
export LD_LIBRARY_PATH="$HOME/td/tdlib/lib"



go build -trimpath -ldflags="-s -w" -o td-listener.exe main.go


----------------------------------------------------------------------------------------


go build -trimpath -ldflags="-s -w" -o tg-listener.exe main.go


----------------------------------------------------------------------------------------


JIRA: [text](https://helllolworld.atlassian.net/jira/software/projects/KAN/boards/1)