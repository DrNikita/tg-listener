# Stage 1: Сборка TDLib
FROM ubuntu:22.04 AS tdlib-builder

# Установка необходимых зависимостей
RUN apt-get update && apt-get install -y \
    cmake g++ make git wget zlib1g-dev libssl-dev && \
    apt-get clean

# Скачиваем и собираем TDLib
WORKDIR /build
RUN git clone https://github.com/tdlib/td.git && \
    cd td && \
    git clone https://github.com/Microsoft/vcpkg.git && \
    cd vcpkg && \
    git checkout 07b30b49e5136a36100a2ce644476e60d7f3ddc1 && \
    ./bootstrap-vcpkg.bat && \
    ./vcpkg.exe install gperf:x64-windows openssl:x64-windows zlib:x64-windows && \
    cd .. && \
    Remove-Item build -Force -Recurse -ErrorAction SilentlyContinue && \
    mkdir build && \
    cd build && \
    cmake -A x64 -DCMAKE_INSTALL_PREFIX:PATH=../tdlib -DCMAKE_TOOLCHAIN_FILE:FILEPATH=../vcpkg/scripts/buildsystems/vcpkg.cmake .. && \
    cmake --build . --target install --config Release && \
    cd .. && \
    cd .. && \
    dir td/tdlib

# Stage 2: Сборка Go-приложения
FROM golang:1.23.4 AS builder

# Копируем TDLib из предыдущего этапа
COPY --from=tdlib-builder /usr/local /usr/local

# Устанавливаем системные зависимости для TDLib
RUN apt-get update && apt-get install -y libssl-dev zlib1g-dev && apt-get clean

# Настройка рабочей директории
WORKDIR /app

# Копируем проект в контейнер
COPY . .

# Устанавливаем зависимости и собираем приложение
RUN go mod tidy
RUN go build -o tg-listener

# Stage 3: Финальный минимальный образ
FROM debian:bullseye-slim

# Устанавливаем системные зависимости TDLib
RUN apt-get update && apt-get install -y libssl-dev zlib1g-dev && apt-get clean

# Копируем TDLib из этапа сборки
COPY --from=tdlib-builder /usr/local /usr/local

# Копируем скомпилированное приложение
COPY --from=builder /app/tg-listener /tg-listener
COPY wait-for-it.sh /app/wait-for-it.sh
RUN chmod +x /app/wait-for-it.sh

# Устанавливаем переменную окружения для TDLib
ENV LD_LIBRARY_PATH=/usr/local/lib

# Указываем команду для запуска контейнера
CMD ["/app/wait-for-it.sh", "kafka:9092", "--", "/tg-listener"]