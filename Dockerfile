FROM alpine:latest

WORKDIR /

RUN apk add --update \
        ca-certificates

RUN apk add --update --virtual .build-deps \
        alpine-sdk linux-headers git zlib-dev openssl-dev gperf cmake

RUN git clone https://github.com/tdlib/td.git && \
    cd td && \
    git checkout 24893faf75d84b2b885f3f7aeb9d5a3c056fa7be && \
    mkdir build && \
    cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local .. && \
    cmake --build . --target install -j 2 && \
    cd / && \
    rm -rf /td/

# ENV CGO_ENABLED=1
# ENV CGO_CFLAGS="-I/usr/local/td/tdlib/include"
# ENV CGO_LDFLAGS="-L/usr/local/td/tdlib/bin -ltdjson"

# COPY go.mod go.sum ./
# RUN go mod download
# COPY . .

# RUN go build -trimpath -ldflags="-s -w" -o tg-listener.exe main.go

# CMD ["/usr/local/tg-listener", "-n", "--debug"]